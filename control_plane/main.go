package main

import (
	"encoding/json"
	"github.com/itskum47/FluxForge/control_plane/scheduler"

	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	store := NewStore()
	dispatcher := NewDispatcher(store)
	reconciler := NewReconciler(store, dispatcher)

	// Phase 4: Intelligent Scheduler
	sched := scheduler.NewScheduler(reconciler)
	ctx := context.Background()
	sched.Start(ctx)

	api := NewAPI(store, dispatcher, reconciler, sched)

	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/agent/register", api.handleRegister)
	http.HandleFunc("/agent/heartbeat", api.handleHeartbeat)
	http.HandleFunc("/agents", api.handleListAgents)

	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.handleListJobs(w, r)
			return
		}
		// Wrap with idempotency for POST
		api.withIdempotency(api.handleSubmitJob)(w, r)
	})
	http.HandleFunc("/jobs/", api.handleGetJob)
	http.HandleFunc("/jobs/result", api.handleJobResult)

	http.HandleFunc("/states", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.handleListStates(w, r)
			return
		}
		api.withIdempotency(api.handleCreateState)(w, r)
	})
	http.HandleFunc("/states/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost &&
			len(r.URL.Path) > len("/states/") &&
			r.URL.Path[len(r.URL.Path)-len("/reconcile"):] == "/reconcile" {
			api.withIdempotency(api.handleReconcileState)(w, r)
			return
		}
		if r.Method == http.MethodGet {
			api.handleGetState(w, r)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	})

	// Metrics Endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Debug Snapshot Endpoint
	http.HandleFunc("/scheduler/debug/snapshot", func(w http.ResponseWriter, r *http.Request) {
		snapshot := sched.GetSnapshot()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})

	log.Println("FluxForge Control Plane listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
