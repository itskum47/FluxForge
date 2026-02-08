# FluxForge

**A Cloud-Native, Event-Driven Distributed Control Plane for Infrastructure and Applications**

FluxForge is an opinionated control plane designed to orchestrate infrastructure and application operations across distributed environments. It combines remote execution, state enforcement, and event-driven automation into a single, API-first platform.

Think of FluxForge as:

> **Salt + Kubernetes + Event-Driven Automation**
> unified under a simpler, visual, and extensible control plane.

---

## Why FluxForge

Modern infrastructure teams operate across heterogeneous environments: cloud, on-prem, containers, VMs, and edge. Existing tools often solve only part of the problem:

* **Configuration tools** focus on state, but lack real-time orchestration
* **Orchestrators** manage workloads, but not operational workflows
* **Automation tools** execute tasks, but lack a global control plane

FluxForge exists to unify these concerns.

It provides a single control plane that:

* Understands **desired state**
* Reacts to **events**
* Executes **distributed actions**
* Continuously reports **actual state**

---

## Core Concepts

### Control Plane

The central authority that maintains desired state, processes events, schedules workflows, and coordinates agents.

### Agents

Lightweight runtimes deployed on managed nodes. Agents securely execute tasks, report state, stream logs, and respond to control plane commands.

### Event Bus

An internal event system that drives automation. All state changes, task executions, and external triggers flow through events.

### Workflows

Declarative, event-driven workflows modeled as directed graphs. Workflows define how the system reacts to changes in infrastructure or application state.

### Modules

Pluggable execution units that provide capabilities such as system operations, service management, container control, or cloud API interactions.

---

## High-Level Architecture

```
User / API / UI
        ↓
Control Plane API
        ↓
State Engine + Workflow Engine
        ↓
Event Bus
        ↓
Distributed Agent Fleet
        ↓
Infrastructure & Applications
```

---

## Key Features

* Distributed remote execution
* Event-driven automation workflows
* Declarative desired state management
* Secure agent communication
* Modular execution framework
* API-first design
* Visual observability of nodes, tasks, and workflows

---

## Use Cases

* Infrastructure automation at scale
* Event-based remediation and self-healing systems
* Application lifecycle orchestration
* Hybrid and multi-environment operations
* Operational workflows without brittle scripts

---

## Project Structure

```
fluxforge/
├ control_plane/    # Core control plane services
├ agent/            # Distributed agent runtime
├ modules/          # Execution modules (system, cloud, services)
├ workflows/        # Workflow definitions and engine
├ providers/        # Infrastructure provider integrations
├ sdk/              # API clients and SDKs
├ docs/             # Architecture and design documentation
├ tests/            # Automated test suites
└ tools/            # CLI and developer tooling
```

---

## Getting Started

FluxForge is designed to be deployable locally for development and testing.

### Prerequisites

* Docker
* Docker Compose
* Git

### Quick Start

```bash
git clone https://github.com/itskum47/FluxForge.git
cd FluxForge
docker-compose up
```

This starts a local control plane, event bus, and a sample agent.

---

## Documentation

* Architecture overview: `docs/architecture.md`
* Agent protocol: `docs/agent.md`
* Workflow engine: `docs/workflows.md`
* Module development: `docs/modules.md`

---

## Design Philosophy

FluxForge is intentionally opinionated.

* Control plane first
* Event-driven by default
* Explicit state over implicit behavior
* Extensibility through well-defined contracts
* Observability as a first-class concern

---

## Status

FluxForge is under active development. APIs and internal components are evolving as the core architecture is finalized.

---

## Contributing

Contributions, design discussions, and feedback are welcome.
Please review the contribution guidelines before submitting changes.

---

## License

Apache 2.0
