package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/itskum47/FluxForge/control_plane/store"
)

// Dispatcher is responsible for sending jobs to agents.
type Dispatcher struct {
	store store.Store
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(store store.Store) *Dispatcher {
	return &Dispatcher{store: store}
}

// DispatchJob sends a job to the target agent for execution.
// IMPORTANT:
// - HTTP 202 Accepted = success (async execution)
// - Job completion is reported later via /jobs/result
func (d *Dispatcher) DispatchJob(ctx context.Context, agent *store.Agent, job *store.Job) {
	// Check context before starting
	if ctx.Err() != nil {
		log.Printf("DispatchJob skipped: context cancelled (%v)", ctx.Err())
		d.store.UpdateJobStatus(context.Background(), job.JobID, "failed", 0, "", "dispatch cancelled: leadership lost")
		return
	}

	url := fmt.Sprintf("http://%s:%d/execute", agent.IPAddress, agent.Port)

	payload := map[string]string{
		"job_id":  job.JobID,
		"command": job.Command,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		// Use UpdateJobStatus interface method
		d.store.UpdateJobStatus(context.Background(), job.JobID, "failed", 0, "", fmt.Sprintf("failed to marshal payload: %v", err))
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		d.store.UpdateJobStatus(context.Background(), job.JobID, "failed", 0, "", fmt.Sprintf("failed to create request: %v", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute Request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		d.store.UpdateJobStatus(context.Background(), job.JobID, "failed", 0, "", fmt.Sprintf("failed to contact agent: %v", err))
		return
	}
	defer resp.Body.Close()

	// ✅ CORRECT SEMANTICS
	if resp.StatusCode != http.StatusAccepted {
		d.store.UpdateJobStatus(context.Background(), job.JobID, "failed", 0, "", fmt.Sprintf("agent returned status %d", resp.StatusCode))
		return
	}

	// Job accepted for execution
	d.store.UpdateJobStatus(context.Background(), job.JobID, "running", 0, "", "")

	log.Printf("Job %s dispatched to agent %s", job.JobID, agent.NodeID)
}
