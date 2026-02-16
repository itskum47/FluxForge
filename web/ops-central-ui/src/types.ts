export interface DashboardMetrics {
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
