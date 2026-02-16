package store

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MemoryStore holds the in-memory state of registered agents and jobs.
// It implements the Store interface.
type MemoryStore struct {
	mu     sync.RWMutex
	agents map[string]*Agent
	jobs   map[string]*Job
	states map[string]*DesiredState
	epochs map[string]int64
}

// NewMemoryStore initializes a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		agents: make(map[string]*Agent),
		jobs:   make(map[string]*Job),
		states: make(map[string]*DesiredState),
		epochs: make(map[string]int64),
	}
}

// --- Agent Operations ---

func (s *MemoryStore) UpsertAgent(ctx context.Context, tenantID string, a *Agent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.TenantID = tenantID
	key := TenantKey(tenantID, ResourceAgent, a.NodeID)
	s.agents[key] = a
	return nil
}

func (s *MemoryStore) GetAgent(ctx context.Context, tenantID string, nodeID string) (*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := TenantKey(tenantID, ResourceAgent, nodeID)
	a, ok := s.agents[key]
	if !ok {
		return nil, nil // Return nil if not found
	}
	// Return copy
	agentCopy := *a
	return &agentCopy, nil
}

func (s *MemoryStore) ListAgents(ctx context.Context, tenantID string) ([]*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Agent, 0, len(s.agents))
	prefix := TenantPrefix(tenantID, ResourceAgent)

	for key, a := range s.agents {
		// Filter by key prefix to ensure isolation
		if len(key) >= len(prefix) && key[0:len(prefix)] == prefix {
			agentCopy := *a
			result = append(result, &agentCopy)
		}
	}
	return result, nil
}

func (s *MemoryStore) UpdateAgentHeartbeat(ctx context.Context, tenantID string, nodeID string, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := TenantKey(tenantID, ResourceAgent, nodeID)
	agent, ok := s.agents[key]
	if !ok {
		return errors.New("agent not found")
	}
	agent.LastHeartbeat = t
	return nil
}

// --- State Operations ---

func (s *MemoryStore) UpsertState(ctx context.Context, tenantID string, st *DesiredState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	st.TenantID = tenantID
	key := TenantKey(tenantID, ResourceState, st.StateID)
	s.states[key] = st
	return nil
}

func (s *MemoryStore) UpdateStateStatus(ctx context.Context, tenantID string, stateID string, status string, lastError string, lastChecked time.Time, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := TenantKey(tenantID, ResourceState, stateID)
	state, exists := s.states[key]
	if !exists {
		return errors.New("state not found")
	}
	if state.Version != expectedVersion {
		return errors.New("optimistic lock failure: state version changed")
	}

	state.Status = status
	state.LastError = lastError
	state.LastChecked = lastChecked
	return nil
}

func (s *MemoryStore) GetState(ctx context.Context, tenantID string, stateID string) (*DesiredState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := TenantKey(tenantID, ResourceState, stateID)
	st, ok := s.states[key]
	if !ok {
		return nil, nil
	}
	stateCopy := *st
	return &stateCopy, nil
}

func (s *MemoryStore) GetStateByNode(ctx context.Context, tenantID string, nodeID string) (*DesiredState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Linear scan (inefficient)
	// Filter by tenant
	for _, st := range s.states {
		if st.TenantID == tenantID && st.NodeID == nodeID {
			stateCopy := *st
			return &stateCopy, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) ListStates(ctx context.Context, tenantID string) ([]*DesiredState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*DesiredState, 0)
	for _, st := range s.states {
		if st.TenantID == tenantID {
			stateCopy := *st
			result = append(result, &stateCopy)
		}
	}
	return result, nil
}

// --- Job Operations ---

func (s *MemoryStore) CreateJob(ctx context.Context, tenantID string, j *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	j.TenantID = tenantID
	key := TenantKey(tenantID, ResourceJob, j.JobID)
	s.jobs[key] = j
	return nil
}

func (s *MemoryStore) UpdateJobStatus(ctx context.Context, tenantID string, jobID string, status string, exitCode int, stdout, stderr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := TenantKey(tenantID, ResourceJob, jobID)
	j, ok := s.jobs[key]
	if !ok {
		return errors.New("job not found")
	}

	j.Status = status
	if status == "running" {
		now := time.Now()
		j.StartedAt = &now
	} else if status == "completed" || status == "failed" {
		now := time.Now()
		j.FinishedAt = &now
		j.ExitCode = exitCode
		j.Stdout = stdout
		j.Stderr = stderr
	}
	return nil
}

func (s *MemoryStore) GetJob(ctx context.Context, tenantID string, jobID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := TenantKey(tenantID, ResourceJob, jobID)
	j, ok := s.jobs[key]
	if !ok {
		return nil, nil
	}
	jobCopy := *j
	return &jobCopy, nil
}

// ListStatesByStatus returns ALL states (global).
// Currently mostly used by Reconciler which might be global.
// If Reconciler becomes tenant-aware, this needs updating.
func (s *MemoryStore) ListStatesByStatus(ctx context.Context, status string, shardIndex int, shardCount int) ([]*DesiredState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var states []*DesiredState
	for _, state := range s.states {
		if state.Status == status {
			// Sharding Check
			if shardCount > 1 {
				h := fnvHash(state.NodeID)
				if int(h%uint32(shardCount)) != shardIndex {
					continue
				}
			}
			states = append(states, state)
		}
	}
	return states, nil
}

func (s *MemoryStore) CountStatesByStatus(ctx context.Context, tenantID string, status string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, state := range s.states {
		if state.TenantID == tenantID && state.Status == status {
			count++
		}
	}
	return count, nil
}

// Simple hash for memory store sharding simulation
func fnvHash(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h *= 16777619
		h ^= uint32(s[i])
	}
	return h
}

func (s *MemoryStore) ListJobs(ctx context.Context, tenantID string, nodeID string, limit int) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Job, 0)
	count := 0
	for _, j := range s.jobs {
		if j.TenantID == tenantID && j.NodeID == nodeID {
			jobCopy := *j
			result = append(result, &jobCopy)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *MemoryStore) ListJobsByTenant(ctx context.Context, tenantID string, limit int) ([]*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Job, 0)
	count := 0
	for _, j := range s.jobs {
		if j.TenantID == tenantID {
			jobCopy := *j
			result = append(result, &jobCopy)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	return result, nil
}

// --- Coordination Operations ---

func (s *MemoryStore) IncrementDurableEpoch(ctx context.Context, resourceID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newEpoch := s.epochs[resourceID] + 1
	s.epochs[resourceID] = newEpoch
	return newEpoch, nil
}

func (s *MemoryStore) GetDurableEpoch(ctx context.Context, resourceID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.epochs[resourceID], nil
}

// --- Idempotency Operations ---

func (s *MemoryStore) GetIdempotencyRecord(key string) (string, error) {
	// In-memory store doesn't persist idempotency records
	// This is a no-op for testing purposes
	return "", errors.New("not found")
}

func (s *MemoryStore) SetIdempotencyRecord(key string, value string, ttl time.Duration) error {
	// In-memory store doesn't persist idempotency records
	// This is a no-op for testing purposes
	return nil
}

// SetIdempotencyRecordNX atomically sets idempotency record if not exists (no-op for memory store)
func (s *MemoryStore) SetIdempotencyRecordNX(key string, value string, ttl time.Duration) error {
	return nil // No-op: idempotency should use Redis
}
