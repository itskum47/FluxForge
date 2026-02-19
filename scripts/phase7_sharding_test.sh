#!/bin/bash
# Phase 7: Sharding & Load Distribution Verification
# Verifies:
# 1. Write Redirection: Submitting specific write requests to a FOLLOWER node should succeed (proxied to LEADER).
# 2. Read Consistency: Writing to Node A, Reading from Node B should show the data (Eventual Consistency / Raft Replication).
# 3. Round-Robin Distribution: Submitting jobs to all nodes sequentially should work.

set -e

TENANT_ID="default"
LOG_FILE="/tmp/sharding_test.log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }

echo "Starting Sharding Verification..." > $LOG_FILE

# 1. Identify Roles
log_info "Identifying Cluster Roles..."
LEADER_PORT=""
FOLLOWER_PORT=""

for port in 8080 8081 8082; do
    IS_LEADER=$(curl -s http://localhost:$port/metrics | grep "flux_leader_status" | grep " 1" | wc -l)
    if [ "$IS_LEADER" -eq 1 ]; then
        LEADER_PORT=$port
    else
        FOLLOWER_PORT=$port
    fi
done

if [ -z "$LEADER_PORT" ] || [ -z "$FOLLOWER_PORT" ]; then
    log_fail "Could not identify Leader and Follower. Leader: $LEADER_PORT, Follower: $FOLLOWER_PORT"
    exit 1
fi

log_info "Leader: $LEADER_PORT, Follower: $FOLLOWER_PORT"

# 2. Test Write Redirection (Submit to Follower)
log_info "Test 1: Write Redirection (Submit Job to Follower $FOLLOWER_PORT)..."
RES=$(curl -s -X POST "http://localhost:$FOLLOWER_PORT/jobs" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -H "Content-Type: application/json" \
    -d '{"node_id":"agent-1","command":"echo follower_submission"}')

JOB_ID=$(echo $RES | jq -r '.job_id // empty')

if [ -z "$JOB_ID" ]; then
    log_fail "Submission to Follower failed. Internal redirection broken?"
    echo "Response: $RES"
    exit 1
fi
log_pass "Job Accepted by Follower (Proxied). Job ID: $JOB_ID"

# 3. Test Read Consistency (Read from Leader)
log_info "Test 2: Read Consistency (Read Job from Leader $LEADER_PORT)..."
# Give a small moment for replication if needed, though usually strict consistency implies immediate or proxied read
sleep 1
STATUS=$(curl -s -H "X-Tenant-ID: $TENANT_ID" "http://localhost:$LEADER_PORT/jobs/$JOB_ID" | jq -r '.status // "unknown"')

if [[ "$STATUS" == "completed" || "$STATUS" == "running" || "$STATUS" == "queued" ]]; then
    log_pass "Job visible on Leader. Status: $STATUS"
else
    log_fail "Job NOT visible on Leader. Status: $STATUS"
    exit 1
fi

# 4. Test Round-Robin Submission
log_info "Test 3: Round-Robin Submission (All Nodes)..."
SUCCESS_COUNT=0
NODES=(8080 8081 8082)
for i in 0 1 2; do
    NODE=${NODES[$i]}
    log_info "Submitting to Node $NODE..."
    RES=$(curl -s -X POST "http://localhost:$NODE/jobs" \
        -H "X-Tenant-ID: $TENANT_ID" \
        -H "Content-Type: application/json" \
        -d '{"node_id":"agent-1","command":"echo rr_test_'$i'"}')
    
    ID=$(echo $RES | jq -r '.job_id // empty')
    if [ ! -z "$ID" ]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    fi
done

if [ "$SUCCESS_COUNT" -eq 3 ]; then
    log_pass "Round-Robin Submission Successful (3/3)."
else
    log_fail "Round-Robin Submission Failed. Success: $SUCCESS_COUNT/3"
    exit 1
fi

log_pass "Sharding & Load Distribution Verification Passed."
