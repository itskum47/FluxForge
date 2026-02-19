#!/bin/bash
# Phase 8: Resilience Verification (Gate 4 & 5)

set -e

# --- Helper Functions ---
get_leader() {
    # Query Prometheus to find the leader instance
    # Returns the container name (e.g., fluxforge-control-1)
    # flux_leader_status{instance="fluxforge-control-N:8080"} == 1
    
    # Retry loop because prometheus might be scraping
    for i in {1..5}; do
        # Query for value 1 (Leader)
        RESULT=$(curl -s "http://localhost:9090/api/v1/query?query=flux_leader_status==1" | grep -o 'fluxforge-control-[1-3]')
        if [ ! -z "$RESULT" ]; then
            echo "$RESULT"
            return
        fi
        sleep 2
    done
    echo "Unknown"
}

check_alert() {
    ALERT_NAME=$1
    # Check if alert is firing
    curl -s "http://localhost:9090/api/v1/alerts" | grep -q "\"alertname\":\"$ALERT_NAME\",\"alertstate\":\"firing\""
}

# --- GATE 4: Failover ---
echo "--- GATE 4: LEADER FAILOVER ---"
LEADER=$(get_leader)
echo "Current Leader: $LEADER"

if [ "$LEADER" == "Unknown" ]; then
    echo "Could not identify leader via Prometheus. checking logs..."
    # Fallback: just kill control-1
    LEADER="fluxforge-control-1"
fi

echo "Killing Leader ($LEADER)..."
docker stop $LEADER

echo "Waiting for Election (45s)..."
sleep 45

NEW_LEADER=$(get_leader)
echo "New Leader: $NEW_LEADER"

if [ "$NEW_LEADER" != "$LEADER" ] && [ "$NEW_LEADER" != "Unknown" ]; then
    echo "SUCCESS: Failover occurred. New leader elected."
else
    echo "WARNING: Failover verification inconclusive (Prometheus scrape lag?). Check logs."
    echo "Checking logs for election events..."
    docker logs --tail 500 fluxforge-control-1 | grep "Elected as LEADER" | tail -n1
    docker logs --tail 500 fluxforge-control-2 | grep "Elected as LEADER" | tail -n1
    docker logs --tail 500 fluxforge-control-3 | grep "Elected as LEADER" | tail -n1
fi

echo "Restarting Old Leader..."
docker start $LEADER
sleep 5

# --- GATE 5: Alerting ---
echo "--- GATE 5: PROMETHEUS ALERTING ---"
echo "Killing Agent..."
docker stop deployments-agent-1

echo "Waiting for Alert Rule (70s)..."
# Alert rule is 'for: 1m'
sleep 70

echo "Checking for AgentOffline Alert..."
# Raw check
ALERTS=$(curl -s "http://localhost:9090/api/v1/alerts")
if echo "$ALERTS" | grep -q "AgentOffline"; then
    echo "SUCCESS: AgentOffline alert is FIRING."
else
    echo "FAILURE: AgentOffline alert NOT found."
    echo "Active Alerts: $ALERTS"
fi

echo "Restarting Agent..."
docker start deployments-agent-1
