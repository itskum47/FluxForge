<p align="center">
  <img src="docs/assets/fluxforge-banner.png" alt="FluxForge" />
</p>

<br/>

<p align="center">
  <img src="https://img.shields.io/badge/license-Apache--2.0-blue" />
  <img src="https://img.shields.io/badge/status-active--development-orange" />
  <img src="https://img.shields.io/badge/control--plane-distributed-6cc7b9" />
  <img src="https://img.shields.io/badge/event--driven-yes-purple" />
</p>

<br/><br/>

<p align="center">
  <strong>Distributed Control Plane for Infrastructure & Applications</strong><br/>
  <span style="opacity:0.75">Event-driven • Declarative • Cloud-native</span>
</p>

<br/><br/>

<p align="center">
  FluxForge is an opinionated, event-driven control plane for operating infrastructure
  and applications across distributed environments.
</p>

<br/>

<p align="center">
  It unifies remote execution, desired-state management, and automation workflows
  into a single, extensible system.
</p>

<br/><br/>

<p align="center">
  <em>
    Execution power of Salt · Control-loop mindset of Kubernetes · Event-driven automation
  </em>
</p>

<br/><br/>

---

## Why FluxForge

Modern infrastructure is fragmented.

Teams operate across cloud and on-prem environments, containers, virtual machines,
bare metal systems, and multiple automation tools that rarely work together cleanly.

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
The central authority responsible for desired-state management, workflow orchestration,
event processing, and coordination across all managed nodes.

### Agents
Lightweight runtimes deployed on managed infrastructure. Agents securely execute tasks,
maintain heartbeats, stream logs, and continuously report state back to the control plane.

### Event System
An internal event bus that drives all automation. State changes, executions, alerts,
and external triggers are modeled as events and processed uniformly.

### Workflows
Declarative, event-driven workflows modeled as directed graphs. Workflows define how
the system responds to infrastructure and application state
