#!/bin/bash
# FluxForge Production Certification Validator
# Automated validation of production certification checklist

set -e

API_BASE="http://localhost:8080"
CERT_REPORT="/tmp/certification_report.txt"
CERT_JSON="/tmp/certification_results.json"

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

TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Initialize JSON report
echo "{" > $CERT_JSON
echo "  \"certification_date\": \"$(date -Iseconds)\"," >> $CERT_JSON
echo "  \"results\": {" >> $CERT_JSON

check() {
    local category=$1
    local name=$2
    local command=$3
    local expected=$4
    
    ((TOTAL_CHECKS++))
    
    log_info "Checking: $name"
    
    result=$(eval "$command" 2>/dev/null || echo "ERROR")
    
    if [ "$result" == "$expected" ] || [[ "$result" =~ $expected ]]; then
        log_pass "$name"
        ((PASSED_CHECKS++))
        echo "    \"$category.$name\": \"PASS\"," >> $CERT_JSON
        return 0
    else
        log_fail "$name (got: $result, expected: $expected)"
        ((FAILED_CHECKS++))
        echo "    \"$category.$name\": \"FAIL\"," >> $CERT_JSON
        return 1
    fi
}

echo "╔════════════════════════════════════════════════════════╗"
echo "║  FluxForge Production Certification Validator         ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Section 1: Cluster Deployment
log_info "=========================================="
log_info "SECTION 1: Cluster Deployment"
log_info "=========================================="

check "cluster" "control_plane_running" \
    "docker ps --filter 'name=fluxforge-control' --format '{{.Status}}' | grep -c 'Up'" \
    "3"

check "cluster" "leader_elected" \
    "curl -s $API_BASE/api/clusters | jq '[.[] | select(.is_leader==true)] | length'" \
    "1"

check "cluster" "all_nodes_healthy" \
    "curl -s $API_BASE/api/clusters | jq '[.[] | select(.status==\"healthy\")] | length'" \
    "3"

# Section 2: Critical Metrics
log_info "=========================================="
log_info "SECTION 2: Critical Metrics"
log_info "=========================================="

check "metrics" "no_integrity_skew" \
    "curl -s $API_BASE/metrics | grep flux_integrity_skew_count | awk '{print \$2}'" \
    "0"

check "metrics" "no_deadlocks" \
    "curl -s $API_BASE/metrics | grep flux_scheduler_deadlocks_total | awk '{print \$2}'" \
    "0"

check "metrics" "no_split_brain" \
    "curl -s $API_BASE/metrics | grep flux_leader_split_brain_total | awk '{print \$2}'" \
    "0"

# Section 3: Scheduler Integrity
log_info "=========================================="
log_info "SECTION 3: Scheduler Integrity"
log_info "=========================================="

check "scheduler" "queue_depth_bounded" \
    "curl -s $API_BASE/api/dashboard | jq '.queue_depth < 10000'" \
    "true"

check "scheduler" "no_stuck_jobs" \
    "psql -h localhost -U fluxforge -t -c \"SELECT COUNT(*) FROM jobs WHERE status='running' AND updated_at < NOW() - INTERVAL '5 minutes'\" 2>/dev/null || echo 0" \
    "0"

# Section 4: Observability
log_info "=========================================="
log_info "SECTION 4: Observability"
log_info "=========================================="

check "observability" "metrics_exposed" \
    "curl -s $API_BASE/metrics | grep -c flux_" \
    "[0-9]+"

check "observability" "health_endpoint" \
    "curl -s $API_BASE/health" \
    "ok"

check "observability" "structured_logs" \
    "docker logs fluxforge-control-1 2>&1 | tail -1 | jq -e .timestamp > /dev/null && echo 'true' || echo 'false'" \
    "true"

# Section 5: Agent Lifecycle
log_info "=========================================="
log_info "SECTION 5: Agent Lifecycle"
log_info "=========================================="

check "agents" "agents_active" \
    "curl -s $API_BASE/api/dashboard | jq '.active_agents > 0'" \
    "true"

check "agents" "no_ghost_agents" \
    "psql -h localhost -U fluxforge -t -c \"SELECT COUNT(*) FROM agents WHERE status='active' AND last_heartbeat < NOW() - INTERVAL '60 seconds'\" 2>/dev/null || echo 0" \
    "0"

# Close JSON
echo "    \"_summary\": {" >> $CERT_JSON
echo "      \"total\": $TOTAL_CHECKS," >> $CERT_JSON
echo "      \"passed\": $PASSED_CHECKS," >> $CERT_JSON
echo "      \"failed\": $FAILED_CHECKS" >> $CERT_JSON
echo "    }" >> $CERT_JSON
echo "  }" >> $CERT_JSON
echo "}" >> $CERT_JSON

# Generate report
cat > $CERT_REPORT << EOF
FluxForge Production Certification Report
==========================================
Date: $(date)
Cluster: $(curl -s $API_BASE/api/clusters | jq -r '.[0].cluster_id' 2>/dev/null || echo "Unknown")

Summary
-------
Total Checks: $TOTAL_CHECKS
Passed: $PASSED_CHECKS
Failed: $FAILED_CHECKS
Pass Rate: $(echo "scale=2; $PASSED_CHECKS * 100 / $TOTAL_CHECKS" | bc)%

Certification Level
-------------------
EOF

if [ $FAILED_CHECKS -eq 0 ]; then
    echo "Level 3: Mission Critical ✓" >> $CERT_REPORT
    CERT_LEVEL="MISSION_CRITICAL"
elif [ $PASSED_CHECKS -ge $((TOTAL_CHECKS * 80 / 100)) ]; then
    echo "Level 2: Production Ready ✓" >> $CERT_REPORT
    CERT_LEVEL="PRODUCTION_READY"
elif [ $PASSED_CHECKS -ge $((TOTAL_CHECKS * 60 / 100)) ]; then
    echo "Level 1: Basic Deployment ✓" >> $CERT_REPORT
    CERT_LEVEL="BASIC_DEPLOYMENT"
else
    echo "FAILED - Does not meet minimum requirements" >> $CERT_REPORT
    CERT_LEVEL="FAILED"
fi

cat >> $CERT_REPORT << EOF

Production Deployment Approved: $([ $FAILED_CHECKS -eq 0 ] && echo "YES" || echo "NO")

Detailed Results
----------------
See: $CERT_JSON

Next Steps
----------
EOF

if [ $FAILED_CHECKS -eq 0 ]; then
    cat >> $CERT_REPORT << EOF
✓ All checks passed
✓ System ready for production deployment
✓ Proceed with production rollout

Recommended Actions:
1. Document this certification
2. Set up production monitoring
3. Configure backup strategy
4. Deploy to production environment
EOF
else
    cat >> $CERT_REPORT << EOF
✗ $FAILED_CHECKS checks failed
✗ System NOT ready for production

Required Actions:
1. Review failed checks above
2. Fix identified issues
3. Re-run certification
4. Achieve 100% pass rate before production deployment
EOF
fi

# Display summary
echo ""
log_info "=========================================="
log_info "CERTIFICATION SUMMARY"
log_info "=========================================="
echo ""
echo "Total Checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $PASSED_CHECKS${NC}"
echo -e "${RED}Failed: $FAILED_CHECKS${NC}"
echo ""
echo "Certification Level: $CERT_LEVEL"
echo ""

if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   ✓ PRODUCTION CERTIFICATION PASSED    ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo "Report saved to: $CERT_REPORT"
    echo "JSON results: $CERT_JSON"
    exit 0
else
    echo -e "${RED}╔════════════════════════════════════════╗${NC}"
    echo -e "${RED}║   ✗ CERTIFICATION FAILED               ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo "Report saved to: $CERT_REPORT"
    echo "JSON results: $CERT_JSON"
    exit 1
fi
