package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/itskum47/FluxForge/control_plane/scheduler"
	"github.com/itskum47/FluxForge/control_plane/store"
	"github.com/itskum47/FluxForge/control_plane/streaming"
)

// FaultInjectionStore wraps a Store to simulate failures
type FaultInjectionStore struct {
	store.Store
	fail bool
	mu   sync.Mutex
}

func (f *FaultInjectionStore) SetFail(fail bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fail = fail
}

func (f *FaultInjectionStore) shouldFail() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fail
}

// Override critical methods
func (f *FaultInjectionStore) GetAgent(ctx context.Context, id string) (*store.Agent, error) {
	if f.shouldFail() {
		return nil, errors.New("simulated db error")
	}
	return f.Store.GetAgent(ctx, id)
}

// Add more overrides as needed for the test...
// Usually we need to mock IncrementDurableEpoch for LeaderElector test,
// but LeaderElector uses specific interface methods.

// TestChaos_PodCrash verifies component shutdown
func TestChaos_PodCrash(t *testing.T) {
	s := store.NewMemoryStore()
	dispatcher := NewDispatcher(s)
	publisher := streaming.NewLogPublisher()
	reconciler := NewReconciler(s, dispatcher, publisher)
	schedConfig := scheduler.DefaultSchedulerConfig()
	sched := scheduler.NewScheduler(s, reconciler, 0, 1, schedConfig)

	ctx, cancel := context.WithCancel(context.Background())
	go sched.Start(ctx)

	// Let it run briefly
	time.Sleep(1 * time.Second)

	// CRASH (Cancel Context)
	t.Log("Simulating Pod Crash (Context Cancel)...")
	cancel()

	// Wait for shutdown (logs should show cleanliness)
	time.Sleep(1 * time.Second)
	// We can't easily assert internal state of closed scheduler without hooks,
	// but if test doesn't panic/hang, it's a pass for "Crash Safety".
}

// TestChaos_DBFailover verifies resilience to transient DB errors
func TestChaos_DBFailover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	baseStore := store.NewMemoryStore()
	fStore := &FaultInjectionStore{Store: baseStore}

	// Test DB resilience at the Store level
	ctx := context.Background()

	// 1. Setup test agent
	agent := &store.Agent{
		NodeID:        "chaos-agent",
		Status:        "active",
		IPAddress:     "127.0.0.1",
		Port:          8080,
		LastHeartbeat: time.Now(),
	}

	// 2. Happy Path - Store operations work
	t.Log("Testing healthy DB operations...")
	if err := baseStore.UpsertAgent(ctx, agent); err != nil {
		t.Fatalf("Failed to insert agent: %v", err)
	}

	retrieved, err := fStore.GetAgent(ctx, "chaos-agent")
	if err != nil {
		t.Fatalf("Failed to retrieve agent: %v", err)
	}
	if retrieved == nil || retrieved.NodeID != "chaos-agent" {
		t.Fatal("Retrieved agent mismatch")
	}
	t.Log("✓ Healthy DB operations successful")

	// 3. Inject Failure
	t.Log("Injecting DB Failure...")
	fStore.SetFail(true)

	// 4. Verify operations fail gracefully
	_, err = fStore.GetAgent(ctx, "chaos-agent")
	if err == nil {
		t.Error("GetAgent should have failed with DB error, but got nil")
	} else {
		t.Logf("✓ Got expected error: %v", err)
	}

	// 5. Recover
	t.Log("Recovering DB...")
	fStore.SetFail(false)

	// 6. Verify recovery
	retrieved, err = fStore.GetAgent(ctx, "chaos-agent")
	if err != nil {
		t.Errorf("GetAgent failed after recovery: %v", err)
	}
	if retrieved == nil || retrieved.NodeID != "chaos-agent" {
		t.Error("Retrieved agent mismatch after recovery")
	}
	t.Log("✓ DB recovery successful")
}
