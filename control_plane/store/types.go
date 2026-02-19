package store

import (
	"time"
)

// Agent represents a registered execution node.
type Agent struct {
	NodeID        string            `json:"node_id" db:"node_id"`
	TenantID      string            `json:"tenant_id" db:"tenant_id"` // Multi-tenancy
	Hostname      string            `json:"hostname" db:"hostname"`
	IPAddress     string            `json:"ip_address" db:"ip_address"`
	Version       string            `json:"version" db:"version"`
	Status        string            `json:"status" db:"status"` // "active", "offline", "quarantined"
	LastHeartbeat time.Time         `json:"last_heartbeat" db:"last_heartbeat_at"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
	Metadata      map[string]string `json:"metadata" db:"metadata"` // JSONB in Postgres
	Port          int               `json:"port" db:"port"`
	Tier          string            `json:"tier" db:"tier"` // "standard", "premium", "dedicated"
}

// Job represents an execution task history.
type Job struct {
	JobID      string     `json:"job_id" db:"job_id"`
	NodeID     string     `json:"node_id" db:"node_id"`
	TenantID   string     `json:"tenant_id" db:"tenant_id"` // Multi-tenancy
	StateID    string     `json:"state_id" db:"state_id"`
	Command    string     `json:"command" db:"command"`
	Status     string     `json:"status" db:"status"` // "queued", "running", "completed", "failed"
	ExitCode   int        `json:"exit_code" db:"exit_code"`
	Stdout     string     `json:"stdout" db:"stdout"`
	Stderr     string     `json:"stderr" db:"stderr"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	StartedAt  *time.Time `json:"started_at" db:"started_at"`
	FinishedAt *time.Time `json:"finished_at" db:"finished_at"`
	TraceID    string     `json:"trace_id" db:"trace_id"`
}

// DesiredState represents a target configuration for a node.
type DesiredState struct {
	StateID         string    `json:"state_id" db:"state_id"`
	NodeID          string    `json:"node_id" db:"node_id"`
	TenantID        string    `json:"tenant_id" db:"tenant_id"` // Multi-tenancy
	CheckCmd        string    `json:"check_cmd" db:"check_cmd"`
	ApplyCmd        string    `json:"apply_cmd" db:"apply_cmd"`
	DesiredExitCode int       `json:"desired_exit_code" db:"desired_exit_code"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Version         int       `json:"version" db:"version"`
	Status          string    `json:"status" db:"status"` // "compliant", "drifted", "failed"
	LastChecked     time.Time `json:"last_checked" db:"last_checked"`
	LastError       string    `json:"last_error" db:"last_error"`
}

// TimelineEvent represents an audit log entry.
type TimelineEvent struct {
	EventID   string            `json:"event_id" db:"event_id"`
	JobID     string            `json:"job_id" db:"job_id"`
	ReqID     string            `json:"req_id" db:"req_id"`
	Stage     string            `json:"stage" db:"stage"`
	Timestamp time.Time         `json:"timestamp" db:"timestamp"`
	Metadata  map[string]string `json:"metadata" db:"metadata"`
}
