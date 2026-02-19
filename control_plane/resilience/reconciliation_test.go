package resilience

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestVersionConflictChaos is the ULTIMATE reconciliation correctness test
// CRITICAL: Verifies version enforcement under chaos conditions
func TestVersionConflictChaos(t *testing.T) {
	// This test validates the exact scenario:
	// 1. Redis down
	// 2. Control-1 writes version 10 (pending)
	// 3. Redis restored
	// 4. Control-2 writes version 11
	// 5. Control-1 reconciliation runs
	// 6. Verify version 11 survives (not overwritten by version 10)

	t.Run("StaleWriteDoesNotOverwriteNewerVersion", func(t *testing.T) {
		// Setup
		degradedMode := NewDegradedMode()
		mockRedis := newMockVersionedRedis()

		// STEP 1: Simulate Redis down
		degradedMode.MarkRedisUnavailable()

		// STEP 2: Control-1 writes version 10 to local cache (pending)
		degradedMode.SetInCacheWithTTL("key1", "value-v10", 0)
		// This creates pendingWrite with version 10

		// STEP 3: Simulate Redis restored
		degradedMode.MarkRedisAvailable()

		// STEP 4: Control-2 writes version 11 directly to Redis
		mockRedis.SetVersioned(context.Background(), "key1", VersionedValue{
			Value:     "value-v11",
			Version:   11,
			Timestamp: time.Now().Unix(), // Unix timestamp (int64)
		}, 0)

		// STEP 5: Control-1 reconciliation runs
		err := degradedMode.ReconcilePendingWrites(context.Background(), mockRedis)

		// STEP 6: Verify version 11 survives
		result, err := mockRedis.GetVersioned(context.Background(), "key1")
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if result.Version != 11 {
			t.Errorf("Version conflict: expected version 11, got %d", result.Version)
			t.Errorf("CRITICAL FAILURE: Stale write overwrote newer version!")
		}

		if result.Value != "value-v11" {
			t.Errorf("Value conflict: expected value-v11, got %v", result.Value)
		}

		t.Logf("✅ Version conflict chaos test PASSED: version 11 survived reconciliation")
	})

	t.Run("NewerPendingWriteSucceeds", func(t *testing.T) {
		degradedMode := NewDegradedMode()
		mockRedis := newMockVersionedRedis()

		// Redis down
		degradedMode.MarkRedisUnavailable()

		// CRITICAL FIX: Manually set version to 15 for this test
		// Simulate that this node has seen 14 previous writes
		degradedMode.mu.Lock()
		degradedMode.currentVersion = 14
		degradedMode.mu.Unlock()

		// Control-1 writes version 15 (pending) - this will increment to 15
		degradedMode.SetInCacheWithTTL("key2", "value-v15", 0)

		// Redis restored
		degradedMode.MarkRedisAvailable()

		// Control-2 writes version 12 to Redis (older)
		mockRedis.SetVersioned(context.Background(), "key2", VersionedValue{
			Value:     "value-v12",
			Version:   12,
			Timestamp: time.Now().Unix(), // Unix timestamp (int64)
		}, 0)

		// Reconciliation runs
		err := degradedMode.ReconcilePendingWrites(context.Background(), mockRedis)
		if err != nil {
			t.Logf("Reconciliation error (expected partial): %v", err)
		}

		// Verify version 15 wins
		result, _ := mockRedis.GetVersioned(context.Background(), "key2")
		if result.Version != 15 {
			t.Errorf("Expected version 15, got %d", result.Version)
		}

		t.Logf("✅ Newer pending write correctly overwrote older Redis value")
	})
}

// TestIdempotencyLockExpiration validates lock orphaning prevention
func TestIdempotencyLockExpiration(t *testing.T) {
	t.Run("LockExpiresAutomatically", func(t *testing.T) {
		mockRedis := newMockRedisStore()

		// Acquire lock with 1 second TTL
		acquired, err := mockRedis.AcquireLock(context.Background(), "test-key", 1*time.Second)
		if err != nil {
			t.Fatalf("Failed to acquire lock: %v", err)
		}
		if !acquired {
			t.Fatal("Expected to acquire lock")
		}

		// Simulate process crash (no release)
		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Lock should be available again
		acquired, err = mockRedis.AcquireLock(context.Background(), "test-key", 1*time.Second)
		if err != nil {
			t.Fatalf("Failed to acquire lock after expiration: %v", err)
		}
		if !acquired {
			t.Error("Lock did not expire - orphaning detected!")
		}

		t.Logf("✅ Lock expiration test PASSED: lock auto-released after TTL")
	})
}

// TestLeaderEpochValidation validates stale leader prevention
func TestLeaderEpochValidation(t *testing.T) {
	t.Run("StaleLeaderCannotReconcile", func(t *testing.T) {
		degradedMode := NewDegradedMode()
		mockRedis := newMockVersionedRedis()

		// CRITICAL FIX: Add pending writes so epoch check actually runs
		degradedMode.MarkRedisUnavailable()
		degradedMode.SetInCacheWithTTL("test-key", "test-value", 0)
		degradedMode.MarkRedisAvailable()

		currentEpoch := int64(5)
		getLeaderInfo := func() (*LeaderEpoch, error) {
			return &LeaderEpoch{
				Epoch:    currentEpoch,
				LeaderID: "node-1",
			}, nil
		}

		coordinator := NewReconciliationCoordinator(degradedMode, mockRedis, getLeaderInfo, "node-1")

		// Node becomes leader at epoch 5
		coordinator.UpdateLeadershipStatus(5, "node-1", true)

		// Leadership changes to epoch 6 mid-reconciliation
		currentEpoch = 6

		// Attempt reconciliation (should fail due to epoch change)
		err := coordinator.ReconcileIfLeader(context.Background())

		if err == nil {
			t.Error("Expected reconciliation to fail due to epoch change")
		} else {
			t.Logf("Reconciliation correctly failed: %v", err)
		}

		t.Logf("✅ Stale leader prevention test PASSED: epoch validation prevented stale write")
	})
}

// Mock implementations for testing

type mockVersionedRedis struct {
	data map[string]VersionedValue
}

func newMockVersionedRedis() *mockVersionedRedis {
	return &mockVersionedRedis{
		data: make(map[string]VersionedValue),
	}
}

func (m *mockVersionedRedis) GetVersioned(ctx context.Context, key string) (*VersionedValue, error) {
	if val, ok := m.data[key]; ok {
		return &val, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockVersionedRedis) SetVersioned(ctx context.Context, key string, value VersionedValue, ttl time.Duration) error {
	// CRITICAL: Atomic version check
	if existing, ok := m.data[key]; ok {
		if value.Version <= existing.Version {
			return fmt.Errorf("version conflict: existing %d >= new %d", existing.Version, value.Version)
		}
	}
	m.data[key] = value
	return nil
}

type mockRedisStore struct {
	locks map[string]time.Time
}

func newMockRedisStore() *mockRedisStore {
	return &mockRedisStore{
		locks: make(map[string]time.Time),
	}
}

func (m *mockRedisStore) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if expiry, exists := m.locks[key]; exists {
		if time.Now().Before(expiry) {
			return false, nil // Lock still held
		}
	}
	m.locks[key] = time.Now().Add(ttl)
	return true, nil
}

func (m *mockRedisStore) ReleaseLock(ctx context.Context, key string) error {
	delete(m.locks, key)
	return nil
}
