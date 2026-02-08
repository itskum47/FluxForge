<p align="center">
  <img src="docs/assets/fluxforge-banner.png" alt="FluxForge" />
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache--2.0-blue.svg" /></a>
  <img src="https://img.shields.io/badge/status-active%20development-orange" />
  <img src="https://img.shields.io/badge/type-control--plane-6cc7b9" />
  <img src="https://img.shields.io/badge/cloud--native-yes-success" />
  <img src="https://img.shields.io/badge/event--driven-yes-purple" />
</p>

---

# FluxForge

**A Cloud-Native, Event-Driven Distributed Control Plane for Infrastructure & Applications**

FluxForge is an opinionated control plane designed to orchestrate infrastructure and application operations across distributed environments. It unifies remote execution, desired-state management, and event-driven automation into a single, extensible platform.

Think of FluxForge as:

- The execution power of Salt  
- The control-loop mindset of Kubernetes  
- The flexibility of event-driven automation  

Combined into a simpler, visual, and API-first control plane.

---

## Why FluxForge

Modern infrastructure is fragmented.

Teams operate across:
- Cloud and on-prem environments
- Containers, VMs, and bare metal
- Multiple automation and orchestration tools

Most existing solutions solve only part of the problem:
- **Configuration tools** enforce state but lack real-time orchestration
- **Orchestrators** manage workloads but not operational workflows
- **Automation tools** execute tasks but lack a global control plane

FluxForge exists to unify these concerns.

It provides a single system that:
- Maintains **desired state**
- Reacts to **events**
- Executes **distributed actions**
- Continuously observes **actual state**

---

## Core Concepts

### Control Plane
The central authority responsible for state management, workflow orchestration, event processing, and agent coordination.

### Agents
Lightweight runtimes deployed on managed nodes. Agents securely execute tasks, report state, stream logs, and maintain heartbeats with the control plane.

### Event System
An internal event bus that drives automation. State changes, task executions, alerts, and external triggers are all modeled as events.

### Workflows
Declarative, event-driven workflows modeled as directed graphs. Workflows define how the system responds to changes in infrastructure or application state.

### Modules
Pluggable execution units that provide operational capabilities such as system management, service control, container operations, and cloud API interactions.

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

- Distributed remote execution
- Event-driven automation workflows
- Declarative desired state enforcement
- Secure agent communication
- Modular and extensible execution framework
- API-first architecture
- Built-in observability of nodes, tasks, and workflows

---

## Use Cases

- Infrastructure automation at scale
- Event-based remediation and self-healing systems
- Application lifecycle orchestration
- Hybrid and multi-environment operations
- Operational workflows without brittle scripts

---

## Project Structure

```

fluxforge/
├ control_plane/    # Core control plane services
├ agent/            # Distributed agent runtime
├ modules/          # Execution modules (system, cloud, services)
├ workflows/        # Workflow engine and definitions
├ providers/        # Infrastructure provider integrations
├ sdk/              # API clients and SDKs
├ docs/             # Architecture and design documentation
├ tests/            # Automated test suites
└ tools/            # CLI and developer tooling

````

---

## Getting Started

FluxForge is designed to run locally for development and testing.

### Prerequisites
- Docker
- Docker Compose
- Git

### Quick Start

```bash
git clone https://github.com/itskum47/FluxForge.git
cd FluxForge
docker-compose up
````

This will start a local control plane, event system, and a sample agent.

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

## Project Status

FluxForge is under active development.
Core architecture is being stabilized and APIs may evolve as the platform matures.

---

## Contributing

Design discussions, bug reports, and contributions are welcome.
Please review the contribution guidelines before submitting changes.

---

## License

FluxForge is licensed under the Apache License, Version 2.0.

```
