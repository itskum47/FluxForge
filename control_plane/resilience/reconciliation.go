package resilience

import (
	"context"
	"log"
	"time"
)

// VersionedValue represents a value with version for reconciliation
type VersionedValue struct {
	Value     interface{}
	Version   int64
	Timestamp int64 // Unix timestamp
}

// ReconcilePendingWrites reconciles local cache writes to Redis after recovery
// CRITICAL: Uses versioning to prevent stale overwrites
func (d *DegradedMode) ReconcilePendingWrites(ctx context.Context, redisStore VersionedRedisWriter) error {
	d.mu.Lock()
	pending := make([]PendingWrite, len(d.pendingWrites))
	copy(pending, d.pendingWrites)
	d.mu.Unlock()

	if len(pending) == 0 {
		log.Printf("[DEGRADED MODE] No pending writes to reconcile")
		return nil
	}

	log.Printf("[DEGRADED MODE] Reconciling %d pending writes to Redis...", len(pending))

	successCount := 0
	failCount := 0
	skippedCount := 0

	for i, write := range pending {
		// Skip already reconciled (idempotent reconciliation)
		if write.Reconciled {
			skippedCount++
			continue
		}

		// Check if write is too old (stale)
		age := time.Since(time.Unix(write.Timestamp, 0))
		if age > 5*time.Minute {
			log.Printf("[DEGRADED MODE] Skipping stale write: %s (age: %v)", write.Key, age)

			// Mark as reconciled to prevent retry
			d.mu.Lock()
			if i < len(d.pendingWrites) {
				d.pendingWrites[i].Reconciled = true
			}
			d.mu.Unlock()

			failCount++
			continue
		}

		// CRITICAL: Check existing value version before overwriting
		existing, err := redisStore.GetVersioned(ctx, write.Key)
		if err != nil && err.Error() != "not found" {
			log.Printf("[DEGRADED MODE] Failed to get existing value for %s: %v", write.Key, err)
			failCount++
			continue
		}

		// CRITICAL: Only reconcile if our version is newer
		if existing != nil && existing.Version >= write.Version {
			log.Printf("[DEGRADED MODE] Skipping write %s: Redis has newer version (%d >= %d)",
				write.Key, existing.Version, write.Version)

			// Mark as reconciled (no need to retry)
			d.mu.Lock()
			if i < len(d.pendingWrites) {
				d.pendingWrites[i].Reconciled = true
			}
			d.mu.Unlock()

			skippedCount++
			continue
		}

		// Safe to reconcile: our version is newer
		versionedValue := VersionedValue{
			Value:     write.Value,
			Version:   write.Version,
			Timestamp: write.Timestamp,
		}

		err = redisStore.SetVersioned(ctx, write.Key, versionedValue, write.TTL)
		if err != nil {
			log.Printf("[DEGRADED MODE] Failed to reconcile write %s: %v", write.Key, err)
			failCount++
			continue
		}

		// Mark as reconciled atomically
		d.mu.Lock()
		if i < len(d.pendingWrites) {
			d.pendingWrites[i].Reconciled = true
		}
		d.mu.Unlock()

		successCount++
		log.Printf("[DEGRADED MODE] Reconciled %s (version: %d)", write.Key, write.Version)
	}

	// Clean up reconciled writes
	d.mu.Lock()
	unreconciled := make([]PendingWrite, 0)
	for _, write := range d.pendingWrites {
		if !write.Reconciled {
			unreconciled = append(unreconciled, write)
		}
	}
	d.pendingWrites = unreconciled
	d.mu.Unlock()

	log.Printf("[DEGRADED MODE] Reconciliation complete: %d succeeded, %d skipped (newer version), %d failed",
		successCount, skippedCount, failCount)

	if failCount > 0 {
		return &ReconciliationError{
			Total:   len(pending),
			Success: successCount,
			Skipped: skippedCount,
			Failed:  failCount,
		}
	}

	return nil
}

// VersionedRedisWriter interface for versioned reconciliation
type VersionedRedisWriter interface {
	GetVersioned(ctx context.Context, key string) (*VersionedValue, error)
	SetVersioned(ctx context.Context, key string, value VersionedValue, ttl time.Duration) error
}

// MarkRedisAvailableWithReconciliation marks Redis as available and triggers reconciliation
func (d *DegradedMode) MarkRedisAvailableWithReconciliation(ctx context.Context, redisStore VersionedRedisWriter) error {
	d.mu.Lock()
	wasUnavailable := !d.redisAvailable
	d.redisAvailable = true
	d.checkDegradedMode()
	d.mu.Unlock()

	if wasUnavailable {
		log.Printf("[DEGRADED MODE] Redis recovered, triggering versioned reconciliation...")
		return d.ReconcilePendingWrites(ctx, redisStore)
	}

	return nil
}
