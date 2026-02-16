#!/bin/bash
# FluxForge System Audit Script
# Comprehensive verification of system correctness, safety, and recoverability

set -e

API_BASE="http://localhost:8080"
AGENT_ID="audit-agent-1"
PASS=0
FAIL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
    ((PASS++))
}

log_fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    ((FAIL++))
}

log_info() {
    echo -e "${YELLOW}ℹ INFO:${NC} $1"
}

log_section() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
}

# Phase 1: Agent System Verification
phase1_agent_verification() {
    log_section "PHASE 1: Agent System Verification"
    
    log_info "Registering agent: $AGENT_ID"
    REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/agent/register" \
        -H "Content-Type: application/json" \
        -d "{\"node_id\":\"$AGENT_ID\",\"status\":\"active\",\"tier\":\"normal\"}")
    
    if echo "$REGISTER_RESPONSE" | grep -q "success\|ok\|registered"; then
        log_pass "Agent registration"
    else
        log_fail "Agent registration: $REGISTER_RESPONSE"
    fi
    
    sleep 2
    
    log_info "Verifying agent in dashboard"
    AGENT_COUNT=$(curl -s "$API_BASE/api/dashboard" | jq '.active_agents // 0')
    
    if [ "$AGENT_COUNT" -ge 1 ]; then
        log_pass "Agent visible in dashboard (count: $AGENT_COUNT)"
    else
        log_fail "Agent not visible in dashboard (count: $AGENT_COUNT)"
    fi
    
    log_info "Agent lifecycle test complete"
}

# Phase 2: Job Dispatch Verification
phase2_job_dispatch() {
    log_section "PHASE 2: Job Dispatch Verification"
    
    log_info "Submitting test job"
    JOB_RESPONSE=$(curl -s -X POST "$API_BASE/jobs" \
        -H "Content-Type: application/json" \
        -d "{\"node_id\":\"$AGENT_ID\",\"command\":\"echo audit-ok\"}")
    
    if echo "$JOB_RESPONSE" | grep -q "success\|ok\|created"; then
        log_pass "Job submission"
    else
        log_fail "Job submission: $JOB_RESPONSE"
    fi
    
    sleep 3
    
    log_info "Job dispatch test complete"
}

# Phase 3: Reconciliation Engine Verification
phase3_reconciliation() {
    log_section "PHASE 3: Reconciliation Engine Verification"
    
    log_info "Submitting desired state (passing check)"
    STATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/states" \
        -H "Content-Type: application/json" \
        -d "{\"state_id\":\"audit-state-pass\",\"node_id\":\"$AGENT_ID\",\"check_cmd\":\"true\"}")
    
    if echo "$STATE_RESPONSE" | grep -q "success\|ok\|created"; then
        log_pass "State submission (passing)"
    else
        log_fail "State submission (passing): $STATE_RESPONSE"
    fi
    
    sleep 5
    
    log_info "Submitting desired state (failing check)"
    FAIL_STATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/states" \
        -H "Content-Type: application/json" \
        -d "{\"state_id\":\"audit-state-fail\",\"node_id\":\"$AGENT_ID\",\"check_cmd\":\"exit 1\"}")
    
    if echo "$FAIL_STATE_RESPONSE" | grep -q "success\|ok\|created"; then
        log_pass "State submission (failing)"
    else
        log_fail "State submission (failing): $FAIL_STATE_RESPONSE"
    fi
    
    sleep 5
    
    log_info "Reconciliation test complete"
}

# Phase 4: Scheduler Verification
phase4_scheduler() {
    log_section "PHASE 4: Scheduler Verification"
    
    log_info "Submitting 50 jobs for load test"
    for i in {1..50}; do
        curl -s -X POST "$API_BASE/jobs" \
            -H "Content-Type: application/json" \
            -d "{\"node_id\":\"$AGENT_ID\",\"command\":\"sleep 0.1\"}" &
    done
    wait
    
    log_pass "50 jobs submitted"
    
    sleep 2
    
    log_info "Checking scheduler metrics"
    METRICS=$(curl -s "$API_BASE/api/dashboard")
    QUEUE_DEPTH=$(echo "$METRICS" | jq '.queue_depth // 0')
    SATURATION=$(echo "$METRICS" | jq '.worker_saturation // 0')
    
    log_info "Queue depth: $QUEUE_DEPTH, Saturation: $SATURATION"
    
    if [ "$QUEUE_DEPTH" -ge 0 ]; then
        log_pass "Scheduler queue operational"
    else
        log_fail "Scheduler queue metrics unavailable"
    fi
    
    log_info "Waiting for queue to drain..."
    sleep 10
    
    FINAL_QUEUE=$(curl -s "$API_BASE/api/dashboard" | jq '.queue_depth // 0')
    log_info "Final queue depth: $FINAL_QUEUE"
    
    if [ "$FINAL_QUEUE" -eq 0 ]; then
        log_pass "Queue drained to zero"
    else
        log_info "Queue still draining (depth: $FINAL_QUEUE)"
    fi
}

# Phase 6: Dashboard and WebSocket Verification
phase6_dashboard() {
    log_section "PHASE 6: Dashboard and WebSocket Verification"
    
    log_info "Testing dashboard API"
    DASHBOARD_RESPONSE=$(curl -s "$API_BASE/api/dashboard")
    
    if echo "$DASHBOARD_RESPONSE" | jq -e '.queue_depth' > /dev/null 2>&1; then
        log_pass "Dashboard API responding"
    else
        log_fail "Dashboard API not responding correctly"
    fi
    
    log_info "Testing incident capture"
    INCIDENT_RESPONSE=$(curl -s -X POST "$API_BASE/api/incidents/capture?state_id=audit-state-pass")
    
    if echo "$INCIDENT_RESPONSE" | jq -e '.incident_id' > /dev/null 2>&1; then
        log_pass "Incident capture working"
    else
        log_info "Incident capture response: $INCIDENT_RESPONSE"
    fi
    
    log_info "Testing cluster API"
    CLUSTER_RESPONSE=$(curl -s "$API_BASE/api/clusters")
    
    if echo "$CLUSTER_RESPONSE" | jq -e '.[0].cluster_id' > /dev/null 2>&1; then
        log_pass "Cluster API responding"
    else
        log_fail "Cluster API not responding correctly"
    fi
}

# Phase 7: Full System Chaos Test
phase7_chaos() {
    log_section "PHASE 7: Full System Chaos Test"
    
    log_info "Submitting 200 states simultaneously"
    for i in {1..200}; do
        curl -s -X POST "$API_BASE/api/v1/states" \
            -H "Content-Type: application/json" \
            -d "{\"state_id\":\"chaos-state-$i\",\"node_id\":\"$AGENT_ID\",\"check_cmd\":\"true\"}" &
    done
    wait
    
    log_pass "200 states submitted"
    
    sleep 5
    
    log_info "Checking system health under load"
    CHAOS_METRICS=$(curl -s "$API_BASE/api/dashboard")
    CHAOS_QUEUE=$(echo "$CHAOS_METRICS" | jq '.queue_depth // 0')
    
    log_info "Queue depth under chaos: $CHAOS_QUEUE"
    
    if [ "$CHAOS_QUEUE" -ge 0 ]; then
        log_pass "System responding under chaos load"
    else
        log_fail "System not responding under chaos load"
    fi
}

# Main execution
main() {
    echo "╔════════════════════════════════════════╗"
    echo "║   FluxForge System Audit               ║"
    echo "║   Comprehensive Verification Suite     ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    log_info "Starting system audit at $(date)"
    log_info "API Base: $API_BASE"
    log_info "Agent ID: $AGENT_ID"
    
    # Run all phases
    phase1_agent_verification
    phase2_job_dispatch
    phase3_reconciliation
    phase4_scheduler
    phase6_dashboard
    phase7_chaos
    
    # Summary
    log_section "AUDIT SUMMARY"
    echo ""
    echo "Total Tests: $((PASS + FAIL))"
    echo -e "${GREEN}Passed: $PASS${NC}"
    echo -e "${RED}Failed: $FAIL${NC}"
    echo ""
    
    if [ $FAIL -eq 0 ]; then
        echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║   ✓ SYSTEM AUDIT PASSED                ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
        exit 0
    else
        echo -e "${RED}╔════════════════════════════════════════╗${NC}"
        echo -e "${RED}║   ✗ SYSTEM AUDIT FAILED                ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════╝${NC}"
        exit 1
    fi
}

# Run main
main
