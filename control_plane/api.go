package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/itskum47/FluxForge/control_plane/coordination"
	"github.com/itskum47/FluxForge/control_plane/idempotency"
	"github.com/itskum47/FluxForge/control_plane/incident"
	"github.com/itskum47/FluxForge/control_plane/middleware"
	"github.com/itskum47/FluxForge/control_plane/observability"
	"github.com/itskum47/FluxForge/control_plane/scheduler"
	"github.com/itskum47/FluxForge/control_plane/store"
)

type API struct {
	store      store.Store
	dispatcher *Dispatcher
	reconciler *Reconciler
	scheduler  *scheduler.Scheduler
	elector    *coordination.LeaderElector

	// Services
	dashboardService *DashboardService
	wsHub            *MetricsHub

	idempotency *idempotency.Store

	// Storm Protection
	heartbeatLimiter *rate.Limiter
	reconcileLimiter *rate.Limiter
}

func NewAPI(store store.Store, dispatcher *Dispatcher, reconciler *Reconciler, scheduler *scheduler.Scheduler, elector *coordination.LeaderElector, idempotencyStore *idempotency.Store) *API {
	api := &API{
		store:       store,
		dispatcher:  dispatcher,
		reconciler:  reconciler,
		scheduler:   scheduler,
		elector:     elector,
		idempotency: idempotencyStore,
		// Allow 100 heartbeats/sec, burst 200
		heartbeatLimiter: rate.NewLimiter(rate.Limit(100), 200),
		// Allow 10 reconciles/sec, burst 20
		reconcileLimiter: rate.NewLimiter(rate.Limit(10), 20),
	}

	// Initialize Services
	api.dashboardService = NewDashboardService(store, scheduler, elector)

	// Initialize WebSocket hub
	api.wsHub = NewMetricsHub(api)

	return api
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

		if resp, found := a.idempotency.Get(r.Context(), key); found {
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

		a.idempotency.Set(r.Context(), key, idempotency.Response{
			StatusCode: rec.statusCode,
			Body:       rec.body,
			Headers:    rec.Header(),
		})
	}
}

// -- Phase 1: Agent Registration & Heartbeat --

// writeRateLimitError writes a 429 response with Jittered Retry-After
func (a *API) writeRateLimitError(w http.ResponseWriter) {
	// Phase 5.1: Track API rate limiting
	observability.APIRateLimited.WithLabelValues("heartbeat").Inc()

	// Jitter: 1s base + 0-1000ms random
	retryAfter := 1000 + rand.Intn(1000)
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter/1000)) // Seconds
	http.Error(w, "Too Many Requests (Storm Protection Active)", http.StatusTooManyRequests)
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Attestation Stub
	if sig := r.Header.Get("X-Agent-Signature"); sig == "" {
		log.Println("Warning: Agent registration missing X-Agent-Signature")
	}

	var agent store.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if agent.NodeID == "" {
		http.Error(w, "NodeID is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if agent.Status == "" {
		agent.Status = "active"
	}
	agent.LastHeartbeat = time.Now()

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	agent.TenantID = tenantID

	if err := a.store.UpsertAgent(r.Context(), tenantID, &agent); err != nil {
		log.Printf("Failed to register agent %s: %v", agent.NodeID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Phase 6.1: Update Node Health with Tier
	// We use 1.0 (perfect health) for new registrations
	a.scheduler.UpdateNodeHealth(agent.NodeID, "registration", 1.0, agent.Tier)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

// handleSetAdmissionMode updates the scheduler admission mode (Pilot Kill Switch).
func (a *API) handleSetAdmissionMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Mode string `json:"mode"` // normal, drain, freeze
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var mode scheduler.AdmissionMode
	switch req.Mode {
	case "normal":
		mode = scheduler.AdmissionNormal
	case "drain":
		mode = scheduler.AdmissionDrain
	case "freeze":
		mode = scheduler.AdmissionFreeze
	default:
		http.Error(w, "Invalid mode. Use: normal, drain, freeze", http.StatusBadRequest)
		return
	}

	a.scheduler.SetAdmissionMode(mode)
	log.Printf("🚨 ADMIN ACTION: Admission Mode set to %s", req.Mode)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated", "mode": req.Mode})
}

func (a *API) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	// Storm Protection
	if !a.heartbeatLimiter.Allow() {
		a.writeRateLimitError(w)
		return
	}

	var req struct {
		NodeID string `json:"node_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NodeID == "" {
		http.Error(w, "NodeID is required", http.StatusBadRequest)
		return
	}

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := a.store.UpdateAgentHeartbeat(r.Context(), tenantID, req.NodeID, time.Now()); err != nil {
		log.Printf("Failed to update heartbeat for %s: %v", req.NodeID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (a *API) handleListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// tenantID := r.URL.Query().Get("tenant_id") // Deprecated
	agents, err := a.store.ListAgents(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// -- Phase 2: Remote Execution --

func (a *API) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var job store.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if job.NodeID == "" || job.Command == "" {
		http.Error(w, "node_id and command are required", http.StatusBadRequest)
		return
	}

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	job.TenantID = tenantID

	// Verify agent exists
	agent, err := a.store.GetAgent(r.Context(), tenantID, job.NodeID)
	if err != nil {
		http.Error(w, "Internal Server Error checking agent", http.StatusInternalServerError)
		return
	}
	if agent == nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	job.JobID = generateUUID()
	job.Status = "queued"
	job.CreatedAt = time.Now()

	if err := a.store.CreateJob(r.Context(), tenantID, &job); err != nil {
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	// Async Dispatch
	go a.dispatcher.DispatchJob(context.Background(), agent, &job)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
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

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	job, err := a.store.GetJob(r.Context(), tenantID, jobID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	// store interface ListJobs(ctx, tenantID, nodeID, limit)
	jobs, err := a.store.ListJobs(r.Context(), tenantID, nodeID, 50) // Default limit 50
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := a.store.UpdateJobStatus(r.Context(), tenantID, result.JobID, result.Status, result.ExitCode, result.Stdout, result.Stderr); err != nil {
		log.Printf("Failed to update job status: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Job %s completed with status: %s", result.JobID, result.Status)
	w.WriteHeader(http.StatusOK)
}

// -- Phase 3: Desired State Management --

func (a *API) handleCreateState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var state store.DesiredState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if state.NodeID == "" {
		http.Error(w, "node_id is required", http.StatusBadRequest)
		return
	}

	// Generate ID if missing
	if state.StateID == "" {
		state.StateID = generateUUID()
	}
	state.CreatedAt = time.Now()
	state.Status = "pending"

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	state.TenantID = tenantID

	if err := a.store.UpsertState(r.Context(), tenantID, &state); err != nil {
		log.Printf("Failed to create state: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Trigger reconciliation directly for now (Phase 3)
	// Note: In Phase 4/5 this should be handled by Scheduler picking up the change
	// or via event stream. But explicit call is fine for now.
	go a.reconciler.Reconcile(context.Background(), tenantID, state.StateID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(state)
}

func (a *API) handleGetState(w http.ResponseWriter, r *http.Request) {
	// Extract StateID from path /states/{state_id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid state ID", http.StatusBadRequest)
		return
	}
	stateID := pathParts[2]

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	state, err := a.store.GetState(r.Context(), tenantID, stateID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// tenantID := r.URL.Query().Get("tenant_id") // Deprecated
	states, err := a.store.ListStates(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

	// Storm Protection
	if !a.reconcileLimiter.Allow() {
		a.writeRateLimitError(w)
		return
	}

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if state exists
	state, err := a.store.GetState(r.Context(), tenantID, stateID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if state == nil {
		http.Error(w, "State not found", http.StatusNotFound)
		return
	}

	// Create Reconciliation Task
	task := &scheduler.ReconciliationTask{
		ReqID:    generateUUID(),
		NodeID:   state.NodeID,
		TenantID: tenantID, // Use actual tenantID
		Priority: 5,        // Default priority
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

// -- Phase 6: Incident Management --

func (a *API) handleCaptureIncident(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stateID := r.URL.Query().Get("state_id")
	if stateID == "" {
		http.Error(w, "state_id is required", http.StatusBadRequest)
		return
	}

	// Access timeline from scheduler (hacky but effective for Phase 6)
	// We need to expose GetTimeline() on Scheduler or just use what we have.
	// Scheduler struct exposes Timeline?
	// In scheduler.go: `timeline *timeline.Store` is unexported.
	// I need to export it or add a getter.
	// Or API should hold reference to Timeline Store directly?
	// API currently doesn't have timeline store field.
	// I will update Scheduler to expose Timeline via getter.

	tl := a.scheduler.GetTimeline()

	tenantID, err := middleware.GetTenantFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := incident.CaptureIncident(r.Context(), a.store, tl, tenantID, stateID)
	if err != nil {
		log.Printf("Failed to capture incident: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if report == nil {
		http.Error(w, "State not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=incident-%s.json", stateID))
	json.NewEncoder(w).Encode(report)
}
