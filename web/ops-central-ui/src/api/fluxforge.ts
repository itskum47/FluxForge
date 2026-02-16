import type { AdmissionMode } from '../types';

const BASE_URL = import.meta.env.VITE_API_URL || '/api';

// Backend Dashboard API types (from control_plane/api_dashboard.go)
export interface BackendDashboardMetrics {
    // Scheduler
    queue_depth: number;
    active_tasks: number;
    max_concurrency: number;
    worker_saturation: number;
    circuit_breaker_state: string;
    admission_mode: string;
    runtime_mode: string;

    // Leadership
    is_leader: boolean;
    current_epoch: number;
    leader_transitions: number;
    node_id: string;

    // Store
    pending_states: number;
    drifted_states: number;
    active_agents: number;

    // Multi-Cluster
    cluster_id: string;
    cluster_role: string;
    region: string;

    // Timestamp
    timestamp: number;
}

// Fetch tenant-scoped dashboard metrics
export const fetchDashboardMetrics = async (tenantID: string): Promise<BackendDashboardMetrics> => {
    const response = await fetch(`${BASE_URL}/dashboard`, {
        headers: {
            'X-Tenant-ID': tenantID,
            'Content-Type': 'application/json'
        }
    });

    if (!response.ok) {
        throw new Error(`Dashboard API failed: ${response.statusText}`);
    }

    return response.json();
};

// Fetch Prometheus metrics (for backward compatibility)
export const fetchPrometheusMetrics = async (): Promise<string> => {
    const response = await fetch('http://localhost:8080/metrics');
    if (!response.ok) throw new Error('Failed to fetch Prometheus metrics');
    return response.text();
};

export const fetchHealth = async (): Promise<boolean> => {
    try {
        const response = await fetch('http://localhost:8080/health');
        return response.ok;
    } catch {
        return false;
    }
}

export const setAdmissionMode = async (mode: AdmissionMode): Promise<boolean> => {
    try {
        const response = await fetch('http://localhost:8080/admin/admission-mode', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ mode }),
        });
        return response.ok;
    } catch (error) {
        console.error('Failed to set admission mode:', error);
        return false;
    }
};
