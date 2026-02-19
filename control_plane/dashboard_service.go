package main

import (
	"context"
	"time"

	"github.com/itskum47/FluxForge/control_plane/coordination"
	"github.com/itskum47/FluxForge/control_plane/scheduler"
	"github.com/itskum47/FluxForge/control_plane/store"
)

// DashboardService provides an abstraction layer for dashboard data retrieval.
// It decouples the API from direct store access and aggregates data from multiple sources.
type DashboardService struct {
	store     store.Store
	scheduler *scheduler.Scheduler
	elector   *coordination.LeaderElector
}

// NewDashboardService creates a new DashboardService.
func NewDashboardService(store store.Store, scheduler *scheduler.Scheduler, elector *coordination.LeaderElector) *DashboardService {
	return &DashboardService{
		store:     store,
		scheduler: scheduler,
		elector:   elector,
	}
}

// GetDashboardMetrics collects and aggregates all metrics for a specific tenant.
func (s *DashboardService) GetDashboardMetrics(ctx context.Context, tenantID string) (DashboardMetrics, error) {
	// 1. Scheduler Metrics
	schedMetrics := s.scheduler.GetMetrics()

	// 2. Leadership Metrics
	var leaderState coordination.LeaderState
	if s.elector != nil {
		leaderState = s.elector.GetState()
	}

	// 3. Store Metrics (Tenant Scoped)
	pending, err := s.store.CountStatesByStatus(ctx, tenantID, "pending")
	if err != nil {
		return DashboardMetrics{}, err
	}

	drifted, err := s.store.CountStatesByStatus(ctx, tenantID, "drifted")
	if err != nil {
		return DashboardMetrics{}, err
	}

	agents, err := s.store.ListAgents(ctx, tenantID)
	if err != nil {
		return DashboardMetrics{}, err
	}

	// 4. Construct Response
	return DashboardMetrics{
		// Scheduler
		QueueDepth:          schedMetrics.QueueDepth,
		ActiveTasks:         schedMetrics.ActiveTasks,
		MaxConcurrency:      schedMetrics.MaxConcurrency,
		WorkerSaturation:    schedMetrics.WorkerSaturation,
		CircuitBreakerState: schedMetrics.CircuitBreakerState,
		AdmissionMode:       schedMetrics.AdmissionMode,
		RuntimeMode:         schedMetrics.RuntimeMode,

		// Leadership
		IsLeader:          leaderState.IsLeader,
		CurrentEpoch:      leaderState.CurrentEpoch,
		LeaderTransitions: leaderState.Transitions,
		NodeID:            leaderState.NodeID,

		// Store
		PendingStates: pending,
		DriftedStates: drifted,
		ActiveAgents:  len(agents),

		// Cluster Info
		ClusterID:   "cluster-primary",
		ClusterRole: getClusterRole(leaderState.IsLeader),
		Region:      "us-east-1",
		Timestamp:   time.Now().Unix(),
	}, nil
}

func getClusterRole(isLeader bool) string {
	if isLeader {
		return "leader"
	}
	return "follower"
}
