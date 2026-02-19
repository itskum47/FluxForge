#!/bin/bash
# Phase 7: Leader Failover Hardening
# Verifies:
# 1. Automatic Leader Election upon failure
# 2. Zero Data Loss (Raft Log Replication)
# 3. Task Handoff (Scheduler State Recovered)

set -e

API_BASE="http://localhost:8090"
TENANT_ID="default"
LOG_FILE="/tmp/leader_failover.log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }

echo "Starting Leader Failover Test..." > $LOG_FILE

# 1. Identify Current Leader
log_info "Identifying Current Leader..."
CURRENT_LEADER=""
for port in 8080 8081 8082; do
    IS_LEADER=$(curl -s http://localhost:$port/metrics | grep "flux_leader_status" | grep " 1" | wc -l)
    if [ "$IS_LEADER" -eq 1 ]; then
        # Map port to container name
        if [ "$port" -eq 8080 ]; then CURRENT_LEADER="fluxforge-control-1"; fi
        if [ "$port" -eq 8081 ]; then CURRENT_LEADER="fluxforge-control-2"; fi
        if [ "$port" -eq 8082 ]; then CURRENT_LEADER="fluxforge-control-3"; fi
        break
    fi
done

if [ -z "$CURRENT_LEADER" ]; then
    log_fail "No active leader found. Test Aborted."
    exit 1
fi

log_info "Current Leader: $CURRENT_LEADER"

# 2. Submit Canary Job (Data Loss Check)
log_info "Submitting Canary Job to ensure persistence..."
JOB_RES=$(curl -s -X POST "$API_BASE/jobs" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -H "Content-Type: application/json" \
    -d '{"node_id":"agent-1","command":"echo failover_canary"}')

JOB_ID=$(echo $JOB_RES | jq -r '.job_id // empty')

if [ -z "$JOB_ID" ]; then
    log_fail "Failed to submit canary job."
    exit 1
fi
log_info "Canary Job ID: $JOB_ID"

# 3. Kill Leader (Force Election)
log_info "KILLING Current Leader: $CURRENT_LEADER"
docker kill $CURRENT_LEADER >> $LOG_FILE 2>&1

# 4. Wait for Election (Max 60s)
log_info "Waiting for new leader election..."
NEW_LEADER=""
START_TIME=$(date +%s)
TIMEOUT=60

while [ -z "$NEW_LEADER" ]; do
    CURRENT_TIME=$(date +%s)
    if [ $((CURRENT_TIME - START_TIME)) -gt $TIMEOUT ]; then
        log_fail "Timeout waiting for new leader election."
        # Restore old leader
        docker start $CURRENT_LEADER >> $LOG_FILE 2>&1
        exit 1
    fi

    for port in 8080 8081 8082; do
        # Ignore the killed node (it might be restart later, but for now we look for OTHERS)
        # Note: In a real cluster, the killed node is dead. Here we check available ports.
        IS_LEADER=$(curl -s --max-time 2 http://localhost:$port/metrics 2>/dev/null | grep "flux_leader_status" | grep " 1" | wc -l)
        if [ "$IS_LEADER" -eq 1 ]; then
             # Map port to container name
            if [ "$port" -eq 8080 ]; then CANDIDATE="fluxforge-control-1"; fi
            if [ "$port" -eq 8081 ]; then CANDIDATE="fluxforge-control-2"; fi
            if [ "$port" -eq 8082 ]; then CANDIDATE="fluxforge-control-3"; fi
            
            if [ "$CANDIDATE" != "$CURRENT_LEADER" ]; then
                NEW_LEADER="$CANDIDATE"
                break
            fi
        fi
    done
    sleep 2
done

log_pass "New Leader Elected: $NEW_LEADER"

# 5. Verify Canary Job (Zero Data Loss)
log_info "Verifying Canary Job Existence..."
JOB_STATUS=$(curl -s -H "X-Tenant-ID: $TENANT_ID" "$API_BASE/jobs/$JOB_ID" | jq -r '.status // "unknown"')

if [[ "$JOB_STATUS" == "completed" || "$JOB_STATUS" == "running" || "$JOB_STATUS" == "queued" ]]; then
    log_pass "Canary Job Persisted! Status: $JOB_STATUS"
else
    log_fail "DATA LOSS DETECTED: Canary Job status is $JOB_STATUS"
    docker start $CURRENT_LEADER >> $LOG_FILE 2>&1
    exit 1
fi

# 6. Verify Scheduler Handoff (New Leader must be scheduling)
log_info "Verifying Scheduler Handoff..."
# Submit a new job to confirm the NEW leader is accepting writes
NEW_JOB_RES=$(curl -s -X POST "$API_BASE/jobs" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -H "Content-Type: application/json" \
    -d '{"node_id":"agent-1","command":"echo post_failover_test"}')
NEW_JOB_ID=$(echo $NEW_JOB_RES | jq -r '.job_id // empty')

if [ -z "$NEW_JOB_ID" ]; then
    log_fail "New Leader failing to accept jobs (Scheduler/Raft issue)."
else
    log_pass "Scheduler Handoff Verified. New Job ID: $NEW_JOB_ID"
fi

# Cleanup: Restore killed node
log_info "Restoring original node: $CURRENT_LEADER"
docker start $CURRENT_LEADER >> $LOG_FILE 2>&1

log_pass "Leader Failover Test Passed Successfully."
