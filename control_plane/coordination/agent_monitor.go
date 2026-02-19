package coordination

import (
	"context"
	"log"
	"time"

	"github.com/itskum47/FluxForge/control_plane/observability"
	"github.com/itskum47/FluxForge/control_plane/store"
)

// AgentMonitor periodically checks for stale agent heartbeats
type AgentMonitor struct {
	store     store.Store
	interval  time.Duration
	threshold time.Duration
}

func NewAgentMonitor(s store.Store, interval time.Duration, threshold time.Duration) *AgentMonitor {
	return &AgentMonitor{
		store:     s,
		interval:  interval,
		threshold: threshold,
	}
}

func (m *AgentMonitor) Start(ctx context.Context) {
	go m.loop(ctx)
}

func (m *AgentMonitor) loop(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	log.Printf("Starting Agent Liveness Monitor (Interval: %v, Threshold: %v)", m.interval, m.threshold)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkLiveness(ctx)
		}
	}
}

func (m *AgentMonitor) checkLiveness(ctx context.Context) {
	// In a real system, we'd use a ZSET of heartbeats.
	// Here we list all agents.
	agents, err := m.store.ListAgents(ctx, "")
	if err != nil {
		log.Printf("AgentMonitor: Failed to list agents: %v", err)
		return
	}

	activeCount := 0
	now := time.Now()
	for _, agent := range agents {
		// Debug Log
		diff := now.Sub(agent.LastHeartbeat)
		log.Printf("AgentMonitor: Check %s. Status=%s. Diff=%v. Threshold=%v", agent.NodeID, agent.Status, diff, m.threshold)

		if agent.Status == "offline" {
			continue
		}

		if diff > m.threshold {
			log.Printf("AgentMonitor: Agent %s heartbeat expired (Last: %v). Marking OFFLINE.", agent.NodeID, agent.LastHeartbeat)
			agent.Status = "offline"
			agent.UpdatedAt = now

			if err := m.store.UpsertAgent(ctx, agent.TenantID, agent); err != nil {
				log.Printf("AgentMonitor: Failed to mark agent %s offline: %v", agent.NodeID, err)
			}
		} else {
			// Count active agents
			activeCount++
		}
	}
	// Update Metric
	observability.ConnectedAgents.Set(float64(activeCount))
}
