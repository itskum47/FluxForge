package resilience

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/itskum47/FluxForge/control_plane/observability"
)

// LeaderEpoch represents a leadership term
type LeaderEpoch struct {
	Epoch     int64
	LeaderID  string
	StartTime time.Time
}

// ReconciliationCoordinator ensures only current leader reconciles
// CRITICAL: Validates leader epoch to prevent stale leader writes
type ReconciliationCoordinator struct {
	mu sync.RWMutex

	degradedMode *DegradedMode
	redisStore   VersionedRedisWriter
	nodeID       string // ADD: node ID for logging

	// Leadership tracking
	currentEpoch  int64
	leaderID      string
	isLeader      bool
	getLeaderInfo func() (*LeaderEpoch, error)
}

// NewReconciliationCoordinator creates coordinator with leader epoch validation
func NewReconciliationCoordinator(
	degradedMode *DegradedMode,
	redisStore VersionedRedisWriter,
	getLeaderInfo func() (*LeaderEpoch, error),
	nodeID string,
) *ReconciliationCoordinator {
	return &ReconciliationCoordinator{
		degradedMode:  degradedMode,
		redisStore:    redisStore,
		getLeaderInfo: getLeaderInfo,
		nodeID:        nodeID,
	}
}

// UpdateLeadershipStatus updates current leadership status
func (c *ReconciliationCoordinator) UpdateLeadershipStatus(epoch int64, leaderID string, isLeader bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentEpoch = epoch
	c.leaderID = leaderID
	c.isLeader = isLeader

	log.Printf("[RECONCILIATION] Leadership updated: epoch=%d, leader=%s, isLeader=%v",
		epoch, leaderID, isLeader)
}

// ReconcileIfLeader reconciles only if this node is current leader
// CRITICAL: Validates epoch throughout reconciliation to detect leadership changes
func (c *ReconciliationCoordinator) ReconcileIfLeader(ctx context.Context) error {
	// Check leadership at start
	c.mu.RLock()
	if !c.isLeader {
		c.mu.RUnlock()
		log.Printf("[RECONCILIATION] Skipping: not leader")
		return nil
	}
	startEpoch := c.currentEpoch
	c.mu.RUnlock()

	log.Printf("[RECONCILIATION] Starting reconciliation as leader (epoch: %d)", startEpoch)

	// Get pending writes
	pendingCount := c.degradedMode.GetPendingWriteCount()
	if pendingCount == 0 {
		log.Printf("[RECONCILIATION] No pending writes")
		return nil
	}

	// CRITICAL: Validate epoch before reconciliation
	leaderInfo, err := c.getLeaderInfo()
	if err != nil {
		return fmt.Errorf("failed to get leader info: %w", err)
	}

	if leaderInfo.Epoch != startEpoch {
		log.Printf("[RECONCILIATION] Epoch changed before reconcile: %d → %d, ABORTING",
			startEpoch, leaderInfo.Epoch)
		observability.ReconciliationEpochAbort.Inc() // CRITICAL METRIC
		return fmt.Errorf("leadership changed during reconciliation: epoch %d → %d",
			startEpoch, leaderInfo.Epoch)
	}

	// Perform reconciliation
	err = c.degradedMode.ReconcilePendingWrites(ctx, c.redisStore)

	// CRITICAL: Validate epoch after reconciliation (DUAL CHECK)
	c.mu.RLock()
	currentEpoch := c.currentEpoch
	c.mu.RUnlock()

	if currentEpoch != startEpoch {
		log.Printf("[RECONCILIATION] Epoch changed during reconcile: %d → %d, ABORTING commit",
			startEpoch, currentEpoch)
		observability.ReconciliationEpochAbort.Inc() // CRITICAL METRIC
		return fmt.Errorf("leadership changed during reconciliation: epoch %d → %d",
			startEpoch, currentEpoch)
	}

	if err != nil {
		return fmt.Errorf("reconciliation failed: %w", err)
	}

	log.Printf("[RECONCILIATION] Completed successfully (epoch: %d)", startEpoch)
	return nil
}

// StartPeriodicReconciliation starts background reconciliation (leader only)
// CRITICAL: Validates epoch on each iteration
func (c *ReconciliationCoordinator) StartPeriodicReconciliation(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[RECONCILIATION] Starting periodic reconciliation (interval: %v)", interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[RECONCILIATION] Stopping periodic reconciliation")
			return

		case <-ticker.C:
			c.mu.RLock()
			isLeader := c.isLeader
			epoch := c.currentEpoch
			c.mu.RUnlock()

			if !isLeader {
				continue
			}

			log.Printf("[RECONCILIATION] Periodic reconciliation starting (epoch: %d)", epoch)

			err := c.ReconcileIfLeader(ctx)
			if err != nil {
				log.Printf("[RECONCILIATION] Error: %v", err)
			}
		}
	}
}

// ReconcileWithDistributedLock alternative implementation using distributed lock
// CRITICAL: Prevents concurrent reconciliation across nodes
func (c *ReconciliationCoordinator) ReconcileWithDistributedLock(ctx context.Context, lockStore LockStore) error {
	const reconciliationLockKey = "reconciliation-global-lock"
	const reconciliationLockTTL = 5 * time.Minute

	// Acquire distributed lock
	acquired, err := lockStore.AcquireLock(ctx, reconciliationLockKey, reconciliationLockTTL)
	if err != nil {
		return fmt.Errorf("failed to acquire reconciliation lock: %w", err)
	}

	if !acquired {
		log.Printf("[RECONCILIATION] Another node is reconciling")
		return nil
	}

	defer lockStore.ReleaseLock(ctx, reconciliationLockKey)

	log.Printf("[RECONCILIATION] Acquired distributed lock, reconciling...")

	// Perform reconciliation
	return c.degradedMode.ReconcilePendingWrites(ctx, c.redisStore)
}

// LockStore interface for distributed locking
type LockStore interface {
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
}
