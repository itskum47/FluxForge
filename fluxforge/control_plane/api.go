package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type API struct {
	store      *Store
	dispatcher *Dispatcher
}

func NewAPI(store *Store, dispatcher *Dispatcher) *API {
	return &API{store: store, dispatcher: dispatcher}
}

// handleRegister processes agent registration requests.
func (api *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Agent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NodeID == "" {
		http.Error(w, "node_id required", http.StatusBadRequest)
		return
	}

	// Overwrite existing agent is OK.
	// Capture explicit address/port if provided
	if req.Address == "" {
		req.Address = r.RemoteAddr // Fallback (though strict requirements say Agent provides it, this is a safety net)
	}
	
	api.store.UpsertAgent(&req)
	log.Printf("Agent registered: %s (%s) at %s:%d", req.NodeID, req.Hostname, req.Address, req.Port)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

// handleHeartbeat processes agent heartbeat requests.
func (api *API) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID    string `json:"node_id"`
		Timestamp int64  `json:"timestamp"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Important: Do NOT auto-register on heartbeat.
	// Only update if agent exists.
	ok := api.store.UpdateHeartbeat(req.NodeID, time.Now().Unix())
	if !ok {
		http.Error(w, "agent not registered", http.StatusNotFound)
		return
	}

	log.Printf("Heartbeat received from %s", req.NodeID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// AgentInfo is the DTO for listing agents.
type AgentInfo struct {
	NodeID   string `json:"node_id"`
	Hostname string `json:"hostname"`
	LastSeen int64  `json:"last_seen"`
}

// handleListAgents returns a list of all registered agents.
func (api *API) handleListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents := api.store.ListAgents()
	
	// Convert to DTO
	var resp []AgentInfo
	for _, a := range agents {
		resp = append(resp, AgentInfo{
			NodeID:   a.NodeID,
			Hostname: a.Hostname,
			LastSeen: a.LastSeen,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSubmitJob handles job submission from the user.
func (api *API) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID  string `json:"node_id"`
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NodeID == "" || req.Command == "" {
		http.Error(w, "node_id and command required", http.StatusBadRequest)
		return
	}

	// Validate agent exists
	agent := api.store.GetAgent(req.NodeID)
	if agent == nil {
		http.Error(w, "Agent not found", http.StatusBadRequest)
		return
	}

	// Create Job
	jobID := generateUUID() // Reusing existing helper or need to make it public? It's in agent/config.go, need to copy or move to shared. 
	// specific-fix: generateUUID was in agent/config.go. I need to duplicate it or move it. 
	// For now, I'll duplicate a simple UUID generator here to avoid package refactor chaos in Phase 2.
	
	job := &Job{
		JobID:     jobID,
		NodeID:    req.NodeID,
		Command:   req.Command,
		Status:    "queued",
		CreatedAt: time.Now().Unix(),
	}

	api.store.UpsertJob(job)

	// Synchronous dispatch (blocks until accepted)
	api.dispatcher.DispatchJob(agent, job)

	// Always return queued to keep API deterministic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": job.JobID,
		"status": "queued",
	})
}

// handleGetJob returns job details.
func (api *API) handleGetJob(w http.ResponseWriter, r *http.Request) {
	// Simple path parsing /jobs/{id}
	// Assuming mux or just stripping prefix if using standard library
	// For standard lib without mux, we might need request parsing logic if the router isn't smart.
	// But main.go uses http.HandleFunc("/jobs/", ...) ? No, I need to check main.go.
	// Implementation plan said: GET /jobs/{id}
	
	id := r.URL.Path[len("/jobs/"):]
	if id == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	job := api.store.GetJob(id)
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// handleJobResult processes job results from agents.
func (api *API) handleJobResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JobID    string `json:"job_id"`
		NodeID   string `json:"node_id"`
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exit_code"`
		Status   string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job := api.store.GetJob(req.JobID)
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Security: Verify NodeID matches
	if job.NodeID != req.NodeID {
		http.Error(w, "NodeID mismatch", http.StatusBadRequest)
		log.Printf("Security alert: NodeID mismatch for job %s. Expected %s, got %s", req.JobID, job.NodeID, req.NodeID)
		return
	}

	// Update Job
	job.Stdout = req.Stdout
	job.Stderr = req.Stderr
	job.ExitCode = req.ExitCode
	job.Status = req.Status // completed or failed

	api.store.UpsertJob(job)
	
	w.WriteHeader(http.StatusOK)
}

// Local helper for UUID (since we don't have a shared pkg yet)
func generateUUID() string {
	// simplistic implementation for now, or copy from agent.
	// actually better to use the one from agent/config.go if I can? No, separate packages (main vs main).
	// I will copy the implementation.
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[8] = b[8]&0x3f | 0x80
	b[6] = b[6]&0x0f | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
