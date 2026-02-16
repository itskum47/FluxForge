export type AdmissionMode = "normal" | "drain" | "freeze";
export type RuntimeMode = "pilot" | "production" | "shadow";

export interface DashboardMetrics {
    // Core Metrics
    schedulerQueueDepth: number;
    workerSaturation: number;
    integritySkewCount: number;
    leaderTransitionsTotal: number;

    // Infrastructure
    circuitBreakerOpen: number; // 0 or 1
    redisLatency: number; // seconds
    epochDrift: number; // milliseconds

    // Runtime State
    runtimeMode: RuntimeMode;
    admissionMode: AdmissionMode;

    // Performance
    intentAgeP99: number;
    reconciliationSuccessRate: number;
    successfulReconciliations: number;
    failedReconciliations: number;
    activeIntents: number;

    // Status
    circuitBreakerStatus: "Closed" | "Open" | "Half-Open";

    // Additional fields from new backend API
    queueDepth?: number;
    activeTasks?: number;
    activeAgents?: number;
    pendingStates?: number;
    driftedStates?: number;

    // Timestamps
    lastUpdated: number;
}

export interface MetricPoint {
    timestamp: number;
    value: number;
}
