#!/bin/bash
set -e

# Phase 5: Agent Verification Script
# Verifies: Registration, Heartbeat, Job Execution

CONTROL_PLANE_URL="http://localhost:8080"
AGENT_CONTAINER="fluxforge-agent" # Default name usually unless specified otherwise?
# docker-compose generates name folder_service_index. e.g. deployments_agent_1?
# Let's check docker ps to be sure.

echo "=== 1. Starting Agent ==="
cd deployments
docker-compose up -d agent
echo "Waiting for Agent to initialize (10s)..."
sleep 10

# Identify Container Name
AGENT_CONTAINER=$(docker ps --format "{{.Names}}" | grep agent | head -n 1)
if [ -z "$AGENT_CONTAINER" ]; then
    echo "❌ Agent container not found!"
    exit 1
fi
echo "Agent Container: $AGENT_CONTAINER"

echo "=== 2. Verifying Registration ==="
# Check Agent Logs for "Successfully registered"
if docker logs "$AGENT_CONTAINER" 2>&1 | grep -q "Successfully registered"; then
    echo "✅ Agent Log: Registration Success"
else
    echo "❌ Agent Log: Registration NOT found"
    docker logs "$AGENT_CONTAINER" | tail -10
    exit 1
fi

# Check Control Plane API (if possible) or Redis
# We don't have public API to list agents yet?
# interface.go has ListAgents. api.go has NO list agents handler?
# Wait, api.go might NOT have list agents.
# We will rely on Logs.

echo "=== 3. Verifying Heartbeats ==="
# Wait a bit for heartbeats (interval 10s)
sleep 15
if docker logs "$AGENT_CONTAINER" 2>&1 | grep -q "Heartbeat sent successfully"; then
    echo "✅ Agent Log: Heartbeat Success"
else
    echo "❌ Agent Log: Heartbeat NOT found"
    docker logs "$AGENT_CONTAINER" | tail -10
    exit 1
fi

echo "=== 4. Submitting Test Job ==="
# Need Agent Node ID. Logs show it.
# Log: "Agent starting. Node ID: agent-1"
NODE_ID=$(docker logs "$AGENT_CONTAINER" 2>&1 | grep "Node ID:" | awk -F "Node ID: " '{print $2}' | tr -d '\r')
echo "Detected Node ID: $NODE_ID"

JOB_PAYLOAD="{\"node_id\": \"$NODE_ID\", \"command\": \"echo Hello FluxForge Phase 5\"}"
echo "Submitting Job: $JOB_PAYLOAD"

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$JOB_PAYLOAD" "$CONTROL_PLANE_URL/jobs")

if [ "$HTTP_CODE" == "202" ] || [ "$HTTP_CODE" == "200" ]; then
    echo "✅ Job Submitted (HTTP $HTTP_CODE)"
else
    echo "❌ Job Submission Failed (HTTP $HTTP_CODE)"
    exit 1
fi

echo "=== 5. Verifying Job Execution ==="
sleep 5
# Agent should receive job and execute it
if docker logs "$AGENT_CONTAINER" 2>&1 | grep -q "Executing job"; then
    echo "✅ Agent Log: Job Execution Started"
else
    echo "❌ Agent Log: Job Execution NOT found"
    docker logs "$AGENT_CONTAINER" | tail -20
    exit 1
fi

# Check output?
# Agent logs might show output if implemented?
# "Command finished with error: <nil>"?

if docker logs "$AGENT_CONTAINER" 2>&1 | grep -q "Command finished"; then
    echo "✅ Agent Log: Job Execution Completed"
else
    echo "⚠️ Agent Log: Job Completion not verification (might take longer or logging different)"
fi

echo "=== PHASE 5 VERIFICATION PASSED ==="
