package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Dispatcher handles job dispatching to agents.
type Dispatcher struct {
	store *Store
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(store *Store) *Dispatcher {
	return &Dispatcher{store: store}
}

// DispatchJob sends a job to the target agent.
// It waits for the agent to accept the job (HTTP 202).
// It updates the job status to "running" or "failed".
func (d *Dispatcher) DispatchJob(agent *Agent, job *Job) {
	// Construct payload
	payload := map[string]string{
		"job_id":  job.JobID,
		"command": job.Command,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		d.failJob(job, fmt.Sprintf("Failed to marshal job payload: %v", err))
		return
	}

	// Construct URL
	// Note: Agent is responsible for providing a valid address/port.
	url := fmt.Sprintf("http://%s:%d/execute", agent.Address, agent.Port)

	// Send request with short timeout for acceptance
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		// Dispatch failure: Agent unreachable or network error
		d.failJob(job, fmt.Sprintf("Failed to connect to agent: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		d.failJob(job, "Agent busy (Concurrency limit reached)")
		return
	}

	if resp.StatusCode != http.StatusAccepted {
		d.failJob(job, fmt.Sprintf("Agent rejected job with status: %d", resp.StatusCode))
		return
	}

	// Success: Mark as running
	d.updateJobStatus(job, "running")
	log.Printf("Job %s dispatched to %s (%s)", job.JobID, agent.NodeID, url)
}

func (d *Dispatcher) failJob(job *Job, reason string) {
	log.Printf("Job %s dispatch failed: %s", job.JobID, reason)
	job.Status = "failed"
	job.Stderr = reason // Record dispatch error in stderr for visibility
	d.store.UpsertJob(job)
}

func (d *Dispatcher) updateJobStatus(job *Job, status string) {
	job.Status = status
	d.store.UpsertJob(job)
}
