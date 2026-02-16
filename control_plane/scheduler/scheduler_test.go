package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/itskum47/FluxForge/control_plane/store"
	"github.com/itskum47/FluxForge/control_plane/timeline"
)

type MockStore struct{}

func (m *MockStore) ListStatesByStatus(ctx context.Context, status string, shardIndex int, shardCount int) ([]*store.DesiredState, error) {
	return nil, nil
}

type MockReconciler struct {
	processed  []string
	shouldFail bool
}

func (m *MockReconciler) Reconcile(ctx context.Context, stateID string) error {
	m.processed = append(m.processed, stateID)
	if m.shouldFail {
		return errors.New("simulated failure")
	}
	return nil
}

func TestSchedulerPriority(t *testing.T) {
	mockRec := &MockReconciler{}
	mockStore := &MockStore{}
	sched := NewScheduler(mockStore, mockRec, 0, 1, DefaultSchedulerConfig())
	sched.RehydrateQueue(context.Background()) // Activate scheduler

	// Submit Low Priority Task OLD (should have aged to be higher priority than recent Medium)
	sched.Submit(&ReconciliationTask{
		ReqID:      "low-old",
		NodeID:     "node-1",
		Priority:   10,
		StateID:    "state-low",
		SubmitTime: time.Now().Add(-2 * time.Minute), // Aged 2 minutes
	})

	// Submit High Priority Task RECENT
	sched.Submit(&ReconciliationTask{
		ReqID:      "high-recent",
		NodeID:     "node-1",
		Priority:   0,
		StateID:    "state-high",
		SubmitTime: time.Now(),
	})

	// Start Scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched.Start(ctx)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)
}

func TestQueueOrdering(t *testing.T) {
	q := NewThreadSafeQueue()
	now := time.Now()

	// P10 but old (Effective P ~ -2)
	q.Push(&ReconciliationTask{Priority: 10, StateID: "old-low", SubmitTime: now.Add(-2 * time.Minute)})

	// P0 recent (Effective P 0)
	q.Push(&ReconciliationTask{Priority: 0, StateID: "recent-high", SubmitTime: now})

	// P5 recent (Effective P 5)
	q.Push(&ReconciliationTask{Priority: 5, StateID: "recent-medium", SubmitTime: now})

	// Expected: old-low, recent-high, recent-medium
	first := q.Pop()
	if first.StateID != "old-low" {
		t.Errorf("Expected old-low first due to aging, got %s", first.StateID)
	}

	second := q.Pop()
	if second.StateID != "recent-high" {
		t.Errorf("Expected recent-high second, got %s", second.StateID)
	}
}

func TestNodeHealth(t *testing.T) {
	mockRec := &MockReconciler{}
	mockStore := &MockStore{}
	sched := NewScheduler(mockStore, mockRec, 0, 1, DefaultSchedulerConfig())
	sched.RehydrateQueue(context.Background()) // Activate scheduler

	// Set node as quarantined (score 0.0 < 0.4 threshold)
	sched.UpdateNodeHealth("node-bad", "external", 0.0, "")

	// Submit task for bad node
	sched.Submit(&ReconciliationTask{
		ReqID:    "task-bad",
		NodeID:   "node-bad",
		Priority: 5,
		StateID:  "state-bad",
	})

	// Start Scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched.Start(ctx)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify task was dropped (not processed)
	if len(mockRec.processed) > 0 {
		t.Errorf("Expected 0 processed tasks (quarantined), got %d", len(mockRec.processed))
	}
}

func TestGetSnapshot(t *testing.T) {
	mockRec := &MockReconciler{}
	mockStore := &MockStore{}
	sched := NewScheduler(mockStore, mockRec, 0, 1, DefaultSchedulerConfig())
	sched.RehydrateQueue(context.Background()) // Activate scheduler

	sched.Submit(&ReconciliationTask{
		ReqID:    "snap-1",
		NodeID:   "node-1",
		Priority: 5,
		StateID:  "state-1",
	})

	snap := sched.GetSnapshot()

	if val, ok := snap["queue_depth"]; !ok || val.(int) != 1 {
		t.Errorf("Expected queue_depth 1, got %v", val)
	}

	events := snap["timeline_events"].([]timeline.ReconcileEvent)
	if len(events) != 1 {
		t.Errorf("Expected 1 timeline event (QUEUED), got %d", len(events))
	}
	if events[0].Stage != "QUEUED" {
		t.Errorf("Expected QUEUED stage, got %s", events[0].Stage)
	}
}

func TestSchedulerModes(t *testing.T) {
	mockRec := &MockReconciler{}
	mockStore := &MockStore{}
	sched := NewScheduler(mockStore, mockRec, 0, 1, DefaultSchedulerConfig())
	sched.RehydrateQueue(context.Background()) // Activate scheduler

	// Normal Mode
	err := sched.Submit(&ReconciliationTask{Priority: 10, StateID: "ok"})
	if err != nil {
		t.Errorf("Normal mode rejected task: %v", err)
	}

	// Degraded Mode
	sched.SetMode(ModeDegraded)
	err = sched.Submit(&ReconciliationTask{Priority: 10, StateID: "low-prio"})
	if err == nil {
		t.Error("Degraded mode accepted low priority task")
	}
	err = sched.Submit(&ReconciliationTask{Priority: 0, StateID: "high-prio"})
	if err != nil {
		t.Errorf("Degraded mode rejected high priority task: %v", err)
	}

	// ReadOnly Mode
	sched.SetMode(ModeReadOnly)
	err = sched.Submit(&ReconciliationTask{Priority: 0, StateID: "high-prio"})
	if err == nil {
		t.Error("ReadOnly mode accepted task")
	}
}
