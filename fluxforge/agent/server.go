package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Server is the agent's HTTP server.
type Server struct {
	cfg      *Config
	executor *Executor
	mu       sync.Mutex
	busy     bool
}

// NewServer creates a new Server.
func NewServer(cfg *Config, executor *Executor) *Server {
	return &Server{
		cfg:      cfg,
		executor: executor,
	}
}

// Start starts the HTTP server.
// It blocks until the server stops.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/execute", s.handleExecute)

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	log.Printf("Agent HTTP server listening on %s", addr)
	
	// Note: In a real app we'd handle graceful shutdown of the server itself.
	// For Phase 2, simple ListenAndServe in a goroutine (from main) is sufficient,
	// or blocking here if main spawns it.
	return http.ListenAndServe(addr, mux)
}

// handleExecute processes job execution requests.
func (s *Server) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JobID   string `json:"job_id"`
		Command string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Concurrency Control
	s.mu.Lock()
	if s.busy {
		s.mu.Unlock()
		http.Error(w, "Agent busy", http.StatusConflict) // 409
		return
	}
	s.busy = true
	s.mu.Unlock()

	// Accept the job
	w.WriteHeader(http.StatusAccepted) // 202

	// Execute asynchronously
	go func() {
		defer func() {
			s.mu.Lock()
			s.busy = false
			s.mu.Unlock()
		}()
		s.executor.Execute(req.JobID, req.Command)
	}()
}
