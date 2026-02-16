package incident

import (
	"context"
	"time"

	"github.com/itskum47/FluxForge/control_plane/store"
	"github.com/itskum47/FluxForge/control_plane/timeline"
)

// IncidentReport represents a captured failure context for debugging.
type IncidentReport struct {
	StateID      string                    `json:"state_id"`
	DesiredState *store.DesiredState       `json:"desired_state"`
	Agent        *store.Agent              `json:"agent"`
	Events       []timeline.ReconcileEvent `json:"events"`
	Jobs         []*store.Job              `json:"jobs"`
	CapturedAt   time.Time                 `json:"captured_at"`
	Analysis     string                    `json:"analysis,omitempty"`
}

// StoreInterface defines dependencies needed for capture.
type StoreInterface interface {
	GetState(ctx context.Context, tenantID string, stateID string) (*store.DesiredState, error)
	GetAgent(ctx context.Context, tenantID string, nodeID string) (*store.Agent, error)
	ListJobs(ctx context.Context, tenantID string, nodeID string, limit int) ([]*store.Job, error)
}

// TimelineInterface defines timeline dependencies.
type TimelineInterface interface {
	GetEventsByStateID(stateID string) []timeline.ReconcileEvent
}

// CaptureIncident gathers all relevant data for a state failure.
func CaptureIncident(ctx context.Context, s StoreInterface, tl TimelineInterface, tenantID string, stateID string) (*IncidentReport, error) {
	state, err := s.GetState(ctx, tenantID, stateID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil // Not found
	}

	agent, err := s.GetAgent(ctx, tenantID, state.NodeID)
	if err != nil {
		return nil, err
	}

	// Fetch recent jobs for this node (limit 20)
	// Ideally we filter jobs by StateID, but ListJobs only takes NodeID.
	// We can filter manually or rely on Store update.
	// Store interface has ListJobs(nodeID, limit).
	// We will filter in memory.
	jobs, err := s.ListJobs(ctx, tenantID, state.NodeID, 50)
	if err != nil {
		return nil, err
	}

	var relevantJobs []*store.Job
	for _, j := range jobs {
		if j.StateID == stateID {
			relevantJobs = append(relevantJobs, j)
		}
	}

	// Fetch timeline events
	events := tl.GetEventsByStateID(stateID)

	report := &IncidentReport{
		StateID:      stateID,
		DesiredState: state,
		Agent:        agent,
		Events:       events,
		Jobs:         relevantJobs,
		CapturedAt:   time.Now(),
	}

	return report, nil
}
