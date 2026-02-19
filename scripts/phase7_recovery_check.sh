#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

API_BASE="http://localhost:8090" # Load Balancer
TENANT_ID="default"

# Check 1: Exactly One Leader (with Retry)
# Must query ALL nodes because LB only hits one node (which might be a follower returning 0)
MAX_RETRIES=5
for i in $(seq 1 $MAX_RETRIES); do
    LEADER_COUNT=0
    for port in 8080 8081 8082; do
        # Each node reports 1 if it is leader, 0 otherwise
        # We need to sum these up.
        IS_LEADER=$(curl -s http://localhost:$port/metrics | grep "flux_leader_status" | grep " 1" | wc -l)
        LEADER_COUNT=$((LEADER_COUNT + IS_LEADER))
    done

    if [ "$LEADER_COUNT" -eq 1 ]; then
        break
    fi
    echo -e "${YELLOW}[WARN] Leader count is $LEADER_COUNT (expected 1). Retrying ($i/$MAX_RETRIES)...${NC}"
    sleep 2
done

if [ "$LEADER_COUNT" -ne 1 ]; then
    echo -e "${RED}[FAIL] Leader invariant violated: Found $LEADER_COUNT leaders (expected 1) after retries${NC}"
    exit 1
fi

# Check 2: Agents Reconnect (Active Agents > 0)
# We check Redis directly to ensure keys exist
AGENT_KEYS=$(docker exec fluxforge-redis redis-cli KEYS "fluxforge:tenants:*:agents:*" | wc -l)
if [ "$AGENT_KEYS" -eq 0 ]; then
   echo -e "${RED}[FAIL] No agents found in Redis${NC}"
   exit 1
fi

# Check 3: No Orphaned Locks
LOCK_COUNT=$(docker exec fluxforge-redis redis-cli KEYS "fluxforge:lock:*" | wc -l)
if [ "$LOCK_COUNT" -gt 100 ]; then
    echo -e "${RED}[FAIL] Lock leak detected: $LOCK_COUNT locks${NC}"
    exit 1
fi

echo -e "${GREEN}[PASS] Recovery Invariants Verified${NC}"
