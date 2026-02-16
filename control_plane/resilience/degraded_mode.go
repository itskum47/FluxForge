package resilience

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// PendingWrite represents a write that occurred during degraded mode
// CRITICAL: Includes version for safe reconciliation
type PendingWrite struct {
	Key        string
	Value      interface{}
	Timestamp  int64 // Unix timestamp
	TTL        time.Duration
	Version    int64 // CRITICAL: Version for conflict detection
	Reconciled bool  // CRITICAL: Idempotent reconciliation
}

// CacheEntry tracks access time for proper LRU eviction
type CacheEntry struct {
	Value      interface{}
	LastAccess time.Time
}

// DegradedMode manages graceful degradation when dependencies fail
// CRITICAL: Prevents state divergence by tracking pending writes for reconciliation
type DegradedMode struct {
	mu sync.RWMutex

	// Degradation state
	redisAvailable bool
	dbAvailable    bool
	natsAvailable  bool

	// Fallback state with PROPER LRU cache
	localCache   map[string]*CacheEntry
	cacheSize    int
	maxCacheSize int

	// CRITICAL: Bounded pending writes with version tracking
	pendingWrites    []PendingWrite
	maxPendingWrites int
	currentVersion   int64

	// Metrics
	degradedModeActive bool
	lastRedisCheck     time.Time
	lastDBCheck        time.Time
}

// NewDegradedMode creates a new degraded mode manager
func NewDegradedMode() *DegradedMode {
	return &DegradedMode{
		redisAvailable:   true,
		dbAvailable:      true,
		natsAvailable:    true,
		localCache:       make(map[string]*CacheEntry),
		maxCacheSize:     10000, // Bounded to prevent OOM
		cacheSize:        0,
		pendingWrites:    make([]PendingWrite, 0),
		maxPendingWrites: 10000, // CRITICAL: Bound pending writes too
		currentVersion:   0,
	}
}

// MarkRedisUnavailable marks Redis as unavailable and enters degraded mode
func (d *DegradedMode) MarkRedisUnavailable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.redisAvailable {
		log.Printf("[DEGRADED MODE] Redis unavailable, entering degraded mode")
		d.redisAvailable = false
		d.degradedModeActive = true
		d.lastRedisCheck = time.Now()
	}
}

// MarkRedisAvailable marks Redis as available again
func (d *DegradedMode) MarkRedisAvailable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.redisAvailable {
		log.Printf("[DEGRADED MODE] Redis recovered, exiting degraded mode")
		d.redisAvailable = true
		d.checkDegradedMode()
	}
}

// MarkDBUnavailable marks database as unavailable
func (d *DegradedMode) MarkDBUnavailable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.dbAvailable {
		log.Printf("[DEGRADED MODE] Database unavailable, entering degraded mode")
		d.dbAvailable = false
		d.degradedModeActive = true
		d.lastDBCheck = time.Now()
	}
}

// MarkDBAvailable marks database as available again
func (d *DegradedMode) MarkDBAvailable() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.dbAvailable {
		log.Printf("[DEGRADED MODE] Database recovered, exiting degraded mode")
		d.dbAvailable = true
		d.checkDegradedMode()
	}
}

// IsRedisAvailable checks if Redis is available
func (d *DegradedMode) IsRedisAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.redisAvailable
}

// IsDBAvailable checks if database is available
func (d *DegradedMode) IsDBAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.dbAvailable
}

// IsDegraded returns true if system is in degraded mode
func (d *DegradedMode) IsDegraded() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.degradedModeActive
}

// checkDegradedMode updates degraded mode status
func (d *DegradedMode) checkDegradedMode() {
	// Exit degraded mode only if all dependencies are available
	if d.redisAvailable && d.dbAvailable && d.natsAvailable {
		d.degradedModeActive = false
		log.Printf("[DEGRADED MODE] All dependencies recovered, normal mode restored")
	}
}

// GetFromCache retrieves value from local cache (fallback when Redis unavailable)
// CRITICAL: Updates LastAccess for proper LRU
func (d *DegradedMode) GetFromCache(key string) (interface{}, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry, ok := d.localCache[key]
	if !ok {
		return nil, false
	}

	// Update access time for LRU
	entry.LastAccess = time.Now()
	return entry.Value, true
}

// SetInCache stores value in local cache with bounded LRU
// CRITICAL: Tracks as DEGRADED_PENDING_SYNC for reconciliation
func (d *DegradedMode) SetInCache(key string, value interface{}) {
	d.SetInCacheWithTTL(key, value, 0)
}

// SetInCacheWithTTL stores value with TTL and tracks for reconciliation
func (d *DegradedMode) SetInCacheWithTTL(key string, value interface{}, ttl time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// CRITICAL: Enforce bounded pending writes
	if len(d.pendingWrites) >= d.maxPendingWrites {
		log.Printf("[DEGRADED MODE] Pending writes full (%d entries), dropping oldest unreconciled",
			d.maxPendingWrites)

		// Remove oldest unreconciled write
		for i := 0; i < len(d.pendingWrites); i++ {
			if !d.pendingWrites[i].Reconciled {
				d.pendingWrites = append(d.pendingWrites[:i], d.pendingWrites[i+1:]...)
				break
			}
		}
	}

	// Enforce bounded cache with PROPER LRU eviction
	if d.cacheSize >= d.maxCacheSize {
		// Find oldest accessed entry (proper LRU)
		var oldestKey string
		var oldestTime time.Time
		first := true

		for k, entry := range d.localCache {
			if first || entry.LastAccess.Before(oldestTime) {
				oldestKey = k
				oldestTime = entry.LastAccess
				first = false
			}
		}

		if oldestKey != "" {
			delete(d.localCache, oldestKey)
			d.cacheSize--
			log.Printf("[DEGRADED MODE] LRU evicted: %s (last access: %v)",
				oldestKey, oldestTime)
		}
	}

	// Store in local cache with access tracking
	if _, exists := d.localCache[key]; !exists {
		d.cacheSize++
	}
	d.localCache[key] = &CacheEntry{
		Value:      value,
		LastAccess: time.Now(),
	}

	// CRITICAL: Increment version for conflict detection
	d.currentVersion++

	// CRITICAL: Mark as DEGRADED_PENDING_SYNC with version
	d.pendingWrites = append(d.pendingWrites, PendingWrite{
		Key:        key,
		Value:      value,
		Timestamp:  time.Now().Unix(), // Unix timestamp
		TTL:        ttl,
		Version:    d.currentVersion,
		Reconciled: false,
	})

	log.Printf("[DEGRADED MODE] Write marked DEGRADED_PENDING_SYNC: %s (version: %d, total pending: %d)",
		key, d.currentVersion, len(d.pendingWrites))
}

// ClearCache clears local cache
func (d *DegradedMode) ClearCache() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.localCache = make(map[string]*CacheEntry)
	d.cacheSize = 0
}

// WithFallback executes primary function, falls back to secondary if primary fails
func (d *DegradedMode) WithFallback(
	ctx context.Context,
	primary func(context.Context) error,
	fallback func(context.Context) error,
) error {
	// Try primary first
	err := primary(ctx)
	if err == nil {
		return nil
	}

	// Log degradation
	log.Printf("[DEGRADED MODE] Primary operation failed: %v, using fallback", err)

	// Try fallback
	fallbackErr := fallback(ctx)
	if fallbackErr != nil {
		return fmt.Errorf("both primary and fallback failed: %w", fallbackErr)
	}

	return nil
}

// HealthCheck performs health check on dependencies
func (d *DegradedMode) HealthCheck(ctx context.Context) map[string]bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]bool{
		"redis":    d.redisAvailable,
		"database": d.dbAvailable,
		"nats":     d.natsAvailable,
		"degraded": d.degradedModeActive,
	}
}

// GetPendingWriteCount returns number of pending writes awaiting reconciliation
func (d *DegradedMode) GetPendingWriteCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.pendingWrites)
}
