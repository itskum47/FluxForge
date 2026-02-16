#!/bin/bash
# Phase 7 Production Hardening - Chaos Monkey
# Randomly kills processes and injects faults to test resilience

set -e

DURATION_MINUTES=${1:-60}
CHAOS_INTERVAL_MIN=60
CHAOS_INTERVAL_MAX=300

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[CHAOS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }
log_event() { echo -e "${YELLOW}[EVENT]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"; }

CHAOS_LOG="/tmp/chaos_events.log"
echo "timestamp,event_type,target,result" > $CHAOS_LOG

log_info "=========================================="
log_info "Chaos Monkey Started"
log_info "Duration: $DURATION_MINUTES minutes"
log_info "=========================================="

END_TIME=$(($(date +%s) + DURATION_MINUTES * 60))

chaos_kill_control_plane() {
    # Kill random control plane node
    NODES=("fluxforge-control-1" "fluxforge-control-2" "fluxforge-control-3")
    TARGET=${NODES[$((RANDOM % 3))]}
    
    log_event "Killing control plane node: $TARGET"
    docker kill $TARGET >> $CHAOS_LOG 2>&1
    
    TIMESTAMP=$(date +%s)
    echo "$TIMESTAMP,kill_control_plane,$TARGET,killed" >> $CHAOS_LOG
    
    # Wait for recovery
    sleep 30
    
    # Verify new leader elected
    for port in 8080 8081 8082; do
        IS_LEADER=$(curl -s http://localhost:$port/api/dashboard 2>/dev/null | jq -r '.is_leader // false')
        if [ "$IS_LEADER" == "true" ]; then
            log_event "New leader elected on port $port"
            echo "$TIMESTAMP,leader_elected,port_$port,success" >> $CHAOS_LOG
            return 0
        fi
    done
    
    log_event "WARNING: No leader found after failover"
    echo "$TIMESTAMP,leader_elected,none,failure" >> $CHAOS_LOG
}

chaos_kill_agent() {
    # Kill random agent
    AGENT_COUNT=$(docker ps --filter "name=fluxforge-agent" --format "{{.Names}}" | wc -l)
    if [ "$AGENT_COUNT" -eq 0 ]; then
        log_event "No agents running, skipping"
        return
    fi
    
    TARGET=$(docker ps --filter "name=fluxforge-agent" --format "{{.Names}}" | shuf -n 1)
    log_event "Killing agent: $TARGET"
    docker kill $TARGET >> $CHAOS_LOG 2>&1
    
    TIMESTAMP=$(date +%s)
    echo "$TIMESTAMP,kill_agent,$TARGET,killed" >> $CHAOS_LOG
}

chaos_network_partition() {
    # Simulate network partition on random node
    NODES=("fluxforge-control-1" "fluxforge-control-2" "fluxforge-control-3")
    TARGET=${NODES[$((RANDOM % 3))]}
    
    log_event "Creating network partition on: $TARGET"
    
    # Disconnect from network
    docker network disconnect fluxforge-network $TARGET 2>/dev/null || true
    
    TIMESTAMP=$(date +%s)
    echo "$TIMESTAMP,network_partition,$TARGET,disconnected" >> $CHAOS_LOG
    
    # Wait 30 seconds
    sleep 30
    
    # Reconnect
    log_event "Healing network partition on: $TARGET"
    docker network connect fluxforge-network $TARGET 2>/dev/null || true
    echo "$TIMESTAMP,network_heal,$TARGET,reconnected" >> $CHAOS_LOG
}

chaos_database_restart() {
    log_event "Restarting PostgreSQL database"
    docker restart fluxforge-postgres >> $CHAOS_LOG 2>&1
    
    TIMESTAMP=$(date +%s)
    echo "$TIMESTAMP,database_restart,postgres,restarted" >> $CHAOS_LOG
    
    # Wait for database to be healthy
    sleep 15
    log_event "Database restart complete"
}

chaos_cpu_stress() {
    # Stress CPU on random control plane node
    NODES=("fluxforge-control-1" "fluxforge-control-2" "fluxforge-control-3")
    TARGET=${NODES[$((RANDOM % 3))]}
    
    log_event "Injecting CPU stress on: $TARGET"
    docker exec $TARGET sh -c "dd if=/dev/zero of=/dev/null &" 2>/dev/null || true
    
    TIMESTAMP=$(date +%s)
    echo "$TIMESTAMP,cpu_stress,$TARGET,injected" >> $CHAOS_LOG
    
    # Let it run for 30 seconds
    sleep 30
    
    # Kill stress process
    docker exec $TARGET pkill dd 2>/dev/null || true
    log_event "CPU stress removed from: $TARGET"
}

# Main chaos loop
while [ $(date +%s) -lt $END_TIME ]; do
    # Random sleep between chaos events
    SLEEP_TIME=$((RANDOM % (CHAOS_INTERVAL_MAX - CHAOS_INTERVAL_MIN) + CHAOS_INTERVAL_MIN))
    log_info "Next chaos event in $SLEEP_TIME seconds..."
    sleep $SLEEP_TIME
    
    # Select random chaos action
    ACTIONS=(
        "chaos_kill_control_plane"
        "chaos_kill_agent"
        "chaos_network_partition"
        "chaos_database_restart"
        "chaos_cpu_stress"
    )
    
    ACTION=${ACTIONS[$((RANDOM % ${#ACTIONS[@]}))]}
    log_event "Executing: $ACTION"
    $ACTION
    
    # Verify system still healthy
    HEALTHY=false
    for port in 8080 8081 8082; do
        HEALTH=$(curl -s http://localhost:$port/health 2>/dev/null)
        if [ "$HEALTH" == "ok" ]; then
            HEALTHY=true
            break
        fi
    done
    
    if [ "$HEALTHY" == "true" ]; then
        log_event "System still healthy after chaos"
    else
        log_event "WARNING: System unhealthy after chaos"
    fi
done

log_info "=========================================="
log_info "Chaos Monkey Stopped"
log_info "Events logged to: $CHAOS_LOG"
log_info "=========================================="

# Generate summary
TOTAL_EVENTS=$(tail -n +2 $CHAOS_LOG | wc -l)
KILL_EVENTS=$(grep "kill_control_plane" $CHAOS_LOG | wc -l)
PARTITION_EVENTS=$(grep "network_partition" $CHAOS_LOG | wc -l)
DB_RESTARTS=$(grep "database_restart" $CHAOS_LOG | wc -l)

log_info "Summary:"
log_info "  Total chaos events: $TOTAL_EVENTS"
log_info "  Control plane kills: $KILL_EVENTS"
log_info "  Network partitions: $PARTITION_EVENTS"
log_info "  Database restarts: $DB_RESTARTS"
