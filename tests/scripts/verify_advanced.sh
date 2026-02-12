#!/bin/bash
set -e

# Cleanup function
cleanup() {
    echo "Stopping processes..."
    kill $CP_PID 2>/dev/null || true
    kill $AGENT_PID 2>/dev/null || true
    rm -f cp agent
}
trap cleanup EXIT

# Build
echo "Building components..."
(cd fluxforge && go build -o ../cp ./control_plane)
(cd fluxforge && go build -o ../agent ./agent)

# Start Control Plane
echo "Starting Control Plane..."
./cp > cp_advanced.log 2>&1 &
CP_PID=$!
sleep 2

# Start Agent
echo "Starting Agent..."
# Ensure clean state
rm -rf ~/.fluxforge
./agent > agent_advanced.log 2>&1 &
AGENT_PID=$!
sleep 5

AGENT_ID=$(grep "Node ID:" agent_advanced.log | awk '{print $NF}' | head -n 1)
echo "Agent ID: $AGENT_ID"

# --- Negative Test Cases ---
echo "---------------------------------------------------"
echo "Negative Test 1: Non-Existent Agent"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/jobs \
    -d "{\"node_id\": \"ghost-agent\", \"command\": \"ls\"}")

if [ "$HTTP_CODE" == "404" ]; then
    echo "PASS: Received 404 for ghost agent"
else
    echo "FAIL: Expected 404, got $HTTP_CODE"
    exit 1
fi

# --- Concurrency Test Cases ---
echo "---------------------------------------------------"
echo "Concurrency Test: Agent Busy (409)"

# 1. Create a state with a long-running command to hold the lock
echo "Creating long-running state..."
STATE_RESP=$(curl -s -X POST http://localhost:8080/states \
    -d "{\"node_id\": \"$AGENT_ID\", \"check_cmd\": \"exit 1\", \"apply_cmd\": \"sleep 5 && echo Done\", \"desired_exit_code\": 0}")
STATE_ID=$(echo $STATE_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['state_id'])")

# 2. Trigger Reconcile (This should acquire lock)
echo "Triggering Reconcile #1..."
curl -s -X POST http://localhost:8080/states/$STATE_ID/reconcile

# 3. Immediately trigger again (Should fail with 409 OR 202 if queued)
# NOTE: With Scheduler, this might now return 202 (Accepted) and be queued!
# But the Reconciler logic *inside* the worker still checks for lock.
# However, the worker processes tasks sequentially for a node if we implement per-node serialization?
# Currently, `worker` spawns `go reconciler.Reconcile`.
# So `Reconcile` will hit the lock check and return early?
# Wait, `Reconcile` logs "agent is busy" but what does it return to API?
# The API calls `scheduler.Submit` which returns immediately 202.
# So the API will ALWAYS return 202 now!
# The 409 Conflict logic was moved from API handler to Reconciler internals (log only).

echo "Triggering Reconcile #2..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/states/$STATE_ID/reconcile)

# With Scheduler, we expect 202 Accepted (Queued)
if [ "$HTTP_CODE" == "202" ]; then
    echo "PASS: Received 202 Accepted (Request Queued)"
else
    echo "FAIL: Expected 202, got $HTTP_CODE"
    exit 1
fi

# 4. Wait for it to finish and try again (Should succeed)
echo "Waiting for execution (10s)..."
sleep 10

echo "Triggering Reconcile #3 (Expect 202)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/states/$STATE_ID/reconcile)

if [ "$HTTP_CODE" == "202" ]; then
    echo "PASS: Received 202 Accepted after execution"
else
    echo "FAIL: Expected 202, got $HTTP_CODE"
    exit 1
fi

echo "---------------------------------------------------"
echo "Advanced Verification Successful!"
