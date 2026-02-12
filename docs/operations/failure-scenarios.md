# FluxForge Failure Scenarios & Handling

This document details how FluxForge Phase 4 behaves under specific failure conditions.

## 1. Agent Failures

### 1.1 Sudden Disconnect (Process Crash / Network Cut)
- **Detection**: Heartbeat timeout (5s default).
- **Immediate Action**: `AgentStatus` -> `Unhealthy`.
- **Scheduler Action**:
    - Any pending tasks for this `NodeID` are paused.
    - If a task was `DISPATCHED` but not acknowledged, it is marked `UNKNOWN/TIMEOUT` after deadline.
- **Recovery**: Upon reconnect, Agent sends "Register" again. CP resumes scheduling.

### 1.2 "Lying" Agent (Stale/False Reporting)
- **Behavior**: Agent reports "Running" but does nothing.
- **Detection**: "Composite Health Score" degrades. Control Plane observes `TargetState` not reached after X reconciles.
- **Action**: 
    - Observe `FailureRate` increases.
    - `NodeHealth` score drops below 0.3.
    - **Quarantine**: Scheduler stops sending new tasks to this node until score recovers (time decay).

## 2. Infrastructure Failures

### 2.1 Failure Domain Outage (e.g., Availability Zone Down)
- **Scenario**: 50% of nodes in `us-east-1a` stop responding.
- **Detection**: `flux_domain_failure_rate{domain="us-east-1a"}` spikes > 0.3.
- **Action**: **Circuit Breaker** triggers.
    - Scheduler throttles requests to `us-east-1a` (e.g., allows only 1 req/sec).
    - Prevents wasting worker threads on doomed requests.
    - Allows the rest of the cluster to operate at full speed.

### 2.2 Reconciler Storm ("Thundering Herd")
- **Scenario**: 1000 agents simultaneous reconnect and request reconcile.
- **Handling**:
    - **Ingest**: Accepts requests into memory.
    - **Queue**: Buffers tasks.
    - **Self-Protection**: If queue > 1000, new P10 request are dropped (`503 Priority Load Shedding`).
    - **Worker Pool**: Processes at a fixed maximum concurrency.
    - **Result**: API stays up, latency increases, but system does not crash.

## 3. Control Plane Failures

### 3.1 Process Crash (OOM/Panic)
- **Impact**: In-Memory state (Queue, Timeline) is LOST.
- **Recovery**: 
    - Restart process.
    - Agents will re-heartbeat.
    - Users must re-submit "Running" jobs (Since persistence is Phase 5).
    - **Mitigation**: Deploy replicas (Future) + Persistent Store (Phase 5).

### 3.2 Overload (High CPU/Memory)
- **Action**: Enter `DEGRADED` mode.
- **Effect**: 
    - Drops all tasks with Priority > 5.
    - Increases Scheduler Loop sleep intervals (yields CPU).
    - Preserves Critical (P0) flows.
