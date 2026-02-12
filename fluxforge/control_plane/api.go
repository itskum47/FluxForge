package main

import (
	"fluxforge/control_plane/scheduler"

	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"fluxforge/control_plane/idempotency"
)

type API struct {
	store      *Store
	dispatcher *Dispatcher
	reconciler *Reconciler
	scheduler  *scheduler.Scheduler

	idempotency *idempotency.Store
}

func NewAPI(store *Store, dispatcher *Dispatcher, reconciler *Reconciler, sched *scheduler.Scheduler) *API {
	return &API{
		store:       store,
		dispatcher:  dispatcher,
		reconciler:  reconciler,
		scheduler:   sched,
		idempotency: idempotency.NewStore(),
	}
}

// Wrapper for capturing response
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func (a *API) withIdempotency(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Flux-Idempotency-Key")
		if key == "" {
			next(w, r)
			return
		}

		if resp, found := a.idempotency.Get(key); found {
			for k, v := range resp.Headers {
				for _, val := range v {
					w.Header().Add(k, val)
				}
			}
			w.WriteHeader(resp.StatusCode)
			w.Write(resp.Body)
			return
		}

		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next(rec, r)

		a.idempotency.Set(key, idempotency.Response{
			StatusCode: rec.statusCode,
			Body:       rec.body,
			Headers:    rec.Header(),
		})
	}
}

// -- Phase 1: Agent Registration & Heartbeat --

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Attestation Stub
	if sig := r.Header.Get("X-Agent-Signature"); sig == "" {
		// Just log warning for now as it's a stub
		log.Println("Warning: Agent registration missing X-Agent-Signature")
	}

	var agent Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if agent.NodeID == "" {
		http.Error(w, "NodeID is required", http.StatusBadRequest)
		return
	}

	// Update LastSeen
	agent.LastSeen = time.Now().Unix()

	// Register agent
	a.store.UpsertAgent(&agent)

	log.Printf("Agent registered: %s (%s)", agent.NodeID, agent.Hostname)
	w.WriteHeader(http.StatusOK)
}

func (a *API) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
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

	if !a.store.UpdateHeartbeat(req.NodeID, time.Now().Unix()) {
		log.Printf("Heartbeat from unknown agent: %s", req.NodeID)
		http.Error(w, "Agent not registered", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) handleListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents := a.store.ListAgents()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// -- Phase 2: Remote Execution --

func (a *API) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
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

	agent := a.store.GetAgent(req.NodeID)
	if agent == nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	jobID := generateUUID()
	job := &Job{
		JobID:     jobID,
		NodeID:    req.NodeID,
		Command:   req.Command,
		Status:    "queued",
		CreatedAt: time.Now().Unix(),
	}

	a.store.UpsertJob(job)

	// Async dispatch
	go a.dispatcher.DispatchJob(agent, job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "queued",
	})
}

func (a *API) handleGetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract JobID from path /jobs/{job_id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}
	jobID := pathParts[2]

	job := a.store.GetJob(jobID)
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (a *API) handleListJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs := a.store.ListJobs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (a *API) handleJobResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var result struct {
		JobID    string `json:"job_id"`
		Status   string `json:"status"`
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exit_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job := a.store.GetJob(result.JobID)
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Update job state
	job.Status = result.Status
	job.Stdout = result.Stdout
	job.Stderr = result.Stderr
	job.ExitCode = result.ExitCode

	a.store.UpsertJob(job)

	log.Printf("Job %s completed with status: %s", job.JobID, job.Status)
	w.WriteHeader(http.StatusOK)
}

// -- Phase 3: Desired State Management --

func (a *API) handleCreateState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID          string `json:"node_id"`
		CheckCmd        string `json:"check_cmd"`
		ApplyCmd        string `json:"apply_cmd"`
		DesiredExitCode int    `json:"desired_exit_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stateID := generateUUID()
	state := &State{
		StateID:         stateID,
		NodeID:          req.NodeID,
		CheckCmd:        req.CheckCmd,
		ApplyCmd:        req.ApplyCmd,
		DesiredExitCode: req.DesiredExitCode,
		Status:          "pending",
	}

	a.store.UpsertState(state)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"state_id": stateID,
		"status":   "pending",
	})
}

func (a *API) handleGetState(w http.ResponseWriter, r *http.Request) {
	// Extract StateID from path /states/{state_id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid state ID", http.StatusBadRequest)
		return
	}
	stateID := pathParts[2]

	state := a.store.GetState(stateID)
	if state == nil {
		http.Error(w, "State not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (a *API) handleListStates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	states := a.store.ListStates()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(states)
}

func (a *API) handleReconcileState(w http.ResponseWriter, r *http.Request) {
	// Extract StateID from path /states/{state_id}/reconcile
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid state ID", http.StatusBadRequest)
		return
	}
	stateID := pathParts[2]

	// Check if agent is busy
	state := a.store.GetState(stateID)
	if state == nil {
		http.Error(w, "State not found", http.StatusNotFound)
		return
	}

	// Create Reconciliation Task
	task := &scheduler.ReconciliationTask{
		ReqID:    generateUUID(),
		NodeID:   state.NodeID,
		TenantID: "default", // Hardcoded for now, Phase 4.4 will add tenancy
		Priority: 5,         // Default priority
		Deadline: time.Now().Add(1 * time.Minute),
		StateID:  stateID,
	}

	// Submit to Scheduler
	if err := a.scheduler.Submit(task); err != nil {
		log.Printf("Scheduler rejected task: %v", err)
		http.Error(w, "Service Overloaded", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "reconciliation_queued",
		"task_id": task.ReqID,
	})
}
