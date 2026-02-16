#!/bin/bash
# Phase 7 Production Hardening - Multi-Node Deployment Test
# This script validates multi-node control plane deployment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

PASS_COUNT=0
FAIL_COUNT=0

# Test Section 1: Multi-Node Deployment
test_multi_node_deployment() {
    log_info "=========================================="
    log_info "TEST 1: Multi-Node Control Plane Deployment"
    log_info "=========================================="
    
    # Start Docker Compose cluster
    log_info "Starting 3-node control plane cluster..."
    cd "$PROJECT_ROOT/deployments"
    docker-compose up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to become healthy..."
    sleep 30
    
    # Verify all containers running
    RUNNING_COUNT=$(docker-compose ps --services --filter "status=running" | wc -l)
    if [ "$RUNNING_COUNT" -ge 6 ]; then
        log_pass "All containers running (count: $RUNNING_COUNT)"
        ((PASS_COUNT++))
    else
        log_fail "Not all containers running (count: $RUNNING_COUNT)"
        ((FAIL_COUNT++))
    fi
    
    # Verify leader election
    log_info "Verifying leader election..."
    sleep 10
    
    LEADER_COUNT=0
    for port in 8080 8081 8082; do
        IS_LEADER=$(curl -s http://localhost:$port/api/dashboard 2>/dev/null | jq -r '.is_leader // false')
        if [ "$IS_LEADER" == "true" ]; then
            ((LEADER_COUNT++))
            log_info "Node on port $port is leader"
        fi
    done
    
    if [ "$LEADER_COUNT" -eq 1 ]; then
        log_pass "Exactly one leader elected"
        ((PASS_COUNT++))
    else
        log_fail "Invalid leader count: $LEADER_COUNT (expected 1)"
        ((FAIL_COUNT++))
    fi
    
    # Measure leader election time
    log_info "Testing leader election timing..."
    docker-compose restart control-1 control-2 control-3
    
    START_TIME=$(date +%s)
    LEADER_FOUND=false
    
    for i in {1..60}; do
        for port in 8080 8081 8082; do
            IS_LEADER=$(curl -s http://localhost:$port/api/dashboard 2>/dev/null | jq -r '.is_leader // false')
            if [ "$IS_LEADER" == "true" ]; then
                END_TIME=$(date +%s)
                ELECTION_TIME=$((END_TIME - START_TIME))
                log_pass "Leader elected in ${ELECTION_TIME} seconds"
                LEADER_FOUND=true
                break 2
            fi
        done
        sleep 1
    done
    
    if [ "$LEADER_FOUND" == "true" ] && [ "$ELECTION_TIME" -lt 30 ]; then
        log_pass "Leader election time < 30s"
        ((PASS_COUNT++))
    else
        log_fail "Leader election too slow or failed"
        ((FAIL_COUNT++))
    fi
}

# Test Section 2: Failover Testing
test_failover() {
    log_info "=========================================="
    log_info "TEST 2: Leader Failover"
    log_info "=========================================="
    
    # Find current leader
    LEADER_PORT=""
    for port in 8080 8081 8082; do
        IS_LEADER=$(curl -s http://localhost:$port/api/dashboard 2>/dev/null | jq -r '.is_leader // false')
        if [ "$IS_LEADER" == "true" ]; then
            LEADER_PORT=$port
            break
        fi
    done
    
    if [ -z "$LEADER_PORT" ]; then
        log_fail "No leader found"
        ((FAIL_COUNT++))
        return
    fi
    
    log_info "Current leader on port $LEADER_PORT"
    
    # Kill leader
    LEADER_CONTAINER=""
    case $LEADER_PORT in
        8080) LEADER_CONTAINER="fluxforge-control-1" ;;
        8081) LEADER_CONTAINER="fluxforge-control-2" ;;
        8082) LEADER_CONTAINER="fluxforge-control-3" ;;
    esac
    
    log_info "Killing leader container: $LEADER_CONTAINER"
    docker kill $LEADER_CONTAINER
    
    # Measure failover time
    START_TIME=$(date +%s)
    NEW_LEADER_FOUND=false
    
    for i in {1..60}; do
        for port in 8080 8081 8082; do
            if [ "$port" == "$LEADER_PORT" ]; then
                continue
            fi
            
            IS_LEADER=$(curl -s http://localhost:$port/api/dashboard 2>/dev/null | jq -r '.is_leader // false')
            if [ "$IS_LEADER" == "true" ]; then
                END_TIME=$(date +%s)
                FAILOVER_TIME=$((END_TIME - START_TIME))
                log_pass "New leader elected on port $port in ${FAILOVER_TIME} seconds"
                NEW_LEADER_FOUND=true
                break 2
            fi
        done
        sleep 1
    done
    
    if [ "$NEW_LEADER_FOUND" == "true" ] && [ "$FAILOVER_TIME" -lt 30 ]; then
        log_pass "Failover time < 30s"
        ((PASS_COUNT++))
    else
        log_fail "Failover too slow or failed"
        ((FAIL_COUNT++))
    fi
    
    # Restart killed container
    docker-compose up -d $LEADER_CONTAINER
    sleep 10
}

# Test Section 3: Rolling Restart
test_rolling_restart() {
    log_info "=========================================="
    log_info "TEST 3: Rolling Restart Validation"
    log_info "=========================================="
    
    # Start continuous job submission
    log_info "Starting continuous job submission..."
    (
        for i in {1..100}; do
            curl -s -X POST http://localhost:8080/jobs \
                -H "Content-Type: application/json" \
                -d '{"node_id":"agent-1","command":"echo test"}' \
                >> /tmp/job_submissions.log 2>&1
            sleep 0.5
        done
    ) &
    JOB_SUBMITTER_PID=$!
    
    sleep 2
    
    # Rolling restart
    for container in fluxforge-control-2 fluxforge-control-3 fluxforge-control-1; do
        log_info "Restarting $container..."
        docker-compose restart $container
        sleep 15
        
        # Verify cluster still has leader
        LEADER_EXISTS=false
        for port in 8080 8081 8082; do
            IS_LEADER=$(curl -s http://localhost:$port/api/dashboard 2>/dev/null | jq -r '.is_leader // false')
            if [ "$IS_LEADER" == "true" ]; then
                LEADER_EXISTS=true
                break
            fi
        done
        
        if [ "$LEADER_EXISTS" == "true" ]; then
            log_pass "$container restarted, leader still exists"
            ((PASS_COUNT++))
        else
            log_fail "$container restarted, no leader found"
            ((FAIL_COUNT++))
        fi
    done
    
    # Wait for job submitter to finish
    wait $JOB_SUBMITTER_PID
    
    log_pass "Rolling restart complete"
}

# Test Section 4: Container Crash Recovery
test_container_crash_recovery() {
    log_info "=========================================="
    log_info "TEST 4: Container Crash Recovery"
    log_info "=========================================="
    
    # Kill random container
    RANDOM_CONTAINER="fluxforge-control-$((RANDOM % 3 + 1))"
    log_info "Killing $RANDOM_CONTAINER..."
    docker kill $RANDOM_CONTAINER
    
    # Wait for auto-restart
    sleep 10
    
    # Verify container restarted
    STATUS=$(docker inspect -f '{{.State.Status}}' $RANDOM_CONTAINER)
    if [ "$STATUS" == "running" ]; then
        log_pass "Container auto-restarted"
        ((PASS_COUNT++))
    else
        log_fail "Container did not auto-restart (status: $STATUS)"
        ((FAIL_COUNT++))
    fi
    
    # Verify cluster health
    sleep 10
    HEALTHY_COUNT=0
    for port in 8080 8081 8082; do
        HEALTH=$(curl -s http://localhost:$port/health 2>/dev/null)
        if [ "$HEALTH" == "ok" ]; then
            ((HEALTHY_COUNT++))
        fi
    done
    
    if [ "$HEALTHY_COUNT" -eq 3 ]; then
        log_pass "All nodes healthy after crash recovery"
        ((PASS_COUNT++))
    else
        log_fail "Not all nodes healthy (count: $HEALTHY_COUNT)"
        ((FAIL_COUNT++))
    fi
}

# Test Section 5: Database Connection Resilience
test_database_resilience() {
    log_info "=========================================="
    log_info "TEST 5: Database Connection Resilience"
    log_info "=========================================="
    
    # Restart database
    log_info "Restarting PostgreSQL..."
    docker-compose restart postgres
    
    # Wait for database to be healthy
    sleep 15
    
    # Verify control plane reconnects
    RECONNECTED_COUNT=0
    for port in 8080 8081 8082; do
        HEALTH=$(curl -s http://localhost:$port/health 2>/dev/null)
        if [ "$HEALTH" == "ok" ]; then
            ((RECONNECTED_COUNT++))
        fi
    done
    
    if [ "$RECONNECTED_COUNT" -eq 3 ]; then
        log_pass "All nodes reconnected to database"
        ((PASS_COUNT++))
    else
        log_fail "Not all nodes reconnected (count: $RECONNECTED_COUNT)"
        ((FAIL_COUNT++))
    fi
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    cd "$PROJECT_ROOT/deployments"
    docker-compose down -v
    rm -f /tmp/job_submissions.log
}

# Main execution
main() {
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║  FluxForge Phase 7: Production Hardening Tests        ║"
    echo "║  Multi-Node Deployment Validation                     ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""
    
    # Run tests
    test_multi_node_deployment
    test_failover
    test_rolling_restart
    test_container_crash_recovery
    test_database_resilience
    
    # Summary
    echo ""
    log_info "=========================================="
    log_info "TEST SUMMARY"
    log_info "=========================================="
    echo "Total Tests: $((PASS_COUNT + FAIL_COUNT))"
    echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
    echo -e "${RED}Failed: $FAIL_COUNT${NC}"
    echo ""
    
    # Cleanup
    cleanup
    
    if [ $FAIL_COUNT -eq 0 ]; then
        echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║   ✓ ALL TESTS PASSED                  ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
        exit 0
    else
        echo -e "${RED}╔════════════════════════════════════════╗${NC}"
        echo -e "${RED}║   ✗ SOME TESTS FAILED                 ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════╝${NC}"
        exit 1
    fi
}

# Run main
main
