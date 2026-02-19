package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRaceCondition_AgentStateRace tests concurrent agent state changes
func TestRaceCondition_AgentStateRace(t *testing.T) {
	t.Log("Testing race condition: agent heartbeat vs timeout detection")

	const iterations = 100
	var violations atomic.Int32

	for i := 0; i < iterations; i++ {
		var wg sync.WaitGroup
		wg.Add(4)

		// Simulate simultaneous operations
		go func() {
			defer wg.Done()
			// Kill agent
			_ = killAgent("race-agent-1")
		}()

		go func() {
			defer wg.Done()
			// Send heartbeat
			_ = sendHeartbeat("race-agent-1")
		}()

		go func() {
			defer wg.Done()
			// Dispatch job
			_ = dispatchJob("race-agent-1", "echo test")
		}()

		go func() {
			defer wg.Done()
			// Restart agent
			_ = startAgent("race-agent-1")
		}()

		wg.Wait()

		// Verify invariant: agent cannot be both dead AND alive
		if err := checkAgentStateInvariant("race-agent-1"); err != nil {
			violations.Add(1)
			t.Errorf("Iteration %d: %v", i, err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	if violations.Load() > 0 {
		t.Fatalf("RACE CONDITION DETECTED: %d violations in %d iterations", violations.Load(), iterations)
	}
}

// TestRaceCondition_JobAssignment tests duplicate job assignment
func TestRaceCondition_JobAssignment(t *testing.T) {
	t.Log("Testing race condition: duplicate job assignment")

	jobID := "race-job-1"
	var wg sync.WaitGroup
	wg.Add(2)

	// Two schedulers try to assign same job
	go func() {
		defer wg.Done()
		_ = assignJobToAgent(jobID, "agent-1")
	}()

	go func() {
		defer wg.Done()
		_ = assignJobToAgent(jobID, "agent-2")
	}()

	wg.Wait()

	// Verify invariant: job assigned to exactly one agent
	agentCount, err := getJobAssignmentCount(jobID)
	if err != nil {
		t.Fatal(err)
	}

	if agentCount != 1 {
		t.Fatalf("DUPLICATE ASSIGNMENT: job assigned to %d agents (expected 1)", agentCount)
	}
}

// TestLivelock_SchedulerStuck tests scheduler livelock detection
func TestLivelock_SchedulerStuck(t *testing.T) {
	t.Log("Testing livelock detection: scheduler stuck")

	detector := NewLivelockDetector()

	// Submit 10 jobs
	for i := 0; i < 10; i++ {
		submitJob(fmt.Sprintf("livelock-job-%d", i))
	}

	// Monitor for 2 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Test passed - no livelock detected
			return

		case <-ticker.C:
			metrics := getDashboardMetrics()

			if err := detector.Check(metrics.QueueDepth, metrics.ActiveTasks); err != nil {
				t.Fatalf("LIVELOCK DETECTED: %v", err)
			}

			// If queue drained, test passed
			if metrics.QueueDepth == 0 {
				t.Log("Queue drained successfully - no livelock")
				return
			}
		}
	}
}

// TestMaliciousAgent_DishonestReport tests dishonest agent detection
func TestMaliciousAgent_DishonestReport(t *testing.T) {
	t.Log("Testing malicious agent: dishonest success report")

	jobID := "malicious-job-1"

	// Submit job
	submitJob(jobID)

	// Agent reports success without executing
	err := reportJobResult(jobID, "success", "fake-result", "")

	// System should reject dishonest report
	if err == nil {
		t.Fatal("SECURITY VIOLATION: system accepted dishonest report")
	}

	t.Logf("Correctly rejected dishonest report: %v", err)
}

// TestMaliciousAgent_DuplicateReport tests duplicate result rejection
func TestMaliciousAgent_DuplicateReport(t *testing.T) {
	t.Log("Testing malicious agent: duplicate result report")

	jobID := "duplicate-job-1"

	// Submit and execute job
	submitJob(jobID)
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup
	wg.Add(2)

	var err1, err2 error

	// Send duplicate results simultaneously
	go func() {
		defer wg.Done()
		err1 = reportJobResult(jobID, "success", "result-1", "token-1")
	}()

	go func() {
		defer wg.Done()
		err2 = reportJobResult(jobID, "success", "result-2", "token-1")
	}()

	wg.Wait()

	// Exactly one should succeed
	if err1 == nil && err2 == nil {
		t.Fatal("DUPLICATE ACCEPTED: both results accepted")
	}

	// Verify only one result in DB
	count, err := getJobResultCount(jobID)
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("DUPLICATE STORED: found %d results (expected 1)", count)
	}
}

// TestCrashRecovery_JobAssignment tests crash during job assignment
func TestCrashRecovery_JobAssignment(t *testing.T) {
	t.Log("Testing crash recovery: job assignment")

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	// Assign job
	_, err = tx.Exec("INSERT INTO job_assignments (job_id, agent_id, status) VALUES ($1, $2, $3)",
		"crash-job-1", "crash-agent-1", "assigned")
	if err != nil {
		t.Fatal(err)
	}

	// Simulate crash (rollback instead of commit)
	tx.Rollback()

	// Restart system
	restartControlPlane()

	// Verify no duplicate jobs
	count, err := getJobAssignmentCount("crash-job-1")
	if err != nil {
		t.Fatal(err)
	}

	if count > 1 {
		t.Fatalf("DUPLICATE AFTER CRASH: found %d assignments", count)
	}

	// Verify job is either assigned or pending (not both)
	status, err := getJobStatus("crash-job-1")
	if err != nil {
		t.Fatal(err)
	}

	if status != "pending" && status != "assigned" {
		t.Fatalf("INVALID STATE AFTER CRASH: job status is %s", status)
	}
}

// TestIdempotency_Reconciliation tests reconciliation idempotency
func TestIdempotency_Reconciliation(t *testing.T) {
	t.Log("Testing idempotency: reconciliation")

	stateID := "idempotent-state-1"

	// Submit state
	submitDesiredState(stateID, "true")

	// Wait for applying
	time.Sleep(3 * time.Second)

	// Kill and restart 3 times
	for i := 0; i < 3; i++ {
		killControlPlane()
		time.Sleep(1 * time.Second)
		startControlPlane()
		time.Sleep(5 * time.Second)
	}

	// Verify state reached compliant exactly once
	count, err := getStateTransitionCount(stateID, "compliant")
	if err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("NOT IDEMPOTENT: state transitioned to compliant %d times (expected 1)", count)
	}
}

// TestSustainedChaos runs sustained load test
func TestSustainedChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained chaos test in short mode")
	}

	t.Log("Testing sustained chaos: 5 states/sec for 30 minutes")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stateCount := 0
	startMemory := getMemoryUsage()
	startGoroutines := getGoroutineCount()

	for {
		select {
		case <-ctx.Done():
			t.Logf("Sustained chaos complete: %d states submitted", stateCount)

			// Verify no resource leaks
			endMemory := getMemoryUsage()
			endGoroutines := getGoroutineCount()

			if endMemory > startMemory*2 {
				t.Fatalf("MEMORY LEAK: grew from %d to %d bytes", startMemory, endMemory)
			}

			if endGoroutines > startGoroutines+1000 {
				t.Fatalf("GOROUTINE LEAK: grew from %d to %d", startGoroutines, endGoroutines)
			}

			return

		case <-ticker.C:
			// Submit 5 states
			for i := 0; i < 5; i++ {
				submitDesiredState(fmt.Sprintf("sustained-%d-%d", time.Now().Unix(), i), "true")
				stateCount++
			}

			// Check metrics every minute
			if stateCount%300 == 0 {
				metrics := getDashboardMetrics()
				t.Logf("Progress: %d states, queue: %d, memory: %d MB, goroutines: %d",
					stateCount, metrics.QueueDepth, getMemoryUsage()/1024/1024, getGoroutineCount())
			}
		}
	}
}

// TestSchedulerFairness tests that short jobs aren't starved
func TestSchedulerFairness(t *testing.T) {
	t.Log("Testing scheduler fairness")

	// Submit 1 long job
	longJobID := "long-job-1"
	submitJob(longJobID, "sleep 300")
	longStart := time.Now()

	time.Sleep(1 * time.Second)

	// Submit 100 short jobs
	shortJobIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		shortJobIDs[i] = fmt.Sprintf("short-job-%d", i)
		submitJob(shortJobIDs[i], "echo fast")
	}
	shortStart := time.Now()

	// Wait for short jobs to complete
	for _, jobID := range shortJobIDs {
		waitForJobCompletion(jobID, 60*time.Second)
	}
	shortEnd := time.Now()

	// Verify short jobs completed before long job
	longStatus, _ := getJobStatus(longJobID)
	if longStatus == "completed" {
		longEnd := getJobCompletionTime(longJobID)
		if longEnd.Before(shortEnd) {
			t.Fatal("FAIRNESS VIOLATION: long job completed before short jobs")
		}
	}

	t.Logf("Fairness verified: short jobs completed in %v", shortEnd.Sub(shortStart))
}

// LivelockDetector detects scheduler livelocks
type LivelockDetector struct {
	lastQueueDepth int
	lastCheck      time.Time
	mu             sync.Mutex
}

func NewLivelockDetector() *LivelockDetector {
	return &LivelockDetector{
		lastCheck: time.Now(),
	}
}

func (d *LivelockDetector) Check(queueDepth, activeTasks int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if queueDepth > 0 && activeTasks > 0 {
		if queueDepth == d.lastQueueDepth {
			if time.Since(d.lastCheck) > 60*time.Second {
				return fmt.Errorf("LIVELOCK: queue stuck at %d for 60s with %d active tasks", queueDepth, activeTasks)
			}
		} else {
			d.lastQueueDepth = queueDepth
			d.lastCheck = time.Now()
		}
	} else {
		d.lastQueueDepth = queueDepth
		d.lastCheck = time.Now()
	}

	return nil
}

// System Invariants
type Invariant struct {
	Name  string
	Check func(*sql.DB) error
}

var SystemInvariants = []Invariant{
	{
		Name: "Job cannot be both completed and running",
		Check: func(db *sql.DB) error {
			var count int
			err := db.QueryRow(`
				SELECT COUNT(*) FROM jobs 
				WHERE status IN ('completed', 'failed') 
				AND id IN (SELECT job_id FROM active_jobs)
			`).Scan(&count)
			if err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("found %d jobs both completed and running", count)
			}
			return nil
		},
	},
	{
		Name: "Dead agent cannot execute job",
		Check: func(db *sql.DB) error {
			var count int
			err := db.QueryRow(`
				SELECT COUNT(*) FROM job_assignments ja
				JOIN agents a ON ja.agent_id = a.node_id
				WHERE a.status = 'dead'
				AND ja.status = 'running'
			`).Scan(&count)
			if err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("found %d jobs running on dead agents", count)
			}
			return nil
		},
	},
	{
		Name: "Only one leader exists",
		Check: func(db *sql.DB) error {
			var count int
			err := db.QueryRow(`SELECT COUNT(*) FROM nodes WHERE is_leader = true`).Scan(&count)
			if err != nil {
				return err
			}
			if count != 1 {
				return fmt.Errorf("found %d leaders (expected 1)", count)
			}
			return nil
		},
	},
}

// RunInvariantChecker continuously verifies system invariants
func RunInvariantChecker(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, inv := range SystemInvariants {
				if err := inv.Check(db); err != nil {
					log.Fatalf("INVARIANT VIOLATION [%s]: %v", inv.Name, err)
				}
			}
		}
	}
}

// Helper functions (to be implemented)
func checkAgentStateInvariant(agentID string) error   { return nil }
func killAgent(agentID string) error                  { return nil }
func sendHeartbeat(agentID string) error              { return nil }
func dispatchJob(agentID, cmd string) error           { return nil }
func startAgent(agentID string) error                 { return nil }
func assignJobToAgent(jobID, agentID string) error    { return nil }
func getJobAssignmentCount(jobID string) (int, error) { return 0, nil }
func submitJob(jobID string, cmd ...string) error     { return nil }
func getDashboardMetrics() struct{ QueueDepth, ActiveTasks int } {
	return struct{ QueueDepth, ActiveTasks int }{}
}
func reportJobResult(jobID, status, result, token string) error      { return nil }
func getJobResultCount(jobID string) (int, error)                    { return 0, nil }
func restartControlPlane()                                           {}
func getJobStatus(jobID string) (string, error)                      { return "", nil }
func submitDesiredState(stateID, checkCmd string) error              { return nil }
func killControlPlane()                                              {}
func startControlPlane()                                             {}
func getStateTransitionCount(stateID, toStatus string) (int, error)  { return 0, nil }
func getMemoryUsage() int64                                          { return 0 }
func getGoroutineCount() int                                         { return 0 }
func waitForJobCompletion(jobID string, timeout time.Duration) error { return nil }
func getJobCompletionTime(jobID string) time.Time                    { return time.Now() }
