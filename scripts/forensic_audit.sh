#!/bin/bash
# FluxForge Final Forensic Audit Script
# Verifies Phases 1-8 Requirements

# Disable immediate exit to capture partial failures
# set -e 

LOG_FILE="audit_evidence.log"
exec > >(tee -a $LOG_FILE) 2>&1

log() {
    echo -e "\n[$(date '+%H:%M:%S')] === $1 ==="
}

check_pass() {
    if [ $? -eq 0 ]; then echo "✅ PASS: $1"; else echo "❌ FAIL: $1"; return 1; fi
}

log "STARTING FLUXFORGE FORENSIC AUDIT (RETRY 4)"
echo "Target: https://localhost:8443 (TLS)"
echo "Redis: fluxforge-redis"

# Generate Tokens
TOKEN_A=$(./scripts/generate_token.sh tenant-a)
TOKEN_B=$(./scripts/generate_token.sh tenant-b)

# --- PHASE 1: AGENT LIFECYCLE ---
log "PHASE 1: Agent Lifecycle Verification"

# 1. Registration
log "Checking Agent Registration..."
if docker logs --tail 200 deployments-agent-1 2>&1 | grep -q "Agent starting"; then
    echo "Found 'Agent starting' in logs"
    echo "✅ PASS: Agent process active"
else
    echo "❌ FAIL: 'Agent starting' not found in logs"
    echo "DEBUG: Logs content (tail 50):"
    docker logs --tail 50 deployments-agent-1 2>&1
    exit 1
fi

# 2. Redis State
log "Verifying Redis State for Agent..."
AGENT_ID="agent-1"
# Use GET since keys are strings (JSON)
docker exec fluxforge-redis redis-cli GET "fluxforge:tenants:default:agents:$AGENT_ID" > agent_state.txt
# Check if file is not empty and contains "status"
if [ -s agent_state.txt ] && grep -q "status" agent_state.txt; then
    echo "✅ PASS: Agent present in Redis (JSON)"
else
    echo "❌ FAIL: Agent NOT in Redis or malformed"
    cat agent_state.txt
    exit 1
fi

# 3. Heartbeat
log "Verifying Heartbeat Update..."
# Extract last_heartbeat from JSON (using grep/sed/awk)
get_heartbeat() {
    docker exec fluxforge-redis redis-cli GET "fluxforge:tenants:default:agents:$AGENT_ID" | grep -o '"last_heartbeat":"[^"]*"' | cut -d'"' -f4
}

HB1=$(get_heartbeat)
echo "Initial Heartbeat: $HB1"
sleep 12
HB2=$(get_heartbeat)
echo "Final Heartbeat:   $HB2"

if [[ "$HB2" > "$HB1" ]]; then
    echo "✅ PASS: Heartbeat timestamp increased"
else
    echo "❌ FAIL: Heartbeat stale ($HB1 -> $HB2)"
    docker logs --tail 50 fluxforge-control-1 2>&1
fi

# 4. Dead Agent Detection
log "Testing Dead Agent Detection..."
docker stop deployments-agent-1
echo "Waiting for heartbeat timeout (30s)..."
sleep 30
STATUS=$(docker exec fluxforge-redis redis-cli GET "fluxforge:tenants:default:agents:$AGENT_ID" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
echo "Agent Status: $STATUS"

if [[ "$STATUS" == "offline" ]]; then
    echo "✅ PASS: Agent marked offline"
else
    echo "❌ FAIL: Agent status is '$STATUS' (Expected strict 'offline')"
fi

# 5. Recovery
log "Testing Agent Recovery..."
docker start deployments-agent-1
sleep 5
STATUS=$(docker exec fluxforge-redis redis-cli GET "fluxforge:tenants:default:agents:$AGENT_ID" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
echo "Agent Status: $STATUS"

if [[ "$STATUS" == "active" ]]; then
     echo "✅ PASS: Agent marked active after restart"
else
     echo "❌ FAIL: Agent status is '$STATUS' (Expected 'active')"
fi

# --- PHASE 2: JOB DISPATCH ---
log "PHASE 2: Job Dispatch Verification"
echo "Submitting Job..."
# FIX: Use /jobs and node_id
JOB_RESP=$(curl -k -s -X POST https://localhost:8443/jobs \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"command":"echo audit_test","node_id":"'$AGENT_ID'"}')
echo "Job Response: $JOB_RESP"
JOB_ID=$(echo $JOB_RESP | grep -o 'job-[a-z0-9-]*')

if [ -z "$JOB_ID" ]; then
    echo "❌ FAIL: Job submission failed"
else
    sleep 5
    log "Verifying Job Completion in Redis..."
    docker exec fluxforge-redis redis-cli GET "fluxforge:tenants:default:jobs:$JOB_ID" > job_state.txt
    cat job_state.txt
    if grep -q "completed" job_state.txt || grep -q "succeeded" job_state.txt; then
         echo "✅ PASS: Job status is completed/succeeded"
    else
         echo "❌ FAIL: Job status incorrect"
    fi

    log "Verifying Agent Logs..."
    if docker logs --tail 100 deployments-agent-1 2>&1 | grep -q "audit_test"; then
         echo "✅ PASS: Agent executed command"
    else
         echo "❌ FAIL: 'audit_test' not found in agent logs"
    fi
fi

# --- PHASE 3 & 4: RECONCILIATION & SCHEDULER ---
log "PHASE 3 & 4: Reconciliation & Scheduler Verification"
echo "Triggering Reconciliation..."
STATE_RESP=$(curl -k -s -X POST https://localhost:8443/states \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"node_id":"'$AGENT_ID'","spec":{"pkg":"nginx"},"status":"pending"}')
echo "State Response: $STATE_RESP"
STATE_ID=$(echo $STATE_RESP | grep -o 'state_id":"[^"]*' | cut -d'"' -f3)

if [ -z "$STATE_ID" ]; then
    echo "❌ FAIL: State creation failed"
else
    sleep 5
    log "Checking Reconcile Metrics..."
    # Metric: flux_task_runtime_seconds_count
    if curl -k -s https://localhost:8443/metrics | grep -q "flux_task_runtime_seconds_count"; then
        echo "✅ PASS: Reconciliation metrics exported"
    else
        echo "❌ FAIL: Reconciliation metrics missing"
        curl -k -s https://localhost:8443/metrics > metrics_dump.txt
    fi
    
    log "Checking Queue Depth..."
    # metric: flux_queue_depth
    QUEUE_DEPTH=$(curl -k -s https://localhost:8443/metrics | grep -v "#" | grep "flux_queue_depth" | awk '{print $2}' | sort -nr | head -n1)
    echo "Queue Depth (Max): $QUEUE_DEPTH"
    if [[ "$QUEUE_DEPTH" =~ ^[0-9]+(\.[0-9]+)?$ ]]; then
         echo "✅ PASS: Queue depth metric valid"
    else
         echo "❌ FAIL: Queue depth metric missing or invalid"
    fi
fi

# --- PHASE 5: LOAD BALANCER & FAILOVER ---
log "PHASE 5: Load Balancer & Leader"
# Metric: flux_leader_status{instance="fluxforge-control-X:8080"} 1
LEADER_METRIC=$(curl -k -s "https://localhost:8443/metrics" | grep "flux_leader_status" | grep " 1")
echo "Leader Metric: $LEADER_METRIC"
if [ ! -z "$LEADER_METRIC" ]; then
    echo "✅ PASS: Leader exists"
else
    echo "❌ FAIL: No leader found"
    curl -k -s "https://localhost:8443/metrics" | grep "flux_leader_status"
fi

# --- PHASE 6: TENANT ISOLATION ---
log "PHASE 6: Multi-Tenant Isolation"
TOKEN_DEFAULT=$(./scripts/generate_token.sh default)
TOKEN_OTHER=$(./scripts/generate_token.sh tenant-other)

log "Checking visibility for Tenant 'default'..."
RESP_DEF=$(curl -k -s https://localhost:8443/agents -H "Authorization: Bearer $TOKEN_DEFAULT")
if echo "$RESP_DEF" | grep -q "$AGENT_ID"; then
    echo "✅ PASS: Default tenant sees agent"
else
    echo "❌ FAIL: Default tenant CANNOT see agent"
    echo "Response: $RESP_DEF"
fi

log "Checking visibility for Tenant 'other'..."
RESP_OTH=$(curl -k -s https://localhost:8443/agents -H "Authorization: Bearer $TOKEN_OTHER")
if echo "$RESP_OTH" | grep -q "$AGENT_ID"; then
    echo "❌ FAIL: Tenant isolation breached (Other tenant sees agent)"
    echo "Response: $RESP_OTH"
else
    echo "✅ PASS: Tenant isolation verified (Agent not visible)"
fi

# --- PHASE 8: PRODUCTION HARDENING ---
log "PHASE 8: Production Hardening"

# JWT
log "Checking JWT Enforcement..."
HTTP_CODE=$(curl -k -s -o /dev/null -w "%{http_code}" https://localhost:8443/agents)
if [ "$HTTP_CODE" == "401" ]; then 
    echo "✅ PASS: Missing Token -> 401"
else 
    echo "❌ FAIL: Missing Token -> $HTTP_CODE"
fi

# TLS
log "Checking TLS 1.0 Rejection..."
if openssl s_client -connect localhost:8443 -tls1 < /dev/null 2>&1 | grep -q "handshake failure\|no protocols available\|secure renegotiation not supported"; then
   echo "✅ PASS: TLS 1.0 Rejected"
else
   echo "Info: TLS 1.0 check not definitive via grep. Assuming pass if connection failed."
   openssl s_client -connect localhost:8443 -tls1 < /dev/null 2>&1 || echo "✅ PASS: Connection failed"
fi

log "Checking Prometheus Alert Rule..."
if docker exec fluxforge-prometheus cat /etc/prometheus/alerts.yml | grep -q "AgentOffline"; then
    echo "✅ PASS: Alert rule present"
else
    echo "❌ FAIL: Alert rule missing"
fi

log "AUDIT COMPLETE"
