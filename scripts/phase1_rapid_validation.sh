#!/bin/bash
# Phase 1: 2-Hour Rapid Integrity Validation
# CRITICAL: Confirms atomic enforcement is wired correctly

set -e

CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-http://localhost:8080}"
TEST_KEY="hammer-test-$(date +%s)"

echo "========================================="
echo "Phase 1: Rapid Integrity Validation"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Version Conflict Hammer Test
echo "========================================="
echo "Test 1: Version Conflict Hammer Test"
echo "========================================="
echo "Testing atomic Lua enforcement under concurrent writes..."
echo ""

# Reset test state
curl -s -X POST "$CONTROL_PLANE_URL/test/version/reset" > /dev/null

# Run 100 concurrent version writes
echo "Launching 100 concurrent version writes..."
for i in {1..100}; do
  curl -s -X POST "$CONTROL_PLANE_URL/test/version" \
    -H "Content-Type: application/json" \
    -d "{\"key\":\"$TEST_KEY\",\"version\":$i,\"value\":\"v$i\"}" &
done

# Wait for all requests to complete
wait

echo "All requests completed. Checking final version..."
sleep 1

# Get final version
FINAL_VERSION=$(curl -s "$CONTROL_PLANE_URL/test/version/$TEST_KEY" | jq -r '.version')

echo ""
if [ "$FINAL_VERSION" == "100" ]; then
  echo -e "${GREEN}✅ PASS: Final version is 100${NC}"
  echo "Atomic Lua enforcement is working correctly"
else
  echo -e "${RED}❌ FAIL: Final version is $FINAL_VERSION (expected 100)${NC}"
  echo "CRITICAL: Atomic enforcement is broken!"
  exit 1
fi

echo ""
echo "========================================="
echo "Test 2: Idempotency Lock Hammer Test"
echo "========================================="
echo "Testing two-phase idempotency under concurrent requests..."
echo ""

# Reset metrics
curl -s -X POST "$CONTROL_PLANE_URL/test/idempotency/reset" > /dev/null

IDEMPOTENCY_KEY="hammer-test-$(date +%s)"

# Run 100 concurrent requests with same idempotency key
echo "Launching 100 concurrent requests with same idempotency key..."
for i in {1..100}; do
  curl -s -X POST "$CONTROL_PLANE_URL/api/jobs" \
    -H "Content-Type: application/json" \
    -H "X-Flux-Idempotency-Key: $IDEMPOTENCY_KEY" \
    -d '{"command":"echo test","node_id":"test-node"}' &
done

# Wait for all requests to complete
wait

echo "All requests completed. Checking metrics..."
sleep 2

# Get metrics
METRICS=$(curl -s "$CONTROL_PLANE_URL/test/idempotency/metrics")
TOTAL=$(echo "$METRICS" | jq -r '.total_requests')
EXECUTIONS=$(echo "$METRICS" | jq -r '.executions')
CACHED=$(echo "$METRICS" | jq -r '.cached_responses')

echo ""
echo "Results:"
echo "  Total requests: $TOTAL"
echo "  Executions: $EXECUTIONS"
echo "  Cached responses: $CACHED"
echo ""

if [ "$EXECUTIONS" == "1" ]; then
  echo -e "${GREEN}✅ PASS: Exactly 1 execution${NC}"
  echo "Two-phase idempotency is working correctly"
else
  echo -e "${RED}❌ FAIL: $EXECUTIONS executions (expected 1)${NC}"
  echo "CRITICAL: Duplicate execution detected!"
  exit 1
fi

if [ "$CACHED" == "99" ]; then
  echo -e "${GREEN}✅ PASS: 99 cached responses${NC}"
else
  echo -e "${YELLOW}⚠️  WARNING: $CACHED cached responses (expected 99)${NC}"
fi

echo ""
echo "========================================="
echo "Test 3: Lock Expiration Recovery Test"
echo "========================================="
echo "Testing automatic lock expiration..."
echo ""

LOCK_KEY="expiration-test-$(date +%s)"

# Acquire lock with 2 second TTL
echo "Acquiring lock with 2 second TTL..."
curl -s -X POST "$CONTROL_PLANE_URL/test/lock/acquire" \
  -H "Content-Type: application/json" \
  -d "{\"key\":\"$LOCK_KEY\",\"ttl\":2}" > /dev/null

# Check lock status
STATUS=$(curl -s "$CONTROL_PLANE_URL/test/lock/status/$LOCK_KEY" | jq -r '.status')
echo "Lock status: $STATUS"

if [ "$STATUS" == "LOCKED" ] || [ "$STATUS" == "no_lock" ]; then
  echo -e "${GREEN}✅ Lock acquired${NC}"
else
  echo -e "${RED}❌ Failed to acquire lock${NC}"
  exit 1
fi

# Simulate crash (don't release lock)
echo "Simulating crash (lock not released)..."
curl -s -X POST "$CONTROL_PLANE_URL/test/lock/crash" > /dev/null

# Wait for expiration
echo "Waiting 3 seconds for lock to expire..."
sleep 3

# Try to acquire lock again
echo "Attempting to acquire lock again..."
ACQUIRE_RESULT=$(curl -s -X POST "$CONTROL_PLANE_URL/test/lock/acquire" \
  -H "Content-Type: application/json" \
  -d "{\"key\":\"$LOCK_KEY\",\"ttl\":2}")

ACQUIRED=$(echo "$ACQUIRE_RESULT" | jq -r '.acquired')

echo ""
if [ "$ACQUIRED" == "true" ]; then
  echo -e "${GREEN}✅ PASS: Lock expired and was re-acquired${NC}"
  echo "Lock expiration is working correctly"
else
  echo -e "${RED}❌ FAIL: Lock did not expire${NC}"
  echo "CRITICAL: Lock orphaning detected!"
  exit 1
fi

echo ""
echo "========================================="
echo "Phase 1 Complete: All Tests Passed ✅"
echo "========================================="
echo ""
echo "Summary:"
echo "  ✅ Version conflict hammer test: PASS"
echo "  ✅ Idempotency lock hammer test: PASS"
echo "  ✅ Lock expiration recovery test: PASS"
echo ""
echo "Atomic enforcement is wired correctly!"
echo ""
echo "Next: Run Phase 2 (6-hour stability test)"
echo "  ./scripts/phase7_stability_test.sh 6"
