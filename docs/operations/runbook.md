# FluxForge Operations Runbook

## 1. Deployment Requirements
- **Binaries**: `cp` (Control Plane), `agent` (Edge Node).
- **Ports**: 
    - CP: `:8080` (API/Metrics).
    - Agent: `:9090` (Execution).
- **Environment**:
    - `FLUX_LOG_LEVEL=info`
    - `FLUX_SCHEDULER_MODE=NORMAL`

## 2. Monitoring (Golden Signals)

| Signal | Metric | Alert Threshold | Action |
| :--- | :--- | :--- | :--- |
| **Latencyp** | `flux_reconcile_duration_seconds` | > 5s (p99) | Check DB latency / Agent network. |
| **Traffic** | `flux_scheduler_decisions_total` | Zero (flatline) | Check API availability. |
| **Errors** | `flux_api_errors_total` | > 1% rate | Check logs for panics/validation errors. |
| **Saturation** | `flux_queue_depth` | > 800 (Cap 1000) | Scale out CP or check "Stuck Worker". |
| **Staleness** | `flux_queue_oldest_task_age` | > 60s | Investigate Scheduler starvation. |

## 3. Investigating Incidents

### Scenario: "Tasks aren't running"
1.  **Check Scheduler Mode**:
    ```bash
    curl localhost:8080/scheduler/debug/snapshot | jq .mode
    ```
    If `READ_ONLY` or `DEGRADED`, check who set it or if self-protection triggered.

2.  **Check Queue Depth**:
    If `queue_depth` is high but `domain_active` tasks are low, workers might be stuck.

3.  **Check Domain Isolation**:
    Look at `domain_failures` in the snapshot. Is a specific zone being throttled?
    ```bash
    curl localhost:8080/metrics | grep domain_failure_rate
    ```

### Scenario: "Agent Flapping"
1.  Check `flux_agent_connected` metric.
2.  If flapping, Controller might be quarantining it due to `CompositeHealthScore`.
3.  Check logs for "Node quarantined due to low health score".

## 4. Emergency Procedures

### Full Restart
FluxForge v0.4.0 uses **In-Memory** state. A restart CLEARS:
- The Queue (Pending tasks lost).
- The Timeline (Audit history lost).
- Idempotency Cache.

**Procedure**:
1.  Set Mode to `DRAINING`: `POST /admin/mode { "mode": "DRAINING" }` (If implemented).
2.  Wait for `queue_depth` to reach 0.
3.  Restart Service.

### Clearing Stuck Queue
If queue is full of "poison" tasks:
1.  Restart CP (Purges queue).
2.  Or use `POST /scheduler/queue/purge` (Future feature).
