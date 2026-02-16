package store

import (
	"context"
	"time"
)

// Coordinator defines the interface for distributed coordination:
// Leader Election, Locks, and Presence.
type Coordinator interface {
	// Leader Election / Locking
	// AcquireLock attempts to acquire a lock (or leadership) for the given key.
	// Returns true if successful, false if lock is held by another.
	AcquireLock(ctx context.Context, key string, ownerID string, ttl time.Duration) (bool, error)

	// RenewLock extends the TTL of a held lock.
	RenewLock(ctx context.Context, key string, ownerID string, ttl time.Duration) (bool, error)

	// ReleaseLock releases the lock if held by ownerID.
	ReleaseLock(ctx context.Context, key string, ownerID string) error

	// GetLockOwner returns the current owner of the lock, or empty if free.
	GetLockOwner(ctx context.Context, key string) (string, error)

	// Lease Semantics (Leader Election & Heartbeats)
	// AcquireLease attempts to acquire a lease for a resource.
	// value should contain metadata (owner_id, req_id, timestamps).
	AcquireLease(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)

	// RenewLease extends the TTL of a held lease if the value matches.
	RenewLease(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)

	// ReleaseLease releases the lease if the value matches.
	ReleaseLease(ctx context.Context, key string, value string) error

	// IsLeaseOwner checks if the current value matches the given value.
	IsLeaseOwner(ctx context.Context, key string, value string) (bool, error)

	// IncrementEpoch increments the epoch counter for a resource and returns the new value.
	// This is used for generating Fencing Tokens.
	IncrementEpoch(ctx context.Context, key string) (int64, error)

	// ScanLocks returns a list of keys matching the pattern (e.g. "fluxforge:lock:*").
	// This is used by the Janitor.
	ScanLocks(ctx context.Context, pattern string) ([]string, error)
}
