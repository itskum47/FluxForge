package store

import (
	"context"
	"time"
)

// Store defines the methods required for a permanent storage backend.
// It abstracts over Postgres (durable) and Redis (ephemeral/fast).
type Store interface {
	// Agent Operations
	UpsertAgent(ctx context.Context, tenantID string, agent *Agent) error
	GetAgent(ctx context.Context, tenantID string, nodeID string) (*Agent, error)
	ListAgents(ctx context.Context, tenantID string) ([]*Agent, error)
	UpdateAgentHeartbeat(ctx context.Context, tenantID string, nodeID string, t time.Time) error

	// State Operations
	UpsertState(ctx context.Context, tenantID string, state *DesiredState) error
	UpdateStateStatus(ctx context.Context, tenantID string, stateID string, status string, lastError string, lastChecked time.Time, expectedVersion int) error
	GetState(ctx context.Context, tenantID string, stateID string) (*DesiredState, error)
	GetStateByNode(ctx context.Context, tenantID string, nodeID string) (*DesiredState, error)
	ListStates(ctx context.Context, tenantID string) ([]*DesiredState, error)
	ListStatesByStatus(ctx context.Context, status string, shardIndex int, shardCount int) ([]*DesiredState, error) // Global scan?
	CountStatesByStatus(ctx context.Context, tenantID string, status string) (int, error)

	// Job Operations
	CreateJob(ctx context.Context, tenantID string, job *Job) error
	UpdateJobStatus(ctx context.Context, tenantID string, jobID string, status string, exitCode int, stdout, stderr string) error
	GetJob(ctx context.Context, tenantID string, jobID string) (*Job, error)
	ListJobs(ctx context.Context, tenantID string, nodeID string, limit int) ([]*Job, error)
	ListJobsByTenant(ctx context.Context, tenantID string, limit int) ([]*Job, error)

	// Coordination Operations
	// IncrementDurableEpoch increments the epoch for a given resource (e.g. "leader_election")
	// and returns the new epoch. This must be atomic and durable.
	IncrementDurableEpoch(ctx context.Context, resourceID string) (int64, error)

	// GetDurableEpoch returns the current epoch without incrementing.
	GetDurableEpoch(ctx context.Context, resourceID string) (int64, error)

	// Idempotency Operations
	// GetIdempotencyRecord retrieves a cached idempotency response
	GetIdempotencyRecord(key string) (string, error)

	// SetIdempotencyRecord stores an idempotency response with TTL
	SetIdempotencyRecord(key string, value string, ttl time.Duration) error

	// SetIdempotencyRecordNX atomically sets only if key doesn't exist (prevents race)
	SetIdempotencyRecordNX(key string, value string, ttl time.Duration) error
}
