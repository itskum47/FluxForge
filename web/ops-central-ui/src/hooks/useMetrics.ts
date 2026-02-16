import { useState, useEffect } from 'react';
import { useTenant } from '../contexts/TenantContext';
import type { DashboardMetrics } from '../types';
import { fetchDashboardMetrics, type BackendDashboardMetrics } from '../api/fluxforge';

const POLL_INTERVAL = 2000;

// Convert backend metrics to frontend DashboardMetrics type
const convertMetrics = (backend: BackendDashboardMetrics): DashboardMetrics => {
    return {
        // Map backend fields to frontend fields
        schedulerQueueDepth: backend.queue_depth,
        workerSaturation: backend.worker_saturation * 100, // Convert to percentage
        integritySkewCount: 0, // Not provided by new API
        leaderTransitionsTotal: backend.leader_transitions,

        // Infrastructure
        circuitBreakerOpen: backend.circuit_breaker_state === 'open' ? 1 : 0,
        redisLatency: 0, // Not provided by new API
        epochDrift: 0, // Not provided by new API

        // Runtime State
        runtimeMode: backend.runtime_mode as any,
        admissionMode: backend.admission_mode as any,

        // Performance
        intentAgeP99: 0, // Not provided by new API
        reconciliationSuccessRate: 99.2, // Mock for now
        successfulReconciliations: 42301, // Mock for now
        failedReconciliations: backend.drifted_states,
        activeIntents: backend.pending_states,

        // Status
        circuitBreakerStatus: backend.circuit_breaker_state === 'open' ? 'Open' : 'Closed',

        // Additional fields for display
        queueDepth: backend.queue_depth,
        activeTasks: backend.active_tasks,
        activeAgents: backend.active_agents,
        pendingStates: backend.pending_states,
        driftedStates: backend.drifted_states,

        // Timestamps
        lastUpdated: backend.timestamp * 1000, // Convert to milliseconds
    };
};

const initialMetrics: DashboardMetrics = {
    schedulerQueueDepth: 0,
    workerSaturation: 0,
    integritySkewCount: 0,
    leaderTransitionsTotal: 0,
    circuitBreakerOpen: 0,
    redisLatency: 0,
    epochDrift: 0,
    runtimeMode: 'production',
    admissionMode: 'normal',
    intentAgeP99: 0,
    reconciliationSuccessRate: 99.2,
    successfulReconciliations: 42301,
    failedReconciliations: 0,
    activeIntents: 0,
    circuitBreakerStatus: "Closed",
    lastUpdated: Date.now(),
};

export const useMetrics = () => {
    const { tenantID } = useTenant();
    const [metrics, setMetrics] = useState<DashboardMetrics>(initialMetrics);
    const [isConnected, setIsConnected] = useState(true);
    const [latency, setLatency] = useState(0);

    useEffect(() => {
        const poll = async () => {
            const start = Date.now();
            try {
                const backendMetrics = await fetchDashboardMetrics(tenantID);
                const convertedMetrics = convertMetrics(backendMetrics);

                setLatency(Date.now() - start);
                setIsConnected(true);
                setMetrics(convertedMetrics);
            } catch (error) {
                console.error('Metrics fetch failed:', error);
                setIsConnected(false);
            }
        };

        poll(); // Initial call
        const interval = setInterval(poll, POLL_INTERVAL);
        return () => clearInterval(interval);
    }, [tenantID]);

    // Format last updated time
    const lastUpdated = new Date(metrics.lastUpdated).toLocaleTimeString();

    return { metrics, isConnected, latency, lastUpdated };
};
