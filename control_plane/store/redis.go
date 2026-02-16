package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"hash/fnv"

	"github.com/itskum47/FluxForge/control_plane/observability"
	"github.com/redis/go-redis/v9"
)

// RedisStore implements the Store interface using Redis.
type RedisStore struct {
	client *redis.Client

	// Preloaded Lua script SHAs for atomic operations
	versionedSetSHA string
	versionedGetSHA string
}

func NewRedisStore(addr string, password string, db int) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	// CRITICAL: Preload all Lua scripts for atomic operations
	// This avoids sending script text over network on every call
	versionedSetSHA, err := client.ScriptLoad(ctx, versionedSetScript).Result()
	if err != nil {
		return nil, errors.New("failed to preload versioned set script: " + err.Error())
	}

	versionedGetSHA, err := client.ScriptLoad(ctx, versionedGetScript).Result()
	if err != nil {
		return nil, errors.New("failed to preload versioned get script: " + err.Error())
	}

	return &RedisStore{
		client:          client,
		versionedSetSHA: versionedSetSHA,
		versionedGetSHA: versionedGetSHA,
	}, nil
}

// AcquireLock attempts to acquire a distributed lock.
// It uses SET key value NX EX ttl.
func (s *RedisStore) AcquireLock(ctx context.Context, key string, ownerID string, ttl time.Duration) (bool, error) {
	// Phase 5.1: Track Redis latency
	start := time.Now()
	defer func() {
		observability.RedisLatency.Observe(time.Since(start).Seconds())
	}()

	success, err := s.client.SetNX(ctx, key, ownerID, ttl).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

// RenewLock extends the TTL if the lock is held by ownerID.
// It uses a Lua script to ensure atomicity.
func (s *RedisStore) RenewLock(ctx context.Context, key string, ownerID string, ttl time.Duration) (bool, error) {
	// Phase 5.1: Track Redis latency
	start := time.Now()
	defer func() {
		observability.RedisLatency.Observe(time.Since(start).Seconds())
	}()

	// Lua script: if get(key) == ownerID then return expire(key, ttl) else return 0
	// TTL in seconds for EXPIRE command? redis-go uses duration for Set, but Expire takes duration.
	// Lua Expire takes seconds (integer). PEXPIRE takes millis.
	// Let's use PEXPIRE for precision.
	// Lua script: detailed diagnostics
	// Returns:
	// 1: Success (TTL extended)
	// 0: PEXPIRE failed (key missing/expired between checks? rare)
	// -1: Key missing (GET returned nil/false)
	// -2: Owner mismatch
	// Lua script: detailed diagnostics
	// Returns:
	// 1: Success (TTL extended)
	// 0: PEXPIRE failed (Key missing? Should be caught by check, but PEXPIRE returns 0 if key missing)
	// -1: Key missing (GET returned nil/false)
	// -2: Owner mismatch
	scriptP := `
		local val = redis.call("get", KEYS[1])
		if not val then
			return -1
		end
		if val == ARGV[1] then
			return redis.call("pexpire", KEYS[1], tonumber(ARGV[2]))
		else
			return -2
		end
	`
	res, err := s.client.Eval(ctx, scriptP, []string{key}, ownerID, int64(ttl/time.Millisecond)).Result()
	if err != nil {
		return false, err
	}

	if val, ok := res.(int64); ok {
		if val == 1 {
			return true, nil
		}
		if val == 0 {
			return false, nil
		}
		if val == -1 {
			return false, nil // Key missing
		}
		if val == -2 {
			return false, nil // Owner mismatch
		}
	}
	return false, errors.New("unexpected return type from lua script")
}

// ReleaseLock releases the lock if held by ownerID.
func (s *RedisStore) ReleaseLock(ctx context.Context, key string, ownerID string) error {
	// Phase 5.1: Track Redis latency
	start := time.Now()
	defer func() {
		observability.RedisLatency.Observe(time.Since(start).Seconds())
	}()

	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := s.client.Eval(ctx, script, []string{key}, ownerID).Result()
	return err
}

// GetLockOwner returns current owner.
func (s *RedisStore) GetLockOwner(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// --- Lease Implementation (Reuse Logic) ---

func (s *RedisStore) AcquireLease(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	// Same as AcquireLock but explicit naming for Lease semantics
	return s.AcquireLock(ctx, key, value, ttl)
}

func (s *RedisStore) RenewLease(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return s.RenewLock(ctx, key, value, ttl)
}

func (s *RedisStore) ReleaseLease(ctx context.Context, key string, value string) error {
	return s.ReleaseLock(ctx, key, value)
}

func (s *RedisStore) IsLeaseOwner(ctx context.Context, key string, value string) (bool, error) {
	val, err := s.GetLockOwner(ctx, key)
	if err != nil {
		return false, err
	}
	return val == value, nil
}

// IncrementEpoch increments the epoch counter for the given key.
// It uses a separate key suffixed with ":epoch".
func (s *RedisStore) IncrementEpoch(ctx context.Context, key string) (int64, error) {
	epochKey := key + ":epoch"
	return s.client.Incr(ctx, epochKey).Result()
}

// ScanLocks returns keys matching the pattern.
func (s *RedisStore) ScanLocks(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// --- Generic Key-Value Operations (Idempotency) ---

func (s *RedisStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	val, err := s.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // Not found
	}
	return val, err
}

// GetIdempotencyRecord retrieves a cached idempotency response from Redis
func (s *RedisStore) GetIdempotencyRecord(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	defer func() {
		observability.RedisLatency.Observe(time.Since(start).Seconds())
	}()

	val, err := s.client.Get(ctx, "idempotency:"+key).Result()
	if err == redis.Nil {
		return "", errors.New("not found")
	}
	return val, err
}

// SetIdempotencyRecord stores an idempotency response in Redis with TTL
func (s *RedisStore) SetIdempotencyRecord(key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	defer func() {
		observability.RedisLatency.Observe(time.Since(start).Seconds())
	}()

	return s.client.Set(ctx, "idempotency:"+key, value, ttl).Err()
}

// --- Store Interface Implementation (Stubs for Coordination usage) ---

func (s *RedisStore) UpsertAgent(ctx context.Context, tenantID string, agent *Agent) error {
	agent.TenantID = tenantID // Enforce binding
	data, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}
	key := TenantKey(tenantID, ResourceAgent, agent.NodeID)
	return s.client.Set(ctx, key, data, 0).Err()
}

func (s *RedisStore) GetAgent(ctx context.Context, tenantID string, nodeID string) (*Agent, error) {
	key := TenantKey(tenantID, ResourceAgent, nodeID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, err
	}
	var agent Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent: %w", err)
	}
	return &agent, nil
}

func (s *RedisStore) ListAgents(ctx context.Context, tenantID string) ([]*Agent, error) {
	// Scan specific tenant namespace
	match := TenantPrefix(tenantID, ResourceAgent) + "*"
	iter := s.client.Scan(ctx, 0, match, 0).Iterator()
	var agents []*Agent
	for iter.Next(ctx) {
		data, err := s.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var agent Agent
		if err := json.Unmarshal(data, &agent); err == nil {
			agents = append(agents, &agent)
		}
	}
	return agents, iter.Err()
}

func (s *RedisStore) UpdateAgentHeartbeat(ctx context.Context, tenantID string, nodeID string, t time.Time) error {
	agent, err := s.GetAgent(ctx, tenantID, nodeID)
	if err != nil {
		return err
	}
	if agent == nil {
		return fmt.Errorf("agent not found: %s", nodeID)
	}
	agent.LastHeartbeat = t
	agent.Status = "active"
	return s.UpsertAgent(ctx, tenantID, agent)
}

func (s *RedisStore) UpsertState(ctx context.Context, tenantID string, state *DesiredState) error {
	state.TenantID = tenantID
	return errors.New("RedisStore.UpsertState not implemented")
}

func (s *RedisStore) UpdateStateStatus(ctx context.Context, tenantID string, stateID string, status string, lastError string, lastChecked time.Time, expectedVersion int) error {
	return errors.New("RedisStore.UpdateStateStatus not implemented")
}

func (s *RedisStore) GetState(ctx context.Context, tenantID string, stateID string) (*DesiredState, error) {
	return nil, errors.New("RedisStore.GetState not implemented")
}

func (s *RedisStore) GetStateByNode(ctx context.Context, tenantID string, nodeID string) (*DesiredState, error) {
	return nil, errors.New("RedisStore.GetStateByNode not implemented")
}

func (s *RedisStore) ListStates(ctx context.Context, tenantID string) ([]*DesiredState, error) {
	return nil, errors.New("RedisStore.ListStates not implemented")
}

// ListStatesByStatus returns all states with the given status, filtered by shard
func (s *RedisStore) ListStatesByStatus(ctx context.Context, status string, shardIndex int, shardCount int) ([]*DesiredState, error) {
	if shardCount <= 0 {
		return nil, errors.New("shardCount must be > 0")
	}

	// Global scan of all tenant states: fluxforge:tenants:*:states:*
	match := "fluxforge:tenants:*:states:*"
	iter := s.client.Scan(ctx, 0, match, 0).Iterator()
	var states []*DesiredState

	for iter.Next(ctx) {
		key := iter.Val()
		// Format: fluxforge:tenants:{tid}:states:{sid}
		parts := strings.Split(key, ":")
		if len(parts) < 5 {
			continue
		}
		// parts[2] is tenantID, parts[4] is stateID
		stateID := parts[4]

		// 2. Apply Sharding Filter based on stateID
		h := fnv.New32a()
		h.Write([]byte(stateID))
		if int(h.Sum32())%shardCount != shardIndex {
			continue // Skip if not owned by this shard
		}

		// 3. Get State and filter by Status
		// Note: We need GetState logic here but GetState is stubbed.
		// So we assume we can GET directly using the key.
		data, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			log.Printf("ListStatesByStatus: Failed to get state %s: %v", stateID, err)
			continue
		}
		var state DesiredState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		if state.Status == status {
			states = append(states, &state)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan states: %w", err)
	}

	return states, nil
}

func (s *RedisStore) CountStatesByStatus(ctx context.Context, tenantID string, status string) (int, error) {
	// Scan specific tenant namespace: fluxforge:tenants:{tid}:states:*
	match := TenantPrefix(tenantID, ResourceState) + "*"
	iter := s.client.Scan(ctx, 0, match, 0).Iterator()
	count := 0
	for iter.Next(ctx) {
		// We need to check the status inside the value
		// Only counting keys is not enough if we filter by status.
		// Optimization: If status is "", count all?
		// For dashboard, we usually want "pending", "drifted".
		val, err := s.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var state DesiredState
		if err := json.Unmarshal(val, &state); err != nil {
			continue
		}
		if state.Status == status {
			count++
		}
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *RedisStore) CreateJob(ctx context.Context, tenantID string, job *Job) error {
	job.TenantID = tenantID
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}
	key := TenantKey(tenantID, ResourceJob, job.JobID)
	return s.client.Set(ctx, key, data, 0).Err()
}

func (s *RedisStore) UpdateJobStatus(ctx context.Context, tenantID string, jobID string, status string, exitCode int, stdout, stderr string) error {
	job, err := s.GetJob(ctx, tenantID, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return fmt.Errorf("job not found: %s", jobID)
	}
	job.Status = status
	job.ExitCode = exitCode
	job.Stdout = stdout
	job.Stderr = stderr
	return s.CreateJob(ctx, tenantID, job) // Reuse Set
}

func (s *RedisStore) GetJob(ctx context.Context, tenantID string, jobID string) (*Job, error) {
	key := TenantKey(tenantID, ResourceJob, jobID)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var job Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}
	return &job, nil
}

func (s *RedisStore) ListJobs(ctx context.Context, tenantID string, nodeID string, limit int) ([]*Job, error) {
	match := TenantPrefix(tenantID, ResourceJob) + "*"
	iter := s.client.Scan(ctx, 0, match, 0).Iterator()
	var jobs []*Job
	for iter.Next(ctx) {
		data, err := s.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var job Job
		if err := json.Unmarshal(data, &job); err == nil {
			if job.NodeID == nodeID {
				jobs = append(jobs, &job)
			}
		}
		if limit > 0 && len(jobs) >= limit {
			break
		}
	}
	return jobs, iter.Err()
}

func (s *RedisStore) ListJobsByTenant(ctx context.Context, tenantID string, limit int) ([]*Job, error) {
	// New method implementation
	match := TenantPrefix(tenantID, ResourceJob) + "*"
	iter := s.client.Scan(ctx, 0, match, 0).Iterator()
	var jobs []*Job
	for iter.Next(ctx) {
		data, err := s.client.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		var job Job
		if err := json.Unmarshal(data, &job); err == nil {
			jobs = append(jobs, &job)
		}
		if limit > 0 && len(jobs) >= limit {
			break
		}
	}
	return jobs, iter.Err()
}

func (s *RedisStore) IncrementDurableEpoch(ctx context.Context, resourceID string) (int64, error) {
	// Re-route to IncrementEpoch (legacy name)
	return s.IncrementEpoch(ctx, resourceID)
}

func (s *RedisStore) GetDurableEpoch(ctx context.Context, resourceID string) (int64, error) {
	// Simple GET
	val, err := s.client.Get(ctx, resourceID+":epoch").Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (s *RedisStore) SetIdempotencyRecordNX(key string, value string, ttl time.Duration) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use SET NX
	ok, err := s.client.SetNX(ctx, "idempotency:"+key, value, ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("key exists")
	}
	return nil
}
