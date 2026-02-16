package scheduler

import (
	"time"
)

// TaskCost represents the estimated resource cost of a task.
type TaskCost struct {
	CPUSeconds float64
	IOOps      int
	NetMB      float64
}

// ReconciliationTask represents a unit of work for the scheduler.
type ReconciliationTask struct {
	ReqID         string
	NodeID        string
	TenantID      string
	Priority      int // 0 (Critical) to 10 (Background)
	Deadline      time.Time
	Attempt       int
	Cost          TaskCost
	FailureDomain string // zone / region / rack
	StateID       string // The ID of the state to reconcile
	TraceContext  map[string]string
	SubmitTime    time.Time // For priority aging
	EnqueuedAt    time.Time // For backpressure telemetry (admission wait time)
}

// SchedulerMode defines the operating mode of the scheduler.
type SchedulerMode string

const (
	ModeNormal   SchedulerMode = "NORMAL"
	ModeDegraded SchedulerMode = "DEGRADED"  // Reject low priority, shed load
	ModeReadOnly SchedulerMode = "READ_ONLY" // Accept no new tasks, process existing
	ModeDraining SchedulerMode = "DRAINING"  // Accept no new tasks, finish existing
)

// AdmissionMode controls ingress traffic (Pilot Kill Switch)
type AdmissionMode int

const (
	AdmissionNormal AdmissionMode = iota
	AdmissionDrain                // Finish running, reject new
	AdmissionFreeze               // Reject everything immediately
)

// SchedulerConfig holds configuration for the scheduler.
type SchedulerConfig struct {
	// MaxTaskExecutionTime is the hard timeout for any single task
	// After this duration, the task context is forcibly cancelled
	// This prevents worker goroutine leaks from hung agents or infinite loops
	MaxTaskExecutionTime time.Duration // Default: 5 minutes

	// MaxConcurrency is the maximum number of concurrent workers
	MaxConcurrency int // Default: 10

	// CircuitBreakerThreshold is the queue depth that triggers circuit open
	CircuitBreakerThreshold int // Default: 1000
}

// DefaultSchedulerConfig returns sensible production defaults.
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		MaxTaskExecutionTime:    5 * time.Minute,
		MaxConcurrency:          10,
		CircuitBreakerThreshold: 1000,
	}
}

// SchedulingDecision represents a structured log entry for scheduler actions.
type SchedulingDecision struct {
	Component string      `json:"component"`
	Decision  string      `json:"decision"` // DISPATCH, RATE_LIMIT_DELAY, QUARANTINE_DROP, DOMAIN_THROTTLE
	ReqID     string      `json:"req_id"`
	TenantID  string      `json:"tenant_id"`
	NodeID    string      `json:"node_id"`
	Priority  int         `json:"priority"`
	DelayMS   int64       `json:"delay_ms,omitempty"`
	Reason    string      `json:"reason,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// NodeHealth tracks the health/status of an agent from the scheduler's perspective.
type NodeHealth struct {
	NodeID string

	// Signals
	AgentReportedHealth float64
	ObservedFailureRate float64
	ExternalProbeScore  float64

	// Derived
	CompositeScore  float64
	Quarantined     bool
	BackoffDuration time.Duration

	// Metadata
	LastSeen time.Time
	Tier     string // normal, canary
}

// CalculateCompositeScore updates the CompositeScore based on weighted inputs.
func (n *NodeHealth) CalculateCompositeScore() {
	n.CompositeScore = (0.2 * n.AgentReportedHealth) +
		(0.5 * n.ObservedFailureRate) +
		(0.3 * n.ExternalProbeScore)
}

// String returns the string representation of AdmissionMode.
func (m AdmissionMode) String() string {
	switch m {
	case AdmissionNormal:
		return "Normal"
	case AdmissionDrain:
		return "Drain"
	case AdmissionFreeze:
		return "Freeze"
	default:
		return "Unknown"
	}
}

// SchedulerMetrics exposes internal state for the dashboard.
type SchedulerMetrics struct {
	QueueDepth          int     `json:"queue_depth"`
	ActiveTasks         int     `json:"active_tasks"`
	MaxConcurrency      int     `json:"max_concurrency"`
	WorkerSaturation    float64 `json:"worker_saturation"`
	CircuitBreakerState string  `json:"circuit_breaker_state"`
	AdmissionMode       string  `json:"admission_mode"`
	RuntimeMode         string  `json:"runtime_mode"`
}
