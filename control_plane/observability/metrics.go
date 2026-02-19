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

	// LeadershipEpoch tracks the current fencing epoch for the leader.
	LeadershipEpoch = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_leader_epoch",
		Help: "Current fencing epoch of the leader",
	}, []string{"node_id"})

	// LeadershipTransitions tracks leadership acquisition and loss events.
	LeadershipTransitions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flux_leader_transitions_total",
		Help: "Total number of leadership transitions",
	}, []string{"node_id", "event"})

	// === Phase 5.1: Critical Production Hardening Metrics ===

	// TaskTimeouts tracks tasks forcibly terminated due to timeout.
	TaskTimeouts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flux_task_timeouts_total",
		Help: "Tasks forcibly terminated due to timeout",
	}, []string{"state_id", "phase", "timeout_reason"}) // timeout_reason: runtime_limit, leadership_loss, shutdown

	// TaskRuntimeSeconds tracks the execution time of tasks (for tuning kill switch).
	TaskRuntimeSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_task_runtime_seconds",
		Help:    "Task execution time distribution",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10), // 1s to ~17min
	})

	// SchedulerQueueDepth tracks current queue depth (circuit breaker signal).
	SchedulerQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "flux_scheduler_queue_depth",
		Help: "Current number of tasks in scheduler queue",
	})

	// SchedulerWorkerSaturation tracks worker utilization (circuit breaker signal).
	SchedulerWorkerSaturation = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "flux_scheduler_worker_saturation",
		Help: "Ratio of active workers to max concurrency (0.0-1.0)",
	})

	// SchedulerRejections tracks tasks rejected by scheduler.
	SchedulerRejections = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flux_scheduler_rejections_total",
		Help: "Tasks rejected by scheduler admission control",
	}, []string{"reason"}) // circuit_open, not_leader, degraded_mode

	// SchedulerCircuitState tracks circuit breaker state.
	SchedulerCircuitState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_scheduler_circuit_state",
		Help: "Circuit breaker state (0=closed, 1=half_open, 2=open)",
	}, []string{"state"})

	// EventPublishFailures tracks failed event publish attempts (non-blocking).
	EventPublishFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flux_event_publish_failures_total",
		Help: "Failed event publish attempts (non-blocking, best-effort)",
	}, []string{"event_type", "reason"})

	// === Phase 6: Pilot Operations Telemetry ===

	// IntentAgeSeconds tracks the age of pending intents (time since state became pending).
	// "North Star" metric for user happiness.
	IntentAgeSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_intent_age_seconds",
		Help:    "Age of pending intents (time from pending to reconciliation start)",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1s to ~1h
	})

	// TaskRetries tracks the total number of task retries.
	// Used to calculate Retry Burn Rate (retries / successes).
	TaskRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_task_retries_total",
		Help: "Total number of task retry attempts",
	})

	// TaskSuccesses tracks the total number of successfully completed tasks.
	// Used to calculate Retry Burn Rate.
	TaskSuccesses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_task_success_total",
		Help: "Total number of successfully completed tasks",
	})

	// DBPendingStates tracks the number of pending states in the DB.
	// Used to detect Queue vs DB skew (checking for lost updates).
	DBPendingStates = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_db_pending_states",
		Help: "Current number of pending states in the database",
	}, []string{"tenant"})

	// SchedulerAdmissionWaitSeconds tracks time tasks wait in the internal queue.
	// Used for Backpressure visibility.
	SchedulerAdmissionWaitSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_scheduler_admission_wait_seconds",
		Help:    "Time tasks wait in the internal queue before being picked up by a worker",
		Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to ~40s
	})

	// -- Phase 6.1: Pilot Operational Metrics --

	RuntimeMode = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_runtime_mode",
		Help: "Current runtime mode configuration (1 = active)",
	}, []string{"mode"})

	IntegritySkew = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flux_integrity_skew_count",
		Help: "Detected count of tasks submitted but lost/unaccounted for (Silent Success Detector)",
	}, []string{"tenant"})

	// === High-Value Observability Metrics ===

	// LeadershipTransitionDuration tracks time taken for leadership transitions.
	LeadershipTransitionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_leader_transition_duration_seconds",
		Help:    "Time taken for leadership transition (step-down to become-leader)",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~100s
	})

	// APIRateLimited tracks API requests rejected by rate limiter.
	APIRateLimited = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "flux_api_rate_limited_total",
		Help: "API requests rejected by rate limiter (storm protection)",
	}, []string{"endpoint"}) // heartbeat, reconcile

	// SchedulerTaskWaitSeconds tracks queue wait time (overload early signal).
	SchedulerTaskWaitSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_scheduler_task_wait_seconds",
		Help:    "Time tasks spend waiting in queue before execution",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~100s
	})

	// RedisLatency tracks Redis operation roundtrip latency.
	RedisLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "flux_redis_roundtrip_latency_seconds",
		Help:    "Redis operation latency (coordination spine health)",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
	})

	// === Atomic Enforcement Metrics (Production Hardening) ===

	// VersionedWriteSuccess tracks successful atomic versioned writes
	VersionedWriteSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_versioned_write_success_total",
		Help: "Total number of successful versioned writes",
	})

	// VersionedWriteConflict tracks version conflicts detected
	VersionedWriteConflict = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_versioned_write_conflict_total",
		Help: "Total number of version conflicts detected",
	})

	// ReconciliationEpochAbort tracks reconciliations aborted due to epoch change
	// CRITICAL: This is the "smoking gun" metric for leader safety enforcement
	ReconciliationEpochAbort = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_reconciliation_epoch_abort_total",
		Help: "Total number of reconciliations aborted due to epoch change mid-reconcile",
	})

	// LeaderStatus tracks current leader status
	LeaderStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "flux_leader_status",
		Help: "Current leader status (1 = leader, 0 = follower)",
	})

	// IdempotencyLockAcquired tracks idempotency locks acquired
	IdempotencyLockAcquired = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_idempotency_lock_acquired_total",
		Help: "Total number of idempotency locks acquired",
	})

	// IdempotencyLockExpired tracks locks that expired
	IdempotencyLockExpired = promauto.NewCounter(prometheus.CounterOpts{
		Name: "flux_idempotency_lock_expired_total",
		Help: "Total number of idempotency locks that expired",
	})

	// ConnectedAgents tracks the number of currently connected agents
	ConnectedAgents = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "flux_connected_agents",
		Help: "Current number of connected agents",
	})
)
