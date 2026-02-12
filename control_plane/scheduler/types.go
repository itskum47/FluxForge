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
}

// SchedulerMode defines the operating mode of the scheduler.
type SchedulerMode string

const (
	ModeNormal   SchedulerMode = "NORMAL"
	ModeDegraded SchedulerMode = "DEGRADED"  // Reject low priority, shed load
	ModeReadOnly SchedulerMode = "READ_ONLY" // Accept no new tasks, process existing
	ModeDraining SchedulerMode = "DRAINING"  // Accept no new tasks, finish existing
)

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

// NodeHealth tracks the health status of a node for scheduling decisions.
type NodeHealth struct {
	NodeID string

	AgentReportedHealth float64
	ObservedFailureRate float64
	ExternalProbeScore  float64

	CompositeScore float64 // 0.2 * Agent + 0.5 * Observed + 0.3 * External

	BackoffDuration time.Duration
	Quarantined     bool
	LastFailure     time.Time
}

// CalculateCompositeScore updates the CompositeScore based on weighted inputs.
func (n *NodeHealth) CalculateCompositeScore() {
	n.CompositeScore = (0.2 * n.AgentReportedHealth) +
		(0.5 * n.ObservedFailureRate) +
		(0.3 * n.ExternalProbeScore)
}
