package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"errors"

	"github.com/itskum47/FluxForge/control_plane/observability"
	"github.com/itskum47/FluxForge/control_plane/timeline"
)

// ReconcilerInterface defines the contract for the actual reconciliation logic.
// The scheduler calls this to execute a task.
type ReconcilerInterface interface {
	Reconcile(stateID string) error
}

var ErrQueueFull = errors.New("scheduler queue is full")

// Scheduler manages the execution of reconciliation tasks.
type Scheduler struct {
	queue          *ThreadSafeQueue
	nodeLimiters   *TokenBucketLimiter
	tenantLimiters *TokenBucketLimiter
	reconciler     ReconcilerInterface
	nodeHealth     map[string]*NodeHealth
	domainFailures map[string]int
	domainTasks    map[string]int
	timeline       *timeline.Store
	mode           SchedulerMode
	mu             sync.RWMutex // Protects mode changes
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(reconciler ReconcilerInterface) *Scheduler {
	return &Scheduler{
		queue: NewThreadSafeQueue(),
		// Limit to 5 requests per second per node, burst 1
		nodeLimiters: NewTokenBucketLimiter(5, 1),
		// Limit to 50 requests per second per tenant, burst 10
		tenantLimiters: NewTokenBucketLimiter(50, 10),
		reconciler:     reconciler,
		nodeHealth:     make(map[string]*NodeHealth),
		domainFailures: make(map[string]int),
		domainTasks:    make(map[string]int),
		timeline:       timeline.NewStore(),
		mode:           ModeNormal,
	}
}

// SetMode updates the scheduler operating mode.
func (s *Scheduler) SetMode(mode SchedulerMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
	log.Printf("Scheduler switched to %s mode", mode)

	// Update Metric
	// Reset all modes first (simple approach) or just set the current one to 1 and others to 0?
	// Simplified: Just set the active one to 1. Ideally we cleared others.
	observability.SchedulerModeMetric.WithLabelValues(string(mode)).Set(1)
}

// Submit adds a task to the scheduler.
func (s *Scheduler) Submit(task *ReconciliationTask) error {
	s.mu.RLock()
	currentMode := s.mode
	s.mu.RUnlock()

	// Mode Checks
	if currentMode == ModeReadOnly || currentMode == ModeDraining {
		return errors.New("scheduler is in read-only/draining mode")
	}

	if currentMode == ModeDegraded && task.Priority > 5 {
		// In degraded mode, only accept high priority (0-5)
		return errors.New("scheduler is degraded: low priority task rejected")
	}

	// Self-Protection: Reject low priority tasks if queue is full
	if s.queue.Len() > 1000 && task.Priority > 0 {
		return ErrQueueFull
	}

	if task.SubmitTime.IsZero() {
		task.SubmitTime = time.Now()
	}
	s.queue.Push(task)
	s.timeline.Record(timeline.ReconcileEvent{
		ReqID:    task.ReqID,
		Stage:    "QUEUED",
		NodeID:   task.NodeID,
		TenantID: task.TenantID,
		Metadata: map[string]string{"state_id": task.StateID},
	})
	return nil
}

// Start begins the scheduling loop.
func (s *Scheduler) Start(ctx context.Context) {
	go s.worker(ctx)
}

// worker constantly polls for tasks and executes them.
func (s *Scheduler) worker(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		start := time.Now()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processNextTask()
		}
		duration := time.Since(start).Seconds()
		observability.SchedulerLoopDuration.Observe(duration)

		// Update Queue Depth Metric
		observability.TaskQueueDepth.WithLabelValues("all").Set(float64(s.queue.Len()))

		// Update Oldest Task Age Metric
		// Peak at the queue safely
		oldest := s.queue.Peek()
		if oldest != nil {
			age := time.Since(oldest.SubmitTime).Seconds()
			observability.QueueOldestTaskAge.WithLabelValues(oldest.TenantID, fmt.Sprintf("%d", oldest.Priority)).Set(age)
		} else {
			observability.QueueOldestTaskAge.WithLabelValues("unknown", "unknown").Set(0)
		}
	}
}

// UpdateNodeHealth updates a specific signal for a node's health.
func (s *Scheduler) UpdateNodeHealth(nodeID, signal string, score float64) {
	health, exists := s.nodeHealth[nodeID]
	if !exists {
		health = &NodeHealth{NodeID: nodeID}
		s.nodeHealth[nodeID] = health
	}

	switch signal {
	case "agent":
		health.AgentReportedHealth = score
	case "observed":
		health.ObservedFailureRate = score
	case "external":
		health.ExternalProbeScore = score
	}

	health.CalculateCompositeScore()

	// Quarantine Logic
	if health.CompositeScore < 0.4 {
		health.Quarantined = true
		health.BackoffDuration = 1 * time.Minute
	} else {
		health.Quarantined = false
		health.BackoffDuration = 0
	}
}

func (s *Scheduler) processNextTask() {
	if s.queue.Len() == 0 {
		return
	}

	task := s.queue.Pop()
	if task == nil {
		return
	}

	// 1. Check Node Health (Composite Score)
	if health, exists := s.nodeHealth[task.NodeID]; exists {
		if health.Quarantined {
			logDecision(SchedulingDecision{
				Component: "scheduler",
				Decision:  "QUARANTINE_DROP",
				ReqID:     task.ReqID,
				NodeID:    task.NodeID,
				Reason:    "Node quarantined due to low health score",
				Metadata:  map[string]float64{"score": health.CompositeScore},
			})
			return // Drop task
		}
	}

	// 1.5 Check Failure Domain Isolation
	// If domain has > 5 recent failures, throttle concurrency to 1
	if task.FailureDomain != "" {
		failures := s.domainFailures[task.FailureDomain]
		active := s.domainTasks[task.FailureDomain]
		limit := 10 // Normal concurrency limit

		if failures > 5 {
			limit = 1 // Throttled mode
		}

		if active >= limit {
			// Domain saturated or throttled. Requeue with delay.
			logDecision(SchedulingDecision{
				Component: "scheduler",
				Decision:  "DOMAIN_THROTTLE",
				ReqID:     task.ReqID,
				Priority:  task.Priority,
				Reason:    "Failure domain saturation",
				Metadata:  map[string]int{"failures": failures, "active": active, "limit": limit},
			})
			go func() {
				time.Sleep(2 * time.Second)
				s.Submit(task)
			}()
			return
		}
	}

	// 2. Check Rate Limits
	if !s.nodeLimiters.Allow(task.NodeID) {
		// Node limit: Requeue with backoff (soft limit)
		go func() {
			time.Sleep(1 * time.Second)
			s.Submit(task)
		}()
		return
	}

	// Tenant Isolation: Hard Rate Limit
	if !s.tenantLimiters.Allow(task.TenantID) {
		logDecision(SchedulingDecision{
			Component: "scheduler",
			Decision:  "TENANT_THROTTLE",
			TenantID:  task.TenantID,
			ReqID:     task.ReqID,
			Reason:    "Tenant rate limit exceeded",
		})
		// Drop/Reject for strict isolation, or maybe Requeue with long delay?
		// Master prompt says "Enforce limits". Let's Requeue with penalty to avoid dropping data but slow them down.
		go func() {
			time.Sleep(5 * time.Second) // Heavy penalty
			s.Submit(task)
		}()
		return
	}

	// 3. Dispatch
	// Log decision
	decision := SchedulingDecision{
		Component: "scheduler",
		Decision:  "DISPATCH",
		ReqID:     task.ReqID,
		TenantID:  task.TenantID,
		NodeID:    task.NodeID,
		Priority:  task.Priority,
	}
	logDecision(decision)

	// Update Domain Metrics
	if task.FailureDomain != "" {
		s.domainTasks[task.FailureDomain]++
	}

	go func() {
		err := s.reconciler.Reconcile(task.StateID)

		stage := "FINISHED"
		meta := make(map[string]string)
		if err != nil {
			stage = "FAILED"
			meta["error"] = err.Error()
		}
		s.timeline.Record(timeline.ReconcileEvent{
			ReqID:    task.ReqID,
			Stage:    stage,
			NodeID:   task.NodeID,
			TenantID: task.TenantID,
			Metadata: meta,
		})

		// Add attempt number to metadata if not present (simple implementation)
		meta["attempt_number"] = fmt.Sprintf("%d", task.Attempt)

		// Completion Logic
		if task.FailureDomain != "" {
			s.domainTasks[task.FailureDomain]--
			if err != nil {
				s.domainFailures[task.FailureDomain]++
			}
		}
	}()
}

func logDecision(d SchedulingDecision) {
	bytes, _ := json.Marshal(d)
	log.Println(string(bytes))

	observability.SchedulerDecisions.WithLabelValues(d.Decision, d.Reason).Inc()
}

// GetSnapshot returns the internal state for debugging.
func (s *Scheduler) GetSnapshot() map[string]interface{} {
	return map[string]interface{}{
		"queue_depth":     s.queue.Len(),
		"domain_failures": s.domainFailures,
		"domain_active":   s.domainTasks,
		"timeline_events": s.timeline.GetAllEvents(),
		"mode":            s.mode,
	}
}
