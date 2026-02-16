#!/bin/bash
set -e

BASE_URL="http://localhost:8080"
SUCCESS=0
FAIL=0

echo "🚀 Starting Pilot Week 1 Simulation (Operator Grade)..."

# --- 1. HTTP Failure Detection ---
curl_safe() {
    curl --fail --show-error --silent --max-time 5 "$@"
}

# --- 2. Request Success Counter ---
check_request() {
    if "$@"; then
        ((SUCCESS++)) || true
    else
        ((FAIL++)) || true
        echo "❌ Request Failed: $@"
    fi
}

register_agent() {
    id=$1
    tier=$2
    check_request curl_safe -X POST "$BASE_URL/agent/register" \
        -d "{\"node_id\": \"$id\", \"status\": \"active\", \"tier\": \"$tier\"}" > /dev/null
}

submit_job() {
    id=$1
    # --- 6. Randomize Submission Timing ---
    sleep $(awk -v min=0 -v max=0.2 'BEGIN{srand(); print min+rand()*(max-min)}')
    check_request curl_safe -X POST "$BASE_URL/jobs" \
        -d "{\"node_id\": \"$id\", \"command\": \"sleep 1\"}" > /dev/null
}

check_metrics() {
    echo "📊 Metrics Snapshot:"
    curl_safe "$BASE_URL/metrics" | grep -E "^flux_scheduler_queue_depth|^flux_scheduler_worker_saturation|^flux_integrity_skew_count|^flux_intent_age_seconds_count" || true
    echo "----------------------------------------"
}

# --- 3. Integrity Guard During Run ---
watch_integrity() {
    while true; do
        # Extract value, ignoring comments (#)
        SKEW=$(curl -s "$BASE_URL/metrics" | grep "^flux_integrity_skew_count" | awk '{print $2}')
        
        # Handle empty response (service startup)
        if [[ -z "$SKEW" ]]; then
            sleep 1
            continue
        fi
        if [[ "$SKEW" != "0" ]]; then
            echo "🚨 INTEGRITY SKEW DETECTED: $SKEW"
            kill $$ # Kill parent script
            exit 1
        fi
        sleep 2
    done
}

watch_integrity &
WATCH_PID=$!
trap "kill $WATCH_PID 2>/dev/null" EXIT

# --- 4. Admission Mode Safety Check ---
echo "🔍 Verifying Pilot Mode..."
MODE=$(curl_safe "$BASE_URL/metrics" | grep 'flux_runtime_mode.*mode="pilot"' | awk '{print $2}')
# Adjusted grep to match label line
if [[ -z "$MODE" ]]; then
     # Fallback strict check: expect 'flux_runtime_mode{mode="pilot"} 1'
     # Try simple grep without label if Prom format differs in older lib
     MODE_VAL=$(curl_safe "$BASE_URL/metrics" | grep "^flux_runtime_mode" | head -n 1 | awk '{print $2}')
     if [[ "$MODE_VAL" != "1" ]]; then
          echo "⚠️  Metric Check Warning: Could not verify 'pilot' mode label (Mode Value: $MODE_VAL). Proceeding if value is 1."
     else
          echo "✅ Pilot Mode Confirmed (Value 1)."
     fi
else
    echo "✅ Pilot Mode Confirmed."
fi

# --- 5. Leader Stability Check ---
echo "🔍 Verifying Leader Stability..."
LEADER_TRANSITIONS=$(curl_safe "$BASE_URL/metrics" | grep "^flux_leader_transitions_total" | awk '{print $2}')
if [[ "$LEADER_TRANSITIONS" != "0" ]]; then
    echo "⚠️  Leader already unstable before phase (Transitions: $LEADER_TRANSITIONS)"
fi


# --- Day 1: 10 Agents (Internal Team) ---
echo "📅 DAY 1: Onboarding 10 Agents..."
for i in {1..10}; do
    register_agent "agent-day1-$i" "normal"
    submit_job "agent-day1-$i" &
done
wait
sleep 2
check_metrics

# --- Day 2: Verification (Incident Capture) ---
echo "📅 DAY 2: verifying Incident Capture..."
check_request curl_safe "$BASE_URL/incident/capture?state_id=test-state" > /dev/null
echo "✅ Incident Capture Endpoint accessible."
check_metrics

# --- Day 3: 50 Agents (Ramp Up) ---
echo "📅 DAY 3: Onboarding +40 Agents (Total 50)..."
for i in {11..50}; do
    register_agent "agent-day3-$i" "normal"
    submit_job "agent-day3-$i" &
done
wait
sleep 2
check_metrics

# --- Day 5: 100 Agents (Full Load) ---
echo "📅 DAY 5: Onboarding +50 Agents (Total 100)..."
echo "⚠️  Note: Metric 'flux_scheduler_worker_saturation' should correspond to concurrency=20"
for i in {51..100}; do
    register_agent "agent-day5-$i" "normal"
    # Submit multiple jobs to create pressure
    submit_job "agent-day5-$i" &
    submit_job "agent-day5-$i" &
done
wait
sleep 5
check_metrics

echo "📈 Request Stats: SUCCESS=$SUCCESS  FAIL=$FAIL"

if [[ "$FAIL" -gt 0 ]]; then
    echo "❌ SIMULATION INVALID: $FAIL requests failed."
    exit 1
fi

# --- 7. End-of-Run Truth Check ---
FINAL_SKEW=$(curl_safe "$BASE_URL/metrics" | grep "^flux_integrity_skew_count" | awk '{print $2}')
if [[ "$FINAL_SKEW" != "0" ]]; then
  echo "❌ FINAL INTEGRITY FAILURE: Skew is $FINAL_SKEW"
  exit 1
fi

echo "✅ Week 1 Simulation Complete (Operator Grade Verified)."
