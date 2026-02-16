#!/bin/bash
# Phase 7 Production Hardening - Long Duration Stability Test
# Runs continuous load for 24-72 hours and monitors for leaks

set -e

DURATION_HOURS=${1:-24}
DURATION_SECONDS=$((DURATION_HOURS * 3600))

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

API_BASE="http://localhost:8080"
METRICS_FILE="/tmp/stability_metrics.csv"
JOBS_FILE="/tmp/stability_jobs.log"

log_info "=========================================="
log_info "Long Duration Stability Test"
log_info "Duration: $DURATION_HOURS hours"
log_info "=========================================="

# Initialize metrics file
echo "timestamp,queue_depth,active_agents,memory_kb,goroutines,cpu_percent" > $METRICS_FILE

# Metrics collection function
collect_metrics() {
    while true; do
        TIMESTAMP=$(date +%s)
        
        # Get dashboard metrics
        METRICS=$(curl -s $API_BASE/api/dashboard 2>/dev/null || echo '{}')
        QUEUE=$(echo $METRICS | jq -r '.queue_depth // 0')
        AGENTS=$(echo $METRICS | jq -r '.active_agents // 0')
        
        # Get process metrics
        CONTROL_PID=$(pgrep -f "control_plane" | head -1)
        if [ -n "$CONTROL_PID" ]; then
            MEMORY=$(ps -o rss= -p $CONTROL_PID)
            CPU=$(ps -o %cpu= -p $CONTROL_PID)
        else
            MEMORY=0
            CPU=0
        fi
        
        # Get goroutine count
        GOROUTINES=$(curl -s $API_BASE/debug/pprof/goroutine?debug=1 2>/dev/null | grep -c "^goroutine" || echo 0)
        
        # Log metrics
        echo "$TIMESTAMP,$QUEUE,$AGENTS,$MEMORY,$GOROUTINES,$CPU" >> $METRICS_FILE
        
        # Display current status
        echo -ne "\r$(date '+%Y-%m-%d %H:%M:%S') | Queue: $QUEUE | Agents: $AGENTS | Mem: ${MEMORY}KB | Goroutines: $GOROUTINES | CPU: ${CPU}%    "
        
        sleep 60
    done
}

# Continuous job submission function
submit_jobs() {
    while true; do
        # Submit 5 jobs per second
        for i in {1..5}; do
            AGENT_ID=$((RANDOM % 10 + 1))
            curl -s -X POST $API_BASE/jobs \
                -H "Content-Type: application/json" \
                -d "{\"node_id\":\"agent-$AGENT_ID\",\"command\":\"sleep 1\"}" \
                >> $JOBS_FILE 2>&1 &
        done
        sleep 1
    done
}

# Start background processes
log_info "Starting metrics collection..."
collect_metrics &
METRICS_PID=$!

log_info "Starting continuous job submission (5 jobs/sec)..."
submit_jobs &
JOBS_PID=$!

log_info "Test running. PIDs: Metrics=$METRICS_PID, Jobs=$JOBS_PID"
log_info "Will run for $DURATION_HOURS hours..."

# Wait for duration
sleep $DURATION_SECONDS

# Stop background processes
log_info "Stopping background processes..."
kill $METRICS_PID $JOBS_PID 2>/dev/null || true

# Analyze results
log_info "=========================================="
log_info "Analyzing Results"
log_info "=========================================="

# Memory analysis
INITIAL_MEM=$(head -2 $METRICS_FILE | tail -1 | cut -d',' -f4)
FINAL_MEM=$(tail -1 $METRICS_FILE | cut -d',' -f4)
MEM_GROWTH=$(echo "scale=2; ($FINAL_MEM - $INITIAL_MEM) / $INITIAL_MEM * 100" | bc)

log_info "Memory: Initial=${INITIAL_MEM}KB, Final=${FINAL_MEM}KB, Growth=${MEM_GROWTH}%"

if (( $(echo "$MEM_GROWTH < 20" | bc -l) )); then
    log_pass "Memory growth < 20% ✓"
else
    log_fail "Memory growth >= 20% (LEAK DETECTED)"
fi

# Goroutine analysis
INITIAL_GR=$(head -2 $METRICS_FILE | tail -1 | cut -d',' -f5)
FINAL_GR=$(tail -1 $METRICS_FILE | cut -d',' -f5)
GR_GROWTH=$(echo "scale=2; ($FINAL_GR - $INITIAL_GR) / $INITIAL_GR * 100" | bc)

log_info "Goroutines: Initial=$INITIAL_GR, Final=$FINAL_GR, Growth=${GR_GROWTH}%"

if (( $(echo "$GR_GROWTH < 10" | bc -l) )); then
    log_pass "Goroutine growth < 10% ✓"
else
    log_fail "Goroutine growth >= 10% (LEAK DETECTED)"
fi

# Queue depth analysis
MAX_QUEUE=$(tail -n +2 $METRICS_FILE | cut -d',' -f2 | sort -n | tail -1)
log_info "Max queue depth: $MAX_QUEUE"

if [ "$MAX_QUEUE" -lt 1000 ]; then
    log_pass "Queue depth remained bounded ✓"
else
    log_fail "Queue depth exceeded 1000"
fi

# Job submission analysis
TOTAL_JOBS=$(wc -l < $JOBS_FILE)
EXPECTED_JOBS=$((DURATION_SECONDS * 5))
log_info "Jobs submitted: $TOTAL_JOBS (expected: ~$EXPECTED_JOBS)"

# Generate report
REPORT_FILE="/tmp/stability_report_${DURATION_HOURS}h.txt"
cat > $REPORT_FILE << EOF
FluxForge Stability Test Report
================================
Duration: $DURATION_HOURS hours
Test Date: $(date)

Memory Analysis:
  Initial: ${INITIAL_MEM} KB
  Final: ${FINAL_MEM} KB
  Growth: ${MEM_GROWTH}%
  Status: $([ $(echo "$MEM_GROWTH < 20" | bc -l) -eq 1 ] && echo "PASS" || echo "FAIL")

Goroutine Analysis:
  Initial: $INITIAL_GR
  Final: $FINAL_GR
  Growth: ${GR_GROWTH}%
  Status: $([ $(echo "$GR_GROWTH < 10" | bc -l) -eq 1 ] && echo "PASS" || echo "FAIL")

Queue Depth:
  Maximum: $MAX_QUEUE
  Status: $([ $MAX_QUEUE -lt 1000 ] && echo "PASS" || echo "FAIL")

Job Submission:
  Total Jobs: $TOTAL_JOBS
  Rate: $(echo "scale=2; $TOTAL_JOBS / $DURATION_SECONDS" | bc) jobs/sec

Metrics File: $METRICS_FILE
Jobs Log: $JOBS_FILE
EOF

log_info "Report saved to: $REPORT_FILE"
cat $REPORT_FILE

log_info "=========================================="
log_info "Stability test complete"
log_info "=========================================="
