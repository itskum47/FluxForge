#!/bin/bash
# Certification Run Monitor
# What to watch during the 9-hour run

CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-http://localhost:8080}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================="
echo "Certification Run Monitor"
echo "========================================="
echo ""
echo "Four critical signals to watch:"
echo ""

# Signal 1: Memory curve must plateau
echo "1. Memory Usage (must plateau, not climb forever)"
echo "----------------------------------------"
MEMORY=$(docker stats fluxforge-control-plane-1 --no-stream --format "{{.MemUsage}}" 2>/dev/null || echo "N/A")
echo "  Current: $MEMORY"
echo "  Check every hour - continuous climb = leak"
echo ""

# Signal 2: Version conflict counter
echo "2. Version Conflict Counter (must increase during hammer test)"
echo "----------------------------------------"
CONFLICTS=$(curl -s "$CONTROL_PLANE_URL/metrics" | grep "flux_versioned_write_conflict_total" | awk '{print $2}' || echo "0")
echo "  Conflicts: $CONFLICTS"
echo "  Conflicts > 0 during hammer test = good"
echo "  Conflicts = 0 during hammer test = enforcement NOT active"
echo ""

# Signal 3: Idempotency lock contention
echo "3. Idempotency Lock Contention (must show contention)"
echo "----------------------------------------"
METRICS=$(curl -s "$CONTROL_PLANE_URL/test/idempotency/metrics" 2>/dev/null)
if [ -n "$METRICS" ]; then
  EXECUTIONS=$(echo "$METRICS" | jq -r '.executions // 0')
  CACHED=$(echo "$METRICS" | jq -r '.cached_responses // 0')
  echo "  Executions: $EXECUTIONS"
  echo "  Cached: $CACHED"
  echo "  Expected: executions=1, cached=99 (proves exactly-once)"
else
  echo "  Metrics not available yet"
fi
echo ""

# Signal 4: Leader changes
echo "4. Leader Changes (must recover cleanly)"
echo "----------------------------------------"
LEADER_CHANGES=$(curl -s "$CONTROL_PLANE_URL/metrics" | grep "flux_leader_changes_total" | awk '{print $2}' || echo "0")
echo "  Leader changes: $LEADER_CHANGES"
echo "  System must stabilize after each change"
echo ""

echo "========================================="
echo "Important Rules"
echo "========================================="
echo ""
echo "❗ DO NOT intervene during the run"
echo "❗ Let chaos happen"
echo "❗ Let failures attempt to occur"
echo "❗ If FluxForge survives without intervention → certification valid"
echo "❗ If you help it survive → certification fake"
echo ""
echo "Run this monitor periodically to check progress."
echo "Do not rush. Do not interrupt. Do not optimize mid-run."
echo ""
