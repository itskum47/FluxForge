package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewStore()
	dispatcher := NewDispatcher(store)
	api := NewAPI(store, dispatcher)

	// Register handlers
	http.HandleFunc("/agent/register", api.handleRegister)
	http.HandleFunc("/agent/heartbeat", api.handleHeartbeat)
	http.HandleFunc("/agents", api.handleListAgents)

	// Job handlers
	http.HandleFunc("/jobs", api.handleSubmitJob)
	http.HandleFunc("/jobs/", api.handleGetJob) // Handles /jobs/{id}
	http.HandleFunc("/jobs/result", api.handleJobResult)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("FluxForge Control Plane listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
