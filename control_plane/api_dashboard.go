package main

import (
	"encoding/json"
	"net/http"

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

	metrics, err := a.dashboardService.GetDashboardMetrics(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS for local dev
	json.NewEncoder(w).Encode(metrics)
}
