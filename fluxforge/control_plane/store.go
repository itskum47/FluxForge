package main

import (
	"sync"
)

// Agent represents a registered agent in the system.
type Agent struct {
	NodeID   string `json:"node_id"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	LastSeen int64  `json:"last_seen"`
	Address  string `json:"address"` // IP or Hostname reachable from CP
	Port     int    `json:"port"`    // HTTP port
}

// Job represents a remote execution task.
type Job struct {
	JobID     string `json:"job_id"`
	NodeID    string `json:"node_id"`
	Command   string `json:"command"`
	Status    string `json:"status"` // queued, running, completed, failed
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	CreatedAt int64  `json:"created_at"`
}

// Store holds the in-memory state of registered agents and jobs.
type Store struct {
	mu     sync.RWMutex
	agents map[string]*Agent
	jobs   map[string]*Job
}

// NewStore initializes a new Store.
func NewStore() *Store {
	return &Store{
		agents: make(map[string]*Agent),
		jobs:   make(map[string]*Job),
	}
}

// UpsertAgent adds or updates an agent in the store.
func (s *Store) UpsertAgent(a *Agent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[a.NodeID] = a
}

// UpdateHeartbeat updates the LastSeen timestamp for an agent.
// Returns true if the agent exists, false otherwise.
func (s *Store) UpdateHeartbeat(nodeID string, ts int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, ok := s.agents[nodeID]
	if !ok {
		return false
	}
	agent.LastSeen = ts
	return true
}

// ListAgents returns a thread-safe snapshot of all registered agents.
func (s *Store) ListAgents() []Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Agent, 0, len(s.agents))
	for _, a := range s.agents {
		// Append a value copy, dereferencing the pointer to avoid exposing internal state.
		result = append(result, *a)
	}
	return result
}

// UpsertJob adds or updates a job in the store.
func (s *Store) UpsertJob(j *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.JobID] = j
}

// GetJob retrieves a job by ID. s.mu.RLock() is used to ensure thread safety.
// Returns nil if not found. Returns a copy to prevent external mutation.
func (s *Store) GetJob(jobID string) *Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	j, ok := s.jobs[jobID]
	if !ok {
		return nil
	}
	// Return copy
	jobCopy := *j
	return &jobCopy
}

// GetAgent retrieves an agent by NodeID.
func (s *Store) GetAgent(nodeID string) *Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	a, ok := s.agents[nodeID]
	if !ok {
		return nil
	}
	// Return copy
	agentCopy := *a
	return &agentCopy
}
