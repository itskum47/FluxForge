package main

import (
	"bytes"
	"encoding/json"
	"github.com/itskum47/FluxForge/control_plane/scheduler"
	"net/http"
	"net/http/httptest"
	"testing"
)

// -- Phase 1: Agent Lifecycle Regression --
func TestRegression_AgentLifecycle(t *testing.T) {
	// Setup Control Plane
	store := NewStore()
	dispatcher := NewDispatcher(store)
	reconciler := NewReconciler(store, dispatcher)
	sched := scheduler.NewScheduler(reconciler)
	api := NewAPI(store, dispatcher, reconciler, sched)

	// 1. Register Agent
	agent := Agent{
		NodeID:   "reg-node-1",
		Hostname: "regression-host",
		Address:  "10.0.0.1",
		Port:     8080,
	}
	body, _ := json.Marshal(agent)
	req := httptest.NewRequest("POST", "/agent/register", bytes.NewBuffer(body))
	req.Header.Set("X-Agent-Signature", "sig-123") // Phase 4 requirement
	w := httptest.NewRecorder()
	api.handleRegister(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Registration failed: %d", w.Code)
	}

	// 2. Verify Listing
	req = httptest.NewRequest("GET", "/agents", nil)
	w = httptest.NewRecorder()
	api.handleListAgents(w, req)

	var agents []Agent
	json.Unmarshal(w.Body.Bytes(), &agents)
	if len(agents) != 1 {
		t.Error("Expected 1 agent in list")
	}
	if agents[0].NodeID != "reg-node-1" {
		t.Errorf("Expected reg-node-1, got %s", agents[0].NodeID)
	}
}

// -- Phase 2 & 3: Reconciliation & Execution --

type MockDispatcher struct {
	dispatched []Job
}

func (m *MockDispatcher) DispatchJob(agent *Agent, job *Job) {
	m.dispatched = append(m.dispatched, *job)
	// Simulate async completion
	job.Status = "running"
}

func TestRegression_ReconciliationLoop(t *testing.T) {
	store := NewStore()
	// Mock Dispatcher to verify remote execution call
	dispatcher := NewDispatcher(store) // using real dispatcher for regression but without network

	reconciler := NewReconciler(store, dispatcher)
	sched := scheduler.NewScheduler(reconciler)
	api := NewAPI(store, dispatcher, reconciler, sched)

	// Start Scheduler
	// (In unit test we might need to tick it manually or rely on Submit)

	// 1. Register Agent
	store.UpsertAgent(&Agent{NodeID: "reg-node-1", Address: "10.0.0.1"})

	// 2. Create Desired State
	stateReq := map[string]interface{}{
		"node_id":           "reg-node-1",
		"check_cmd":         "check_nginx",
		"apply_cmd":         "start_nginx",
		"desired_exit_code": 0,
	}
	body, _ := json.Marshal(stateReq)
	req := httptest.NewRequest("POST", "/states", bytes.NewBuffer(body))
	// Idempotency Key (Phase 4 requirement)
	req.Header.Set("X-Flux-Idempotency-Key", "idemp-1")
	w := httptest.NewRecorder()

	// Wrap with Middleware manually or call handler?
	// The API struct methods don't have middleware inside them (it's in main.go).
	// So we call handleCreateState directly.
	api.handleCreateState(w, req)

	var stateResp map[string]string
	json.Unmarshal(w.Body.Bytes(), &stateResp)
	stateID := stateResp["state_id"]

	// 3. Trigger Reconcile
	req = httptest.NewRequest("POST", "/states/"+stateID+"/reconcile", nil)
	req.Header.Set("X-Flux-Idempotency-Key", "idemp-2")
	w = httptest.NewRecorder()
	api.handleReconcileState(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Reconcile trigger failed: %d", w.Code)
	}

	// 4. Verify Task in Scheduler Queue
	snap := sched.GetSnapshot()
	if snap["queue_depth"].(int) != 1 {
		t.Errorf("Expected 1 task in queue, got %d", snap["queue_depth"])
	}
}
