package coordination

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/itskum47/FluxForge/control_plane/observability"
	"github.com/itskum47/FluxForge/control_plane/store"
)

type LockMetadata struct {
	OwnerPod  string    `json:"owner_pod"`
	Epoch     int64     `json:"epoch"`
	ReqID     string    `json:"req_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LeaderElector struct {
	coordinator  store.Coordinator
	store        store.Store // Durable store for Epochs
	nodeID       string
	lockKey      string
	ttl          time.Duration
	leaderCtx    context.Context // Context valid only while leader
	leaderCancel context.CancelFunc

	mu           sync.RWMutex
	isLeader     bool
	currentValue string // The exact JSON string for the held lease
	currentEpoch int64  // The durable fencing token

	// Callbacks
	onElected func(context.Context)
	onLost    func()

	ctx    context.Context
	cancel context.CancelFunc

	// Phase 5.1: Leadership transition tracking
	stepDownTime time.Time // Time when leadership was lost (for transition duration metric)
	transitions  int64     // Total transitions (acquired + lost)
}

type LeaderState struct {
	IsLeader     bool   `json:"is_leader"`
	CurrentEpoch int64  `json:"current_epoch"`
	Transitions  int64  `json:"transitions"`
	NodeID       string `json:"node_id"`
}

type fencingKey string

const fencingEpochKey fencingKey = "fencing_epoch"

// FencedContext returns a context that is cancelled when leadership is lost.
// It also carries the current Fencing Epoch.
func (l *LeaderElector) FencedContext() context.Context {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.leaderCtx
}

// GetEpochFromContext extracts the fencing epoch from a context.
func GetEpochFromContext(ctx context.Context) (int64, bool) {
	val := ctx.Value(fencingEpochKey)
	if val == nil {
		return 0, false
	}
	epoch, ok := val.(int64)
	return epoch, ok
}

// GetState returns the internal state for the dashboard.
func (l *LeaderElector) GetState() LeaderState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return LeaderState{
		IsLeader:     l.isLeader,
		CurrentEpoch: l.currentEpoch,
		Transitions:  l.transitions,
		NodeID:       l.nodeID,
	}
}

func NewLeaderElector(c store.Coordinator, s store.Store, nodeID string, ttl time.Duration) *LeaderElector {
	ctx, cancel := context.WithCancel(context.Background())
	return &LeaderElector{
		coordinator: c,
		store:       s,
		nodeID:      nodeID,
		lockKey:     "fluxforge:lock:leader",
		ttl:         ttl,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (l *LeaderElector) SetCallbacks(onElected func(ctx context.Context), onLost func()) {
	l.onElected = onElected
	l.onLost = onLost
}

func (l *LeaderElector) Start(ctx context.Context) {
	// Use external context for lifecycle, but internal cancel for deep stop
	go l.loop(ctx)
}

func (l *LeaderElector) Stop() {
	l.cancel()
	if l.IsLeader() {
		l.release()
	}
}

func (l *LeaderElector) loop(ctx context.Context) {
	interval := l.ttl / 3
	minInterval := l.ttl / 3
	maxInterval := 10 * l.ttl

	renewFailures := 0
	const maxRenewFailures = 3

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			if l.IsLeader() {
				l.release()
			}
			return
		case <-timer.C:
			var err error
			if l.IsLeader() {
				var renewed bool
				renewed, err = l.renew(ctx)
				if err == nil {
					// Success
					renewFailures = 0
					if !renewed {
						l.stepDown()
					}
				} else {
					// Error encountered during renew
					renewFailures++
					log.Printf("LeaderElector: Renew failed (%d/%d): %v", renewFailures, maxRenewFailures, err)
					if renewFailures >= maxRenewFailures {
						log.Printf("LeaderElector: Too many renew failures. Stepping down for safety.")
						l.stepDown()
						renewFailures = 0
					}
				}
			} else {
				var acquired bool
				acquired, err = l.acquire(ctx)
				if err == nil && acquired {
					l.becomeLeader()
					renewFailures = 0
				}
			}

			// Backoff Logic
			if err != nil {
				// Exponential Backoff
				interval *= 2
				if interval > maxInterval {
					interval = maxInterval
				}
				log.Printf("LeaderElector: Error encountered, backing off for %v", interval)
			} else {
				// Reset
				interval = minInterval
			}

			timer.Reset(interval)
		}
	}
}

func (l *LeaderElector) IsLeader() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.isLeader
}

func (l *LeaderElector) acquire(ctx context.Context) (bool, error) {
	// 1. Get Epoch from Durable Store (Postgres)
	// This ensures monotonic fencing tokens even if Redis is flushed.
	epoch, err := l.store.IncrementDurableEpoch(ctx, "leader_election")
	if err != nil {
		log.Printf("LeaderElector: Failed to increment durable epoch: %v", err)
		return false, err
	}
	l.mu.Lock()
	// Epoch Drift Alert
	if l.currentEpoch > 0 && epoch > l.currentEpoch+1 {
		log.Printf("⚠️ ALERT: Epoch drift detected! Jamped from %d to %d (Difference: %d). High contention or partition recovery?", l.currentEpoch, epoch, epoch-l.currentEpoch)
		observability.LeadershipTransitions.WithLabelValues(l.nodeID, "epoch_drift").Inc()
	}
	l.currentEpoch = epoch
	l.mu.Unlock()

	// 2. Prepare Metadata
	meta := LockMetadata{
		OwnerPod:  l.nodeID,
		Epoch:     epoch,
		ReqID:     generateUUID(), // Simple UUID stub for now
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(l.ttl),
	}
	valBytes, _ := json.Marshal(meta)
	val := string(valBytes)

	// 3. Acquire Lease
	acquired, err := l.coordinator.AcquireLease(ctx, l.lockKey, val, l.ttl)
	if err != nil {
		log.Printf("LeaderElector: Failed to acquire lease: %v", err)
		return false, err
	}

	if acquired {
		l.mu.Lock()
		l.currentValue = val
		l.mu.Unlock()
	}
	return acquired, nil
}

func (l *LeaderElector) renew(ctx context.Context) (bool, error) {
	l.mu.RLock()
	val := l.currentValue
	l.mu.RUnlock()

	if val == "" {
		return false, nil
	}

	renewed, err := l.coordinator.RenewLease(ctx, l.lockKey, val, l.ttl)
	if err != nil {
		log.Printf("LeaderElector: Renew failed: %v", err)
		return false, err
	}
	return renewed, nil
}

func (l *LeaderElector) release() {
	l.mu.RLock()
	val := l.currentValue
	l.mu.RUnlock()

	if val == "" {
		return
	}

	ctxt, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Use ctxt (timeout context) for release, ignorinig outer context cancellation
	l.coordinator.ReleaseLease(ctxt, l.lockKey, val)
}

func (l *LeaderElector) becomeLeader() {
	l.mu.Lock()
	l.isLeader = true
	// Create a base cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	l.leaderCancel = cancel
	l.transitions++ // Increment transitions

	// Inject the epoch
	l.leaderCtx = context.WithValue(ctx, fencingEpochKey, l.currentEpoch)

	// Phase 5.1: Measure leadership transition duration
	if !l.stepDownTime.IsZero() {
		transitionDuration := time.Since(l.stepDownTime)
		observability.LeadershipTransitionDuration.Observe(transitionDuration.Seconds())
		log.Printf("✅ Node %s became LEADER (Epoch %d) - Transition took %v", l.nodeID, l.currentEpoch, transitionDuration)
		l.stepDownTime = time.Time{} // Reset
	} else {
		log.Printf("LeaderElector: Acquired leadership. Node: %s", l.nodeID)
	}
	l.mu.Unlock()

	// Metrics
	observability.LeadershipTransitions.WithLabelValues(l.nodeID, "acquired").Inc()
	observability.LeadershipEpoch.WithLabelValues(l.nodeID).Set(float64(l.currentEpoch))
	observability.LeaderStatus.Set(1)

	if l.onElected != nil {
		go l.onElected(l.leaderCtx)
	}
}

func (l *LeaderElector) stepDown() {
	l.mu.Lock()
	if !l.isLeader {
		l.mu.Unlock()
		return
	}

	observability.LeaderStatus.Set(0)
	l.isLeader = false

	l.transitions++ // Increment transitions

	// Phase 5.1: Record step-down time for transition duration tracking
	l.stepDownTime = time.Now()

	if l.leaderCancel != nil {
		l.leaderCancel()
	}
	l.mu.Unlock()

	// Metrics
	observability.LeadershipTransitions.WithLabelValues(l.nodeID, "lost").Inc()

	log.Printf("LeaderElector: Lost leadership. Node: %s", l.nodeID)
	if l.onLost != nil {
		l.onLost()
	}
}

// Stub for now
func generateUUID() string {
	return "uuid-" + time.Now().String() // Replace with real if needed
}
