package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Dispatcher is responsible for sending jobs to agents.
type Dispatcher struct {
	store *Store
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(store *Store) *Dispatcher {
	return &Dispatcher{store: store}
}

// DispatchJob sends a job to the target agent for execution.
// IMPORTANT:
// - HTTP 202 Accepted = success (async execution)
// - Job completion is reported later via /jobs/result
func (d *Dispatcher) DispatchJob(agent *Agent, job *Job) {
	url := fmt.Sprintf("http://%s:%d/execute", agent.Address, agent.Port)

	payload := map[string]string{
		"job_id":  job.JobID,
		"command": job.Command,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		job.Status = "failed"
		job.Stderr = fmt.Sprintf("failed to marshal payload: %v", err)
		d.store.UpsertJob(job)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		job.Status = "failed"
		job.Stderr = fmt.Sprintf("failed to contact agent: %v", err)
		d.store.UpsertJob(job)
		return
	}
	defer resp.Body.Close()

	// ✅ CORRECT SEMANTICS
	if resp.StatusCode != http.StatusAccepted {
		job.Status = "failed"
		job.Stderr = fmt.Sprintf("agent returned status %d", resp.StatusCode)
		d.store.UpsertJob(job)
		return
	}

	// Job accepted for execution
	job.Status = "running"
	d.store.UpsertJob(job)

	log.Printf("Job %s dispatched to agent %s", job.JobID, agent.NodeID)
}
