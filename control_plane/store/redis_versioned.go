package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// VersionedValue represents a value with version for conflict detection
type VersionedValue struct {
	Value     interface{} `json:"value"`
	Version   int64       `json:"version"`
	Timestamp int64       `json:"timestamp"` // Unix timestamp
}

// CRITICAL: Lua script for ATOMIC versioned SET
// Single instruction from Redis perspective - no race conditions possible
const versionedSetScript = `
-- KEYS[1] = key
-- ARGV[1] = new_value (JSON)
-- ARGV[2] = new_version
-- ARGV[3] = ttl (seconds, 0 = no expiry)

local current_version = redis.call("HGET", KEYS[1], "version")

-- Only set if new version is greater (or no existing version)
if not current_version or tonumber(ARGV[2]) > tonumber(current_version) then
    redis.call("HMSET", KEYS[1],
        "value", ARGV[1],
        "version", ARGV[2],
        "timestamp", ARGV[4])
    
    -- Set TTL if specified
    if tonumber(ARGV[3]) > 0 then
        redis.call("EXPIRE", KEYS[1], ARGV[3])
    end
    
    return 1  -- Success
else
    return 0  -- Version conflict
end
`

// CRITICAL: Lua script for ATOMIC versioned GET
const versionedGetScript = `
-- KEYS[1] = key

local value = redis.call("HGET", KEYS[1], "value")
local version = redis.call("HGET", KEYS[1], "version")
local timestamp = redis.call("HGET", KEYS[1], "timestamp")

if not value then
    return nil
end

return cjson.encode({
    value = value,
    version = tonumber(version),
    timestamp = tonumber(timestamp)
})
`

// SetVersioned atomically sets value only if version is newer
// CRITICAL: Single atomic operation - no GET/SET race condition
// Uses preloaded Lua script SHA for performance
func (s *RedisStore) SetVersioned(ctx context.Context, key string, value VersionedValue, ttl time.Duration) error {
	// Serialize value
	valueJSON, err := json.Marshal(value.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Execute Lua script atomically using preloaded SHA
	result, err := s.client.EvalSha(ctx, s.versionedSetSHA,
		[]string{key},
		string(valueJSON),
		value.Version,
		int(ttl.Seconds()),
		value.Timestamp,
	).Result()

	// Handle NOSCRIPT error (Redis restarted, scripts lost)
	if err != nil && err.Error() == "NOSCRIPT No matching script. Please use EVAL." {
		// Reload script and retry
		s.versionedSetSHA, _ = s.client.ScriptLoad(ctx, versionedSetScript).Result()
		result, err = s.client.EvalSha(ctx, s.versionedSetSHA,
			[]string{key},
			string(valueJSON),
			value.Version,
			int(ttl.Seconds()),
			value.Timestamp,
		).Result()
	}

	if err != nil {
		return fmt.Errorf("failed to execute versioned set: %w", err)
	}

	// Check result
	wasSet, ok := result.(int64)
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	if wasSet == 0 {
		return fmt.Errorf("version conflict: newer version exists in Redis")
	}

	return nil
}

// GetVersioned retrieves versioned value atomically
// Uses preloaded Lua script SHA for performance
func (s *RedisStore) GetVersioned(ctx context.Context, key string) (*VersionedValue, error) {
	// Execute Lua script atomically using preloaded SHA
	result, err := s.client.EvalSha(ctx, s.versionedGetSHA, []string{key}).Result()

	// Handle NOSCRIPT error (Redis restarted, scripts lost)
	if err != nil && err.Error() == "NOSCRIPT No matching script. Please use EVAL." {
		// Reload script and retry
		s.versionedGetSHA, _ = s.client.ScriptLoad(ctx, versionedGetScript).Result()
		result, err = s.client.EvalSha(ctx, s.versionedGetSHA, []string{key}).Result()
	}

	if err == redis.Nil || result == nil {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get versioned value: %w", err)
	}

	// Parse result
	resultStr, ok := result.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	var value VersionedValue
	err = json.Unmarshal([]byte(resultStr), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return &value, nil
}

// CompareAndSetVersioned implements compare-and-swap with version check
// CRITICAL: Atomic CAS operation
func (s *RedisStore) CompareAndSetVersioned(ctx context.Context, key string, expectedVersion int64, newValue VersionedValue, ttl time.Duration) (bool, error) {
	const casScript = `
-- KEYS[1] = key
-- ARGV[1] = expected_version
-- ARGV[2] = new_value (JSON)
-- ARGV[3] = new_version
-- ARGV[4] = ttl

local current_version = redis.call("HGET", KEYS[1], "version")

-- Check version matches
if current_version and tonumber(current_version) ~= tonumber(ARGV[1]) then
    return 0  -- Version mismatch
end

-- Set new value
redis.call("HMSET", KEYS[1],
    "value", ARGV[2],
    "version", ARGV[3],
    "timestamp", ARGV[5])

if tonumber(ARGV[4]) > 0 then
    redis.call("EXPIRE", KEYS[1], ARGV[4])
end

return 1  -- Success
`

	valueJSON, err := json.Marshal(newValue.Value)
	if err != nil {
		return false, err
	}

	result, err := s.client.Eval(ctx, casScript,
		[]string{key},
		expectedVersion,
		string(valueJSON),
		newValue.Version,
		int(ttl.Seconds()),
		newValue.Timestamp,
	).Result()

	if err != nil {
		return false, err
	}

	success, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type")
	}

	return success == 1, nil
}
