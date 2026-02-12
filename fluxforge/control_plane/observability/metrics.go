package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TaskQueueDepth tracks the number of pending tasks in the queue.
	TaskQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_queue_depth",
		Help: "Current number of tasks in the scheduling queue",
	}, []string{"priority"})

	// SchedulerDecisions tracks the number of decisions made by type.
	SchedulerDecisions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flux_scheduler_decisions_total",
		Help: "Total number of scheduling decisions made",
	}, []string{"decision", "reason"})

	// DomainHealth tracks the failure rate of failure domains.
	DomainHealth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_domain_health",
		Help: "Current failure rate of failure domains (0-1)",
	}, []string{"domain"})

	// SchedulerLoopDuration tracks the duration of the scheduling loop.
	SchedulerLoopDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_scheduler_loop_duration_seconds",
		Help:    "Duration of the main scheduling loop iteration",
		Buckets: prometheus.DefBuckets,
	})

	// QueueOldestTaskAge tracks the age of the oldest task in the queue.
	QueueOldestTaskAge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_queue_oldest_task_age_seconds",
		Help: "Age of the oldest task in the queue in seconds",
	}, []string{"tenant", "priority"})

	// SchedulerMode tracks the current operating mode.
	SchedulerModeMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_scheduler_mode",
		Help: "Current scheduler mode (1=Normal, 2=Degraded, 3=ReadOnly, 4=Draining)",
	}, []string{"mode"})
)
