#!/bin/bash
# Phase 4: Leader Failover During Reconciliation
# Ultimate physics test - validates epoch enforcement

set -e

CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-http://localhost:8080}"

echo "========================================="
echo "Phase 4: Leader Failover Test"
echo "========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Start reconciliation load
echo "Starting reconciliation load..."
for i in {1..1000}; do
  curl -s -X POST "$CONTROL_PLANE_URL/test/version" \
    -H "Content-Type: application/json" \
    -d "{\"key\":\"failover-test-$i\",\"version\":1,\"value\":\"v1\"}" &
done

sleep 2

# Kill leader
echo "Killing leader node..."
docker kill fluxforge-control-plane-1

sleep 5

# Wait for new leader election
echo "Waiting for new leader election..."
sleep 10

# Restart old leader
echo "Restarting old leader..."
docker start fluxforge-control-plane-1

sleep 5

# Verify no corruption
echo "Verifying data integrity..."
ERRORS=0

for i in {1..100}; do
  VERSION=$(curl -s "$CONTROL_PLANE_URL/test/version/failover-test-$i" | jq -r '.version // 0')
  if [ "$VERSION" -ne 1 ]; then
    echo "Corruption detected: key failover-test-$i has version $VERSION"
    ERRORS=$((ERRORS + 1))
  fi
done

echo ""
if [ "$ERRORS" -eq 0 ]; then
  echo -e "${GREEN}✅ Leader failover test PASSED${NC}"
  echo "No data corruption detected"
  exit 0
else
  echo -e "${RED}❌ Leader failover test FAILED${NC}"
  echo "$ERRORS corrupted keys detected"
  exit 1
fi
