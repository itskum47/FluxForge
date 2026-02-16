// Dashboard API Service
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface DashboardMetrics {
    // Scheduler Metrics
    queue_depth: number;
    active_tasks: number;
    max_concurrency: number;
    worker_saturation: number;
    circuit_breaker_state: string;
    admission_mode: string;
    runtime_mode: string;

    // Leadership Metrics
    is_leader: boolean;
    current_epoch: number;
    leader_transitions: number;
    node_id: string;

    // Store Metrics
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

export interface ClusterInfo {
    cluster_id: string;
    region: string;
    role: string;
    is_leader: boolean;
    agent_count: number;
    health_score: number;
    endpoint: string;
}

export interface IncidentSnapshot {
    incident_id: string;
    state_id: string;
    node_id: string;
    failure_reason: string;
    timestamp: number;
    scheduler_snapshot: {
        queue_depth: number;
        active_tasks: number;
        worker_saturation: number;
        circuit_breaker_state: string;
        runtime_mode: string;
    };
    leader_snapshot: {
        is_leader: boolean;
        current_epoch: number;
        node_id: string;
    };
    timeline: Array<{
        req_id: string;
        stage: string;
        node_id: string;
        tenant_id: string;
        timestamp: number;
    }>;
}

class DashboardService {
    private baseUrl: string;

    constructor(baseUrl: string = API_BASE_URL) {
        this.baseUrl = baseUrl;
    }

    async getDashboardMetrics(): Promise<DashboardMetrics> {
        const response = await fetch(`${this.baseUrl}/api/dashboard`);
        if (!response.ok) {
            throw new Error(`Failed to fetch dashboard metrics: ${response.statusText}`);
        }
        return response.json();
    }

    async getClusters(): Promise<ClusterInfo[]> {
        const response = await fetch(`${this.baseUrl}/api/clusters`);
        if (!response.ok) {
            throw new Error(`Failed to fetch clusters: ${response.statusText}`);
        }
        return response.json();
    }

    async getIncidents(): Promise<IncidentSnapshot[]> {
        const response = await fetch(`${this.baseUrl}/api/incidents`);
        if (!response.ok) {
            throw new Error(`Failed to fetch incidents: ${response.statusText}`);
        }
        return response.json();
    }

    async replayIncident(incidentId: string): Promise<any> {
        const response = await fetch(`${this.baseUrl}/api/incidents/replay/${incidentId}`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error(`Failed to replay incident: ${response.statusText}`);
        }
        return response.json();
    }

    async captureIncidentSnapshot(stateId: string): Promise<IncidentSnapshot> {
        const response = await fetch(`${this.baseUrl}/api/incidents/capture?state_id=${stateId}`, {
            method: 'POST',
        });
        if (!response.ok) {
            throw new Error(`Failed to capture incident: ${response.statusText}`);
        }
        return response.json();
    }

    createWebSocketConnection(): WebSocket {
        const wsUrl = this.baseUrl.replace('http', 'ws');
        return new WebSocket(`${wsUrl}/api/dashboard/stream`);
    }
}

export const dashboardService = new DashboardService();
