package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/itskum47/FluxForge/control_plane/timeline"
)

// IncidentSnapshot represents a captured incident for replay.
type IncidentSnapshot struct {
	IncidentID    string `json:"incident_id"`
	StateID       string `json:"state_id"`
	NodeID        string `json:"node_id"`
	TenantID      string `json:"tenant_id"`
	FailureReason string `json:"failure_reason"`
	Timestamp     int64  `json:"timestamp"`

	// Snapshots
	SchedulerSnapshot SchedulerSnapshot         `json:"scheduler_snapshot"`
	LeaderSnapshot    LeaderSnapshot            `json:"leader_snapshot"`
	Timeline          []timeline.ReconcileEvent `json:"timeline"`
}

type SchedulerSnapshot struct {
	QueueDepth          int     `json:"queue_depth"`
	ActiveTasks         int     `json:"active_tasks"`
	WorkerSaturation    float64 `json:"worker_saturation"`
	CircuitBreakerState string  `json:"circuit_breaker_state"`
	RuntimeMode         string  `json:"runtime_mode"`
}

type LeaderSnapshot struct {
	IsLeader     bool   `json:"is_leader"`
	CurrentEpoch int64  `json:"current_epoch"`
	NodeID       string `json:"node_id"`
}

// handleListIncidents returns all captured incidents.
func (a *API) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, return empty list (incidents would be stored in DB)
	incidents := []IncidentSnapshot{}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(incidents)
}

// handleReplayIncident simulates an incident replay.
func (a *API) handleReplayIncident(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract incident ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}
	incidentID := pathParts[4]

	// For now, return a simulated replay timeline
	replay := map[string]interface{}{
		"incident_id": incidentID,
		"status":      "replay_complete",
		"timeline": []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-5 * time.Minute).Unix(),
				"event":     "State submitted",
				"details":   "Desired state created",
			},
			{
				"timestamp": time.Now().Add(-4 * time.Minute).Unix(),
				"event":     "Queued for reconciliation",
				"details":   "Priority: 5, Queue depth: 42",
			},
			{
				"timestamp": time.Now().Add(-3 * time.Minute).Unix(),
				"event":     "Reconciliation started",
				"details":   "Worker assigned",
			},
			{
				"timestamp": time.Now().Add(-2 * time.Minute).Unix(),
				"event":     "Reconciliation failed",
				"details":   "Agent timeout",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(replay)
}

// handleCaptureIncidentSnapshot captures current state for an incident.
func (a *API) handleCaptureIncidentSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stateID := r.URL.Query().Get("state_id")
	if stateID == "" {
		http.Error(w, "state_id is required", http.StatusBadRequest)
		return
	}

	// Capture asynchronously to prevent blocking scheduler
	resultChan := make(chan IncidentSnapshot, 1)
	errorChan := make(chan error, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		snapshot, err := a.captureIncidentAsync(ctx, stateID)
		if err != nil {
			errorChan <- err
			return
		}
		resultChan <- snapshot
	}()

	// Wait for result or timeout
	select {
	case snapshot := <-resultChan:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=incident-%s.json", stateID))
		json.NewEncoder(w).Encode(snapshot)

	case err := <-errorChan:
		http.Error(w, fmt.Sprintf("Failed to capture incident: %v", err), http.StatusInternalServerError)

	case <-time.After(5 * time.Second):
		http.Error(w, "Incident capture timeout", http.StatusRequestTimeout)
	}
}

// captureIncidentAsync performs the actual incident capture with timeout.
func (a *API) captureIncidentAsync(ctx context.Context, stateID string) (IncidentSnapshot, error) {
	// Capture current state
	schedMetrics := a.scheduler.GetMetrics()
	var leaderState LeaderSnapshot
	if a.elector != nil {
		state := a.elector.GetState()
		leaderState = LeaderSnapshot{
			IsLeader:     state.IsLeader,
			CurrentEpoch: state.CurrentEpoch,
			NodeID:       state.NodeID,
		}
	}

	// Get timeline events for this state
	tl := a.scheduler.GetTimeline()
	events := tl.GetAllEvents()

	return IncidentSnapshot{
		IncidentID:    fmt.Sprintf("incident-%d", time.Now().Unix()),
		StateID:       stateID,
		Timestamp:     time.Now().Unix(),
		FailureReason: "Captured via API",
		SchedulerSnapshot: SchedulerSnapshot{
			QueueDepth:          schedMetrics.QueueDepth,
			ActiveTasks:         schedMetrics.ActiveTasks,
			WorkerSaturation:    schedMetrics.WorkerSaturation,
			CircuitBreakerState: schedMetrics.CircuitBreakerState,
			RuntimeMode:         schedMetrics.RuntimeMode,
		},
		LeaderSnapshot: leaderState,
		Timeline:       events,
	}, nil
}
