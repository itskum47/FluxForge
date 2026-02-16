package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotencyState represents the two-phase idempotency state
type IdempotencyState string

const (
	// CRITICAL: Two-phase state distinction
	IdempotencyStateLocked IdempotencyState = "LOCKED" // Execution in progress
	IdempotencyStateResult IdempotencyState = "RESULT" // Execution complete
)

// IdempotencyResult represents cached execution result
type IdempotencyResult struct {
	State      IdempotencyState  `json:"state"`
	StatusCode int               `json:"status_code,omitempty"`
	Body       []byte            `json:"body,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

const (
	// CRITICAL: Lock expiration formula
	// lock_expiry = max_expected_execution_time * 2
	// Example: task timeout = 300s, lock expiry = 600s
	maxExpectedExecutionTime = 300 * time.Second
	lockTTL                  = 2 * maxExpectedExecutionTime // 600s
	resultTTL                = 24 * time.Hour
)

// GetIdempotencyState retrieves current idempotency state
// CRITICAL: Handles both LOCKED and RESULT states
func (s *RedisStore) GetIdempotencyState(ctx context.Context, key string) (*IdempotencyResult, error) {
	// Check result first
	resultKey := "idempotency:result:" + key
	resultData, err := s.client.Get(ctx, resultKey).Result()

	if err == nil {
		// Result exists
		var result IdempotencyResult
		err = json.Unmarshal([]byte(resultData), &result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}

	if err != redis.Nil {
		return nil, err
	}

	// Check lock
	lockKey := "idempotency:lock:" + key
	lockData, err := s.client.Get(ctx, lockKey).Result()

	if err == redis.Nil {
		// No lock, no result
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	// Lock exists
	var locked IdempotencyResult
	err = json.Unmarshal([]byte(lockData), &locked)
	if err != nil {
		return nil, err
	}

	return &locked, nil
}

// StoreIdempotencyResult stores execution result
// CRITICAL: Transitions from LOCKED to RESULT state
func (s *RedisStore) StoreIdempotencyResult(ctx context.Context, key string, result *IdempotencyResult, ttl time.Duration) error {
	resultKey := "idempotency:result:" + key

	// Set state to RESULT
	result.State = IdempotencyStateResult
	result.CreatedAt = time.Now()

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// Store result with TTL
	err = s.client.Set(ctx, resultKey, data, ttl).Err()
	if err != nil {
		return err
	}

	// Release lock (result now available)
	// Use key as ownerID for idempotency locks
	return s.ReleaseLock(ctx, "idempotency:lock:"+key, key)
}

// WaitForIdempotencyResult waits for another request to complete
// CRITICAL: Handles LOCKED state by polling
func (s *RedisStore) WaitForIdempotencyResult(ctx context.Context, key string, timeout time.Duration) (*IdempotencyResult, error) {
	deadline := time.Now().Add(timeout)
	backoff := 100 * time.Millisecond
	maxBackoff := 2 * time.Second

	for time.Now().Before(deadline) {
		state, err := s.GetIdempotencyState(ctx, key)
		if err != nil {
			return nil, err
		}

		if state == nil {
			// Lock expired without result - execution failed
			return nil, fmt.Errorf("idempotency lock expired without result")
		}

		if state.State == IdempotencyStateResult {
			// Result available
			return state, nil
		}

		// Still LOCKED, wait with exponential backoff
		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	return nil, fmt.Errorf("timeout waiting for idempotent request to complete")
}

// ExecuteIdempotent implements the complete two-phase idempotency pattern
// CRITICAL: LOCK → EXECUTE → RESULT
func (s *RedisStore) ExecuteIdempotent(
	ctx context.Context,
	key string,
	execute func(context.Context) (*IdempotencyResult, error),
) (*IdempotencyResult, error) {

	// STEP 1: Check existing state
	existing, err := s.GetIdempotencyState(ctx, key)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		if existing.State == IdempotencyStateResult {
			// Already completed
			return existing, nil
		}

		if existing.State == IdempotencyStateLocked {
			// Another request executing, wait
			return s.WaitForIdempotencyResult(ctx, key, 30*time.Second)
		}
	}

	// STEP 2: Acquire lock
	// Use key as ownerID for idempotency locks
	acquired, err := s.AcquireLock(ctx, "idempotency:lock:"+key, key, lockTTL)
	if err != nil {
		return nil, err
	}

	if !acquired {
		// Another request acquired lock, wait
		return s.WaitForIdempotencyResult(ctx, key, 30*time.Second)
	}

	// STEP 3: Double-check (another may have completed while we acquired lock)
	existing, err = s.GetIdempotencyState(ctx, key)
	if err != nil {
		s.ReleaseLock(ctx, "idempotency:lock:"+key, key)
		return nil, err
	}

	if existing != nil && existing.State == IdempotencyStateResult {
		s.ReleaseLock(ctx, "idempotency:lock:"+key, key)
		return existing, nil
	}

	// STEP 4: Execute (only one request reaches here)
	result, err := execute(ctx)
	if err != nil {
		s.ReleaseLock(ctx, "idempotency:lock:"+key, key)
		return nil, err
	}

	// STEP 5: Store result (transitions LOCKED → RESULT)
	err = s.StoreIdempotencyResult(ctx, key, result, resultTTL)
	if err != nil {
		// Log error but return result (execution succeeded)
		fmt.Printf("[IDEMPOTENCY] Failed to store result: %v\n", err)
	}

	return result, nil
}
