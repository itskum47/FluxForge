package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"syscall"
)

// Executor handles job execution on the agent.
type Executor struct {
	cfg *Config
}

// NewExecutor creates a new Executor.
func NewExecutor(cfg *Config) *Executor {
	return &Executor{cfg: cfg}
}

// Execute runs a command and reports the result.
func (e *Executor) Execute(jobID, command string) {
	log.Printf("Executing job %s: %s", jobID, command)

	// Phase 2 targets Unix-like systems.
	// Windows support would require different shell handling.
	cmd := exec.Command("sh", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	status := "completed"

	err := cmd.Run()
	if err != nil {
		status = "failed"

		// Extract exit code if possible
		if exitErr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = waitStatus.ExitStatus()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 1
			fmt.Fprintf(&stderr, "\nExecution error: %v", err)
		}
	}

	e.sendResult(jobID, stdout.String(), stderr.String(), exitCode, status)
}

func (e *Executor) sendResult(jobID, stdout, stderr string, exitCode int, status string) {
	payload := map[string]interface{}{
		"job_id":    jobID,
		"node_id":   e.cfg.NodeID,
		"stdout":    stdout,
		"stderr":    stderr,
		"exit_code": exitCode,
		"status":    status,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal result for job %s: %v", jobID, err)
		return
	}

	resp, err := http.Post(
		e.cfg.ServerURL+"/jobs/result",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		log.Printf("Failed to send result for job %s: %v", jobID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Control plane rejected result for job %s: %d", jobID, resp.StatusCode)
	} else {
		log.Printf("Result sent for job %s (status: %s)", jobID, status)
	}
}
