# FluxForge Intelligent Scheduler Architecture

## 1. Overview
The Scheduler is the heart of FluxForge Phase 4. It acts as an **Admission Control** layer that guarantees fairness, stability, and failure isolation. It transforms a naive "First-In-First-Out" system into a deterministic, production-grade orchestrator.

## 2. Core Responsibilities
1.  **Prioritization**: Critical tasks (P0) pre-empt background tasks (P10).
2.  **Fairness**: Prevents starvation via Priority Aging.
3.  **Isolation**: Throttles tasks destined for failing domains (e.g., "zone-1").
4.  **Self-Protection**: Rejects traffic when overloaded (`DEGRADED` mode).

## 3. Data Structures

### 3.1 The Task
```go
type ReconciliationTask struct {
    ReqID         string
    Priority      int       // 0 (Highest) - 10 (Lowest)
    Cost          TaskCost  // CPU/IO cost estimation
    FailureDomain string    // "zone-us-east-1a"
    TenantID      string    // For multi-tenant isolation
    SubmitTime    time.Time // Used for aging
}
```

### 3.2 The Queue (Priority Heap)
We use a standard Min-Heap but with a twist: **Effective Priority calculation**.
`EffectivePriority = BasePriority - (WaitTimeSeconds / AgingFactor)`

*Default AgingFactor = 10s.*
- A P10 task waiting for 100s effectively becomes P0.
- This guarantees that no task waits forever.

## 4. Execution Loop Logic

The scheduler loop runs every `200ms` (guarded time budget).

1.  **Global Mode Check**:
    - If `Mode == READ_ONLY`, sleep.
    - If `Mode == DEGRADED`, reject P > 5 tasks.

2.  **Pop High Priority**:
    - Get highest *Effective Priority* task.

3.  **Admission Checks**:
    - **Node Health**: Is `NodeHealth(task.NodeID)` > Threshold?
        - If NO: Quarantine & Drop/Retry.
    - **Domain Health**: Is `FailureRate(task.Domain)` high?
        - If YES: Apply `TokenBucket` throttle.
    - **Tenant Limit**: Has `TenantID` exceeded `MaxConcurrent`?
        - If YES: Requeue with backoff penalty.

4.  **Dispatch**:
    - Send to `Dispatcher` -> Agent.
    - Record event: `DISPATCHED`.

## 5. Global Scheduler Modes

| Mode | Behavior | Use Case |
| :--- | :--- | :--- |
| **NORMAL** | Process all tasks normally. | Standard operation. |
| **DEGRADED** | Reject tasks with Priority > 5. | Partial backend outage, high DB latency. |
| **READ_ONLY** | Accept NO new tasks. Process existing. | Maintenance, upgrade. |
| **DRAINING** | Accept NO new tasks. Finish existing. | Shutdown. |

## 6. Configuration & Tuning

Key tunable parameters in `config/scheduler.yaml` (future):

- `SchedulerQueueCap`: Default 1000. Hard limit.
- `AgingFactorSeconds`: Default 10. Lower = faster aging.
- `DomainFailureThreshold`: Default 0.3 (30% failure rate triggers isolation).
- `TenantRateLimit`: Default 50 RPS.
