package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Reconciler handles desired state reconciliation.
type Reconciler struct {
	store      *Store
	dispatcher *Dispatcher

	// activeReconciles tracks which agents are currently being reconciled.
	// Key: NodeID, Value: true if busy
	activeReconciles map[string]bool
	mu               sync.Mutex
}

// NewReconciler creates a new Reconciler.
func NewReconciler(store *Store, dispatcher *Dispatcher) *Reconciler {
	return &Reconciler{
		store:            store,
		dispatcher:       dispatcher,
		activeReconciles: make(map[string]bool),
	}
}

// IsAgentBusy reports whether an agent is currently being reconciled.
// Read-only check used by the API layer.
func (r *Reconciler) IsAgentBusy(nodeID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.activeReconciles[nodeID]
}

// Reconcile runs the full reconciliation loop for a state.
func (r *Reconciler) Reconcile(stateID string) error {
	state := r.store.GetState(stateID)
	if state == nil {
		log.Printf("Reconcile failed: state %s not found", stateID)
		return fmt.Errorf("state not found")
	}

	// Enforce one reconciliation per agent
	if !r.acquireLock(state.NodeID) {
		log.Printf("Reconcile skipped: agent %s is busy", state.NodeID)
		// This is not a failure of the node, just concurrency limit.
		return nil
	}
	defer r.releaseLock(state.NodeID)

	log.Printf("Starting reconciliation for state %s (node %s)", stateID, state.NodeID)

	agent := r.store.GetAgent(state.NodeID)
	if agent == nil {
		r.updateStatus(state, "failed", "agent not found")
		return fmt.Errorf("agent not found")
	}

	// 1. Check phase
	if !r.runCheck(agent, state) {
		return fmt.Errorf("check phase failed")
	}

	// 2. Apply phase
	if !r.runApply(agent, state) {
		return fmt.Errorf("apply phase failed")
	}

	// 3. Final check
	r.runFinalCheck(agent, state)
	if state.Status == "failed" {
		return fmt.Errorf("final check failed")
	}

	return nil
}

// acquireLock enforces per-agent exclusivity.
func (r *Reconciler) acquireLock(nodeID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.activeReconciles[nodeID] {
		return false
	}
	r.activeReconciles[nodeID] = true
	return true
}

// releaseLock releases the per-agent lock.
func (r *Reconciler) releaseLock(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.activeReconciles, nodeID)
}

// runCheck executes the check command.
func (r *Reconciler) runCheck(agent *Agent, state *State) bool {
	exitCode, err := r.executeJob(agent, state.CheckCmd)
	if err != nil {
		r.updateStatus(state, "failed", fmt.Sprintf("check failed: %v", err))
		return false
	}

	state.LastChecked = time.Now().Unix()

	if exitCode == state.DesiredExitCode {
		r.updateStatus(state, "compliant", "")
		return false
	}

	r.updateStatus(
		state,
		"drifted",
		fmt.Sprintf("exit code %d (expected %d)", exitCode, state.DesiredExitCode),
	)
	return true
}

// runApply executes the apply command.
func (r *Reconciler) runApply(agent *Agent, state *State) bool {
	r.updateStatus(state, "applying", "")

	exitCode, err := r.executeJob(agent, state.ApplyCmd)
	if err != nil {
		r.updateStatus(state, "failed", fmt.Sprintf("apply failed: %v", err))
		return false
	}

	if exitCode != 0 {
		log.Printf("Apply command returned exit code %d for state %s", exitCode, state.StateID)
	}

	return true
}

// runFinalCheck executes the final verification check.
func (r *Reconciler) runFinalCheck(agent *Agent, state *State) {
	exitCode, err := r.executeJob(agent, state.CheckCmd)
	if err != nil {
		r.updateStatus(state, "failed", fmt.Sprintf("final check failed: %v", err))
		return
	}

	state.LastChecked = time.Now().Unix()

	if exitCode == state.DesiredExitCode {
		r.updateStatus(state, "compliant", "")
	} else {
		r.updateStatus(
			state,
			"failed",
			fmt.Sprintf("drift persisted (exit code %d)", exitCode),
		)
	}
}

// executeJob creates a job, dispatches it, and waits for completion.
func (r *Reconciler) executeJob(agent *Agent, command string) (int, error) {
	jobID := generateUUID()

	job := &Job{
		JobID:     jobID,
		NodeID:    agent.NodeID,
		Command:   command,
		Status:    "queued",
		CreatedAt: time.Now().Unix(),
	}

	r.store.UpsertJob(job)

	log.Printf("Dispatching job %s to agent %s: %s", jobID, agent.NodeID, command)

	// IMPORTANT:
	// DispatchJob is async and returns no error.
	// Job state is the source of truth.
	r.dispatcher.DispatchJob(agent, job)

	return r.waitForJob(jobID)
}

// waitForJob polls until the job completes or fails.
func (r *Reconciler) waitForJob(jobID string) (int, error) {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return -1, fmt.Errorf("timeout waiting for job %s", jobID)

		case <-ticker.C:
			job := r.store.GetJob(jobID)
			if job == nil {
				return -1, fmt.Errorf("job %s lost", jobID)
			}

			switch job.Status {
			case "completed":
				return job.ExitCode, nil
			case "failed":
				return -1, fmt.Errorf("job execution failed: %s", job.Stderr)
			}
		}
	}
}

// updateStatus mutates and persists state status.
func (r *Reconciler) updateStatus(state *State, status, lastError string) {
	state.Status = status
	state.LastError = lastError
	r.store.UpsertState(state)
	log.Printf("State %s transitioned to %s", state.StateID, status)
}
