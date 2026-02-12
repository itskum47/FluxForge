#!/bin/bash
set -e

# Cleanup function
cleanup() {
    echo "Stopping processes..."
    kill $CP_PID 2>/dev/null || true
    kill $AGENT_PID 2>/dev/null || true
    rm -f cp agent /tmp/fluxforge_test_file
}
trap cleanup EXIT

# Build
echo "Building components..."
(cd fluxforge && go build -o ../cp ./control_plane)
(cd fluxforge && go build -o ../agent ./agent)

# Start Control Plane
echo "Starting Control Plane..."
./cp > cp_conv.log 2>&1 &
CP_PID=$!
sleep 2

# Start Agent
echo "Starting Agent..."
rm -rf ~/.fluxforge
./agent > agent_conv.log 2>&1 &
AGENT_PID=$!
sleep 5

AGENT_ID=$(grep "Node ID:" agent_conv.log | awk '{print $NF}' | head -n 1)
echo "Agent ID: $AGENT_ID"

echo "---------------------------------------------------"
echo "Convergence Test: File Existence"

# 1. Ensure file does NOT exist initially
rm -f /tmp/fluxforge_test_file

# 2. Create State
echo "Creating converging state..."
STATE_RESP=$(curl -s -X POST http://localhost:8080/states \
    -d "{\"node_id\": \"$AGENT_ID\", \"check_cmd\": \"ls /tmp/fluxforge_test_file\", \"apply_cmd\": \"touch /tmp/fluxforge_test_file\", \"desired_exit_code\": 0}")
STATE_ID=$(echo $STATE_RESP | python3 -c "import sys, json; print(json.load(sys.stdin)['state_id'])")

# 3. Trigger Reconcile
echo "Triggering Reconcile..."
curl -s -X POST http://localhost:8080/states/$STATE_ID/reconcile

# 4. Poll for 'compliant' status
echo "Waiting for convergence..."
for i in {1..20}; do
    STATUS=$(curl -s http://localhost:8080/states/$STATE_ID | python3 -c "import sys, json; print(json.load(sys.stdin)['status'])")
    echo "Current Status: $STATUS"
    
    if [ "$STATUS" == "compliant" ]; then
        echo "PASS: State converged to compliant"
        if [ -f /tmp/fluxforge_test_file ]; then
            echo "PASS: File actually created"
            exit 0
        else
            echo "FAIL: Status compliant but file missing?"
            exit 1
        fi
    fi
    
    if [ "$STATUS" == "failed" ]; then
        echo "FAIL: State failed to converge"
        exit 1
    fi
    sleep 1
done

echo "FAIL: Timed out waiting for compliance"
exit 1
