# FluxForge Architecture Overview

## 1. Introduction
FluxForge is a distributed control plane designed to orchestrate ephemeral tasks across unreliable execution nodes (Agents). It guarantees **at-least-once** execution and **eventual consistency** through a robust reconciliation loop.

## 2. System Components

### 2.1 Control Plane (CP)
The brain of the system.
- **API Server**: REST/gRPC interface for users and agents.
- **Ingest Queue**: Decouples write path from processing.
- **Scheduler**: Admission control and priority management.
- **Reconciler**: The "Operator" loop that drives state towards `DesiredState`.
- **Dispatcher**: Handles reliable delivery to agents.

### 2.2 Agent (Edge Node)
The execution unit.
- **Heartbeat**: Signals liveness and capability (CPU/RAM).
- **Executor**: Runs jobs (Shell/Docker/WASM).
- **Reporter**: Streams stdout/stderr/exit-code back to CP.

### 2.3 Storage (Phase 5+)
- **State Store**: Postgres (Metadata, Desired States).
- **Event Store**: Time-series log of all transitions.
- **Cache**: Redis (Hot queue, Idempotency keys).

## 3. Data Flow

### 3.1 The Reconciliation Loop
1.  **User Definition**: User `POST /states` declaring `check_cmd` and `apply_cmd`.
2.  **Drift Detection**: Reconciler polls Agent. Runs `check_cmd`.
    - Exit Code 0: Compliant. Stop.
    - Exit Code Non-Zero: Drifted. Proceed.
3.  **Scheduling**: `ReconciliationTask` created and pushed to Priority Queue.
4.  **Dispatcher**: Pops task, sends to Agent via HTTP/gRPC.
5.  **Execution**: Agent runs `apply_cmd`.
6.  **Verification**: Reconciler re-runs `check_cmd` to confirm fix.

## 4. Key Decisions & Trade-offs
- **Pull vs Push**: We use **Push-based Dispatch** for lower latency, but **Pull-based Heartbeats** for liveness.
- **Consistency**: We favor **Availability** (AP) for ingestion, but strict **Consistency** (CP) for State transitions (via Lease/Locking).
- **Isolation**: Tenant isolation is enforced at the *Scheduler Admission* level, not just the *Execution* level, protecting CP resources.
