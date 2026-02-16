package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/itskum47/FluxForge/control_plane/idempotency"
	"github.com/itskum47/FluxForge/control_plane/scheduler"
	"github.com/itskum47/FluxForge/control_plane/store"
	"github.com/itskum47/FluxForge/control_plane/streaming"
)

// TestLoadSimulation_10kAgents simulates 10k agents hitting heartbeat endpoint
// to verify Rate Limiting/Storm Protection stability.
func TestLoadSimulation_10kAgents(t *testing.T) {
	// 1. Setup Control Plane
	s := store.NewMemoryStore()
	dispatcher := NewDispatcher(s)
	publisher := streaming.NewLogPublisher() // Mock/Log

	// Create Scheduler
	reconciler := NewReconciler(s, dispatcher, publisher)
	schedConfig := scheduler.DefaultSchedulerConfig()
	sched := scheduler.NewScheduler(s, reconciler, 0, 1, schedConfig) // Shard 0/1

	// Start Scheduler (noop for this test mostly, but needed for submission)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Start(ctx)

	api := NewAPI(s, dispatcher, reconciler, sched, idempotency.NewStore(nil))

	// 2. Create Test Server
	server := httptest.NewServer(http.HandlerFunc(api.handleHeartbeat))
	defer server.Close()

	// 3. Flood Simulation
	const numBatches = 100
	const batchSize = 100
	totalReqs := numBatches * batchSize

	// Pre-register agents to avoid 404s
	fmt.Println("Pre-registering agents...")
	for b := 0; b < numBatches; b++ {
		for i := 0; i < batchSize; i++ {
			nodeID := fmt.Sprintf("agent-%d-%d", b, i)
			s.UpsertAgent(context.Background(), &store.Agent{
				NodeID:        nodeID,
				Status:        "active",
				LastHeartbeat: time.Now(),
			})
		}
	}
	fmt.Println("Agents registered. Starting flood.")

	var successCount int64
	var rateLimitedCount int64
	var errorCount int64

	client := server.Client() // persistent connection

	var wg sync.WaitGroup
	start := time.Now()

	// 100 batches of 100 agents
	for batch := 0; batch < numBatches; batch++ {
		wg.Add(batchSize)
		for i := 0; i < batchSize; i++ {
			go func(b, id int) {
				defer wg.Done()
				nodeID := fmt.Sprintf("agent-%d-%d", b, id)

				body := fmt.Sprintf(`{"node_id": "%s"}`, nodeID)
				resp, err := client.Post(server.URL, "application/json", strings.NewReader(body))
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else if resp.StatusCode == http.StatusTooManyRequests {
					atomic.AddInt64(&rateLimitedCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}(batch, i)
		}
		// Small sleep between batches to simulate ramp-up?
		// No, let's hammer it to test storm protection.
	}

	wg.Wait()
	duration := time.Since(start)

	t.Logf("Computed %d requests in %v", totalReqs, duration)
	t.Logf("Success: %d, RateLimited: %d, Errors: %d", successCount, rateLimitedCount, errorCount)

	// Verification
	// We expect SOME success (buckets start full)
	// We expect MAINLY RateLimits (bucket is 100/sec, we did 10k in <1s probably)

	if rateLimitedCount == 0 {
		t.Error("Expected rate limiting to kick in, but got 0 429s")
	}
	if successCount == 0 {
		t.Error("Expected at least some requests to succeed")
	}
}
