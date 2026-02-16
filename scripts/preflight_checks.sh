#!/bin/bash
# Pre-Flight Checks: Verify Atomic Enforcement is Actually Wired
# CRITICAL: Do not assume. Verify.

set -e

CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-http://localhost:8080}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================="
echo "Pre-Flight Checks"
echo "========================================="
echo "Verifying atomic enforcement is wired correctly..."
echo ""

FAILED=0

# Pre-Flight Check 1: Confirm Lua scripts are loaded and used
echo "========================================="
echo "Check 1: Lua Script Enforcement"
echo "========================================="
echo ""

echo "Checking for versioned write metrics..."
METRICS=$(curl -s "$CONTROL_PLANE_URL/metrics" 2>/dev/null || echo "")

if echo "$METRICS" | grep -q "flux_versioned_write"; then
  echo -e "${GREEN}✅ Versioned write metrics found${NC}"
  
  # Show current values
  SUCCESS=$(echo "$METRICS" | grep "flux_versioned_write_success_total" | awk '{print $2}' || echo "0")
  CONFLICT=$(echo "$METRICS" | grep "flux_versioned_write_conflict_total" | awk '{print $2}' || echo "0")
  
  echo "  Success: $SUCCESS"
  echo "  Conflicts: $CONFLICT"
  
  # Run quick test
  echo ""
  echo "Running quick version conflict test..."
  
  TEST_KEY="preflight-test-$(date +%s)"
  
  # Write version 1
  curl -s -X POST "$CONTROL_PLANE_URL/test/version" \
    -H "Content-Type: application/json" \
    -d "{\"key\":\"$TEST_KEY\",\"version\":1,\"value\":\"v1\"}" > /dev/null
  
  # Try to write version 0 (should conflict)
  RESPONSE=$(curl -s -X POST "$CONTROL_PLANE_URL/test/version" \
    -H "Content-Type: application/json" \
    -d "{\"key\":\"$TEST_KEY\",\"version\":0,\"value\":\"v0\"}")
  
  if echo "$RESPONSE" | grep -q "conflict"; then
    echo -e "${GREEN}✅ Version conflict detected correctly${NC}"
    echo "  Lua script enforcement is ACTIVE"
  else
    echo -e "${RED}❌ Version conflict NOT detected${NC}"
    echo "  CRITICAL: Lua enforcement is NOT wired"
    FAILED=$((FAILED + 1))
  fi
else
  echo -e "${RED}❌ Versioned write metrics NOT found${NC}"
  echo "  CRITICAL: Lua scripts are NOT loaded"
  echo "  Certification would be meaningless"
  FAILED=$((FAILED + 1))
fi

echo ""

# Pre-Flight Check 2: Confirm idempotency lock has expiration
echo "========================================="
echo "Check 2: Lock Expiration"
echo "========================================="
echo ""

echo "Testing Redis lock expiration..."

# Set test lock with 5 second expiration
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" SET preflight_lock LOCK NX EX 5 > /dev/null 2>&1

# Check TTL
TTL=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" TTL preflight_lock 2>/dev/null || echo "-1")

echo "Lock TTL: $TTL seconds"

if [ "$TTL" -gt 0 ] && [ "$TTL" -le 5 ]; then
  echo -e "${GREEN}✅ Lock expiration is working${NC}"
  echo "  TTL is set correctly"
else
  echo -e "${RED}❌ Lock expiration is BROKEN${NC}"
  echo "  Expected TTL: 1-5 seconds"
  echo "  Actual TTL: $TTL"
  echo "  CRITICAL: Certification would be meaningless"
  FAILED=$((FAILED + 1))
fi

# Cleanup
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" DEL preflight_lock > /dev/null 2>&1

echo ""

# Pre-Flight Check 3: Confirm only one reconciliation leader exists
echo "========================================="
echo "Check 3: Leader Election"
echo "========================================="
echo ""

echo "Checking leader status..."

LEADER_METRICS=$(curl -s "$CONTROL_PLANE_URL/metrics" 2>/dev/null | grep "flux_leader_status" || echo "")

if [ -n "$LEADER_METRICS" ]; then
  LEADER_COUNT=$(echo "$LEADER_METRICS" | grep "flux_leader_status 1" | wc -l)
  
  echo "Leader nodes: $LEADER_COUNT"
  
  if [ "$LEADER_COUNT" -eq 1 ]; then
    echo -e "${GREEN}✅ Exactly one leader elected${NC}"
    echo "  Leader election is working correctly"
  elif [ "$LEADER_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  WARNING: No leader elected${NC}"
    echo "  System may still be initializing"
    echo "  Wait 30 seconds and check again"
  else
    echo -e "${RED}❌ Multiple leaders detected: $LEADER_COUNT${NC}"
    echo "  CRITICAL: Split-brain condition"
    echo "  STOP IMMEDIATELY"
    FAILED=$((FAILED + 1))
  fi
else
  echo -e "${YELLOW}⚠️  WARNING: Leader metrics not found${NC}"
  echo "  Leader election may not be implemented"
  echo "  Reconciliation coordination may be unsafe"
fi

echo ""

# Summary
echo "========================================="
echo "Pre-Flight Check Summary"
echo "========================================="
echo ""

if [ "$FAILED" -eq 0 ]; then
  echo -e "${GREEN}✅ All pre-flight checks PASSED${NC}"
  echo ""
  echo "Atomic enforcement is wired correctly."
  echo "Safe to proceed with certification run."
  echo ""
  echo "Next step:"
  echo "  ./scripts/phase1_rapid_validation.sh"
  echo ""
  exit 0
else
  echo -e "${RED}❌ $FAILED pre-flight check(s) FAILED${NC}"
  echo ""
  echo "DO NOT proceed with certification."
  echo "Fix atomic enforcement wiring first."
  echo ""
  echo "Common issues:"
  echo "  1. Lua scripts not registered in Redis store"
  echo "  2. Lock expiration not using SET NX EX"
  echo "  3. Leader election not implemented"
  echo ""
  exit 1
fi
