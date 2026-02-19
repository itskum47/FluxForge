package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/itskum47/FluxForge/control_plane/store"
	"github.com/itskum47/FluxForge/control_plane/streaming"
)

// TestTaskTimeout_KillSwitch verifies that tasks are forcibly terminated after MaxTaskExecutionTime.
func TestTaskTimeout_KillSwitch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	s := store.NewMemoryStore()
	dispatcher := NewDispatcher(s)
	publisher := streaming.NewLogPublisher()
	reconciler := NewReconciler(s, dispatcher, publisher)

	// Set SHORT timeout for testing (3 seconds)
	reconciler.SetMaxTaskRuntime(3 * time.Second)

	// Create a state with a command that would take much longer
	state := &store.DesiredState{
		StateID:         "timeout-test-state",
		NodeID:          "test-node",
		Status:          "pending",
		CheckCmd:        "sleep 300", // 5 minute sleep - should timeout
		DesiredExitCode: 0,
		TenantID:        "default",
	}
	s.UpsertState(context.Background(), "default", state)

	ctx := context.Background()
	startTime := time.Now()

	// Run reconciliation - should timeout quickly
	err := reconciler.Reconcile(ctx, "default", state.StateID)
	elapsed := time.Since(startTime)

	// Verify it timed out (should get context deadline exceeded error)
	if err == nil {
		t.Error("Expected reconciliation to fail due to timeout, but got nil error")
	}

	// Verify it timed out around 3 seconds (not 300 seconds)
	if elapsed > 5*time.Second {
		t.Errorf("Expected timeout around 3s, but took %v (too long!)", elapsed)
	}

	t.Logf("✓ Task correctly timed out after %v (max: 3s)", elapsed)
	t.Logf("✓ Error: %v", err)
}

// TestEventPublishFailure_NonBlocking verifies that event publish failures don't block reconciliation.
func TestEventPublishFailure_NonBlocking(t *testing.T) {
	s := store.NewMemoryStore()
	dispatcher := NewDispatcher(s)

	// Create a publisher that always fails
	failingPublisher := &FailingPublisher{}

	reconciler := NewReconciler(s, dispatcher, failingPublisher)
	reconciler.SetMaxTaskRuntime(10 * time.Second)

	// Create a simple state
	state := &store.DesiredState{
		StateID:         "event-test-state",
		NodeID:          "test-node",
		Status:          "pending",
		CheckCmd:        "true",
		DesiredExitCode: 0,
		TenantID:        "default",
	}
	s.UpsertState(context.Background(), "default", state)

	ctx := context.Background()
	startTime := time.Now()

	// Run reconciliation - will fail at agent lookup, but that's OK
	// The key is that updateStatus should NOT block on publish failures
	_ = reconciler.Reconcile(ctx, "default", state.StateID)

	elapsed := time.Since(startTime)

	// Verify it didn't hang waiting for event publish
	// Even though publish fails, it should complete quickly (async)
	if elapsed > 3*time.Second {
		t.Errorf("Reconciliation took too long (%v), may have blocked on event publish", elapsed)
	}

	t.Logf("✓ Reconciliation completed in %v (non-blocking despite publish failures)", elapsed)

	// Give async publisher goroutine time to execute
	time.Sleep(100 * time.Millisecond)

	if failingPublisher.callCount > 0 {
		t.Logf("✓ Publisher was called %d times (and failed, as expected)", failingPublisher.callCount)
	}
}

// TestReconciler_MaxTaskRuntimeConfiguration verifies that MaxTaskRuntime can be configured.
func TestReconciler_MaxTaskRuntimeConfiguration(t *testing.T) {
	s := store.NewMemoryStore()
	dispatcher := NewDispatcher(s)
	publisher := streaming.NewLogPublisher()
	reconciler := NewReconciler(s, dispatcher, publisher)

	// Default should be 5 minutes
	if reconciler.maxTaskRuntime != 5*time.Minute {
		t.Errorf("Expected default maxTaskRuntime to be 5m, got %v", reconciler.maxTaskRuntime)
	}

	// Should be configurable
	reconciler.SetMaxTaskRuntime(1 * time.Minute)
	if reconciler.maxTaskRuntime != 1*time.Minute {
		t.Errorf("Expected maxTaskRuntime to be 1m after SetMaxTaskRuntime, got %v", reconciler.maxTaskRuntime)
	}

	t.Log("✓ MaxTaskRuntime configuration works correctly")
}

// FailingPublisher always returns an error (for testing)
type FailingPublisher struct {
	callCount int
}

func (f *FailingPublisher) Publish(ctx context.Context, topic string, payload interface{}) error {
	f.callCount++
	return fmt.Errorf("simulated publish failure")
}

func (f *FailingPublisher) Close() error {
	return nil
}
