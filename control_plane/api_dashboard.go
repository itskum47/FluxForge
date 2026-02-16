package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/itskum47/FluxForge/control_plane/coordination"
	"github.com/itskum47/FluxForge/control_plane/middleware"
)

// DashboardMetrics represents the complete dashboard state.
type DashboardMetrics struct {
	// Scheduler Metrics
	QueueDepth          int     `json:"queue_depth"`
	ActiveTasks         int     `json:"active_tasks"`
	MaxConcurrency      int     `json:"max_concurrency"`
	WorkerSaturation    float64 `json:"worker_saturation"`
	CircuitBreakerState string  `json:"circuit_breaker_state"`
	AdmissionMode       string  `json:"admission_mode"`
	RuntimeMode         string  `json:"runtime_mode"`

	// Leadership Metrics
	IsLeader          bool   `json:"is_leader"`
	CurrentEpoch      int64  `json:"current_epoch"`
	LeaderTransitions int64  `json:"leader_transitions"`
	NodeID            string `json:"node_id"`

	// Store Metrics
	PendingStates int `json:"pending_states"`
	DriftedStates int `json:"drifted_states"`
	ActiveAgents  int `json:"active_agents"`

	// Multi-Cluster Support (Phase 6.4)
	ClusterID   string `json:"cluster_id"`
	ClusterRole string `json:"cluster_role"` // leader, follower, standby
	Region      string `json:"region"`

	// Timestamp
	Timestamp int64 `json:"timestamp"`
}

// handleGetDashboard returns the current dashboard metrics.
func (a *API) handleGetDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	metrics := a.collectDashboardMetrics(r.Context(), tenantID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS for local dev
	json.NewEncoder(w).Encode(metrics)
}

// collectDashboardMetrics gathers metrics from all components.
func (a *API) collectDashboardMetrics(ctx context.Context, tenantID string) DashboardMetrics {
	// Scheduler Metrics (Global for now, ideally filtered by tenant)
	schedMetrics := a.scheduler.GetMetrics()

	// Leadership Metrics (Global)
	var leaderState coordination.LeaderState
	if a.elector != nil {
		leaderState = a.elector.GetState()
	}

	// Store Metrics (Tenant Scoped)
	pending, _ := a.store.CountStatesByStatus(ctx, tenantID, "pending")
	drifted, _ := a.store.CountStatesByStatus(ctx, tenantID, "drifted")
	agents, _ := a.store.ListAgents(ctx, tenantID)

	return DashboardMetrics{
		// Scheduler
		QueueDepth:          schedMetrics.QueueDepth,
		ActiveTasks:         schedMetrics.ActiveTasks,
		MaxConcurrency:      schedMetrics.MaxConcurrency,
		WorkerSaturation:    schedMetrics.WorkerSaturation,
		CircuitBreakerState: schedMetrics.CircuitBreakerState,
		AdmissionMode:       schedMetrics.AdmissionMode,
		RuntimeMode:         schedMetrics.RuntimeMode,

		// Leadership
		IsLeader:          leaderState.IsLeader,
		CurrentEpoch:      leaderState.CurrentEpoch,
		LeaderTransitions: leaderState.Transitions,
		NodeID:            leaderState.NodeID,

		// Store
		PendingStates: pending,
		DriftedStates: drifted,
		ActiveAgents:  len(agents),

		// Multi-Cluster (Phase 6.4)
		ClusterID: "cluster-primary", // TODO: Get from config
		ClusterRole: func() string {
			if leaderState.IsLeader {
				return "leader"
			}
			return "follower"
		}(),
		Region: "us-east-1", // TODO: Get from config

		// Timestamp
		Timestamp: time.Now().Unix(),
	}
}
