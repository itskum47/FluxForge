# FluxForge: Production-Grade Distributed Control Plane

FluxForge is a certified, production-ready distributed control plane architecture designed for orchestrating stateful workloads with strict consistency guarantees and high availability.

## Overview

FluxForge solves the problem of reliable orchestration in distributed systems. Unlike traditional CI/CD or simple task runners, FluxForge implements a formal control plane design pattern. It provides a reconciliation loop that continuously drives the actual state of the system towards a declared desired state, ensuring resilience against network partitions, node failures, and drift.

Built on the principles of atomic persistence and distributed consensus, FluxForge acts as the "brain" for your infrastructure, capable of managing thousands of agents with forensic-grade auditability.

## Key Features

- **Distributed Control Plane**: Stateless control nodes with shared state persistence.
- **Leader Election**: Automatic, lease-based leader election using Redis primitives for high availability.
- **Agent Lifecycle Management**: Full lifecycle handling including registration, heartbeats, dead-node detection, and auto-recovery.
- **Atomic Persistence and Reconciliation**: Versioned, CAS (Compare-And-Swap) based state updates ensuring no data corruption.
- **Multi-Tenant Isolation**: Strict logical isolation of resources and agents per tenant.
- **JWT Authentication**: Secure, token-based authentication for all API and agent interactions.
- **TLS Encryption**: Enforced TLS 1.2+ for all transit data.
- **Chaos-Tested Resilience**: Verified against random node kills, network partitions, and process crashes.
- **Backup and Disaster Recovery**: Automated point-in-time recovery and persistence verification.
- **Kubernetes Deployment Support**: Production-ready manifests for cloud-native deployment.
- **Load Balancer Support**: Integrated Nginx load balancing for API traffic distribution.
- **Observability with Prometheus**: Comprehensive metrics for deep system insight and alerting.

## Architecture

The FluxForge architecture follows a strict separation of concerns:

- **Control Plane Nodes**: Stateless Go services that handle API requests, run the reconciliation loop, and manage the scheduler. They form a high-availability cluster.
- **Agents**: Lightweight execution units running on target infrastructure. They pull jobs, execute commands, and report status/heartbeats.
- **Redis Persistence Layer**: The single source of truth. Uses AOF (Append-Only File) for durability and atomic operations for concurrency control.
- **Load Balancer**: An Nginx layer that distributes incoming agent and user traffic across healthy control plane nodes, handling TLS termination.
- **Prometheus Monitoring**: Scrapes metrics from all components to provide real-time visibility into system health, queue depths, and error rates.

**Logical Flow:**
1.  **Desired State Definition**: User submits a target state via API.
2.  **Persistence**: Control Plane saves state to Redis with atomic versioning.
3.  **Scheduler**: Leader node's scheduler picks up the pending state.
4.  **Disptach**: Job is dispatched to the target Agent via the control plane.
5.  **Execution**: Agent executes the logic and reports the result.
6.  **Reconciliation**: Control Plane updates the actual state to match the desired state.

## Production Guarantees

FluxForge provides specific distributed system guarantees:

- **No Split Brain**: Leader election utilizes aggressive fencing tokens (epochs) to ensure only one leader creates schedules at a time.
- **Atomic Persistence**: All state transitions use Optimistic Locking (CAS) to prevent lost updates or dirty reads.
- **Automatic Leader Failover**: System detects leader failure within 15 seconds and elects a new leader automatically.
- **Crash-Safe Reconciliation**: The reconciliation loop is idempotent; crashed tasks are re-queued and retried safely without side effects.
- **Agent Failure Detection and Recovery**: Agents are marked offline after missed heartbeats and automatically reintegrated upon reconnection.
- **Tenant Isolation**: Data access is strictly scoped by Tenant ID; cross-tenant access is cryptographically impossible without a valid token.
- **Secure Authentication**: All endpoints require signed JWTs; unauthenticated requests are rejected at the edge.

## Technology Stack

- **Go**: Core Control Plane and Agent implementation (High concurrency, strict typing).
- **Redis**: Distributed coordination, locking, and persistence store.
- **Docker**: Containerization standard for all components.
- **Kubernetes**: Orchestration platform for production deployment.
- **Nginx**: High-performance load balancing and TLS termination.
- **Prometheus**: Metrics collection, query engine, and alerting.
- **JWT**: JSON Web Tokens for stateless, secure authentication.

## Deployment Options

### Docker Deployment
Ideal for local testing and development.
```bash
docker compose up -d --build
```

### Kubernetes Deployment
Production-grade deployment using manifests.
```bash
kubectl apply -f deployments/kubernetes/
```

### Local Development
Run the control plane binaries directly.
```bash
go run control_plane/main.go
```

## Quick Start Guide

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/itskum47/FluxForge.git
    cd FluxForge
    ```

2.  **Start the Stack**
    ```bash
    docker compose up -d --build
    ```

3.  **Verify Leader Election**
    Confirm a leader has been elected among the control plane nodes.
    ```bash
    curl -k https://localhost:8443/metrics | grep flux_leader_status
    ```

4.  **Start an Agent**
    (Automatically started by Docker Compose, check logs)
    ```bash
    docker logs -f deployments-agent-1
    ```

5.  **Submit a Job**
    Obtain a token and submit a test job.
    ```bash
    TOKEN=$(./scripts/generate_token.sh default)
    curl -k -X POST https://localhost:8443/jobs \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"command":"echo Hello FluxForge","node_id":"agent-1"}'
    ```

## Security

- **JWT Authentication**: Centralized identity management. Tokens are signed with a secure secret and include tenant context.
- **TLS Enforcement**: Communication between Agents, Users, and the Control Plane is encrypted via TLS 1.2+. Non-TLS connections are rejected.
- **Tenant Isolation**: Middleware ensures that a request context is permanently bound to the token's Tenant ID, preventing privilege escalation.

## Chaos Testing and Reliability

FluxForge has undergone rigorous Chaos Engineering validation (Phase 7 Certification).
- **Chaos Monkey**: A dedicated tool (`scripts/phase7_chaos_monkey.sh`) randomly kills control plane nodes and agents during active workload execution.
- **Failover Verification**: Tests confirm that killing the leader node results in immediate failover with zero data loss.
- **Network Partitions**: The system is verified to recover gracefully from network partitions, ensuring eventual consistency.

## Monitoring and Observability

- **Prometheus Metrics**: Exposes Red (Rate, Errors, Duration) metrics for all RPCs and DB operations.
- **Alerting**: Pre-configured alerts for `AgentOffline`, `LeaderLost`, `HighErrorRate`, and `IntegrityMismatch`.
- **Dashboards**: Integrated visualization of cluster health and tenant activity.

## Project Structure

- `control_plane/`: Core orchestration logic, API, and scheduler.
- `agent/`: Source code for the remote execution agent.
- `deployments/`: Docker Compose files, Kubernetes manifests, and Nginx config.
- `scripts/`: Operational scripts for tokens, backups, chaos testing, and audits.
- `api/`: Shared data structures and constants.

## Production Certification Status

| Certification Phase | Status |
|---------------------|--------|
| Phase 1: Core Lifecycle | ✅ PASS |
| Phase 2: Orchestration | ✅ PASS |
| Phase 3: State Engine | ✅ PASS |
| Phase 4: Scheduling | ✅ PASS |
| Phase 5: HA & Failover | ✅ PASS |
| Phase 6: Multi-Tenancy | ✅ PASS |
| Phase 7: Chaos Resilience | ✅ PASS |
| Phase 8: Hardening & Ops | ✅ PASS |

## Future Improvements

- **Raft Consensus**: Migrating the WAL (Write Ahead Log) to embedded Raft for simpler deployment without Redis.
- **Sharding**: Horizontal sharding of the scheduler for massive scale (>100k agents).
- **Distributed Tracing**: Full OpenTelemetry integration for request tracing across microservices.

## License

MIT License
