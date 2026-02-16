import React, { useEffect, useState, useRef } from 'react';
import { useTenant } from '../contexts/TenantContext'; // Correct relative path
import type { DashboardMetrics } from '../types';
import {
    Activity,
    Server,
    Database,
    AlertTriangle,
    Clock,
    Layers
} from 'lucide-react';

export function Dashboard() {
    const { tenantID } = useTenant();
    const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
    const [connected, setConnected] = useState(false);

    const wsRef = useRef<WebSocket | null>(null);

    // Initial fetch via REST
    useEffect(() => {
        const fetchMetrics = async () => {
            try {
                const response = await fetch(`${import.meta.env.VITE_API_URL}/dashboard`, {
                    headers: {
                        'X-Tenant-ID': tenantID
                    }
                });
                if (!response.ok) throw new Error('Failed to fetch metrics');
                const data = await response.json();
                setMetrics(data);
            } catch (err) {
                console.error("Initial fetch failed:", err);
            }
        };

        fetchMetrics();
    }, [tenantID]);


    useEffect(() => {
        // Close existing connection if any
        if (wsRef.current) {
            wsRef.current.close();
        }

        const wsUrl = `${import.meta.env.VITE_WS_URL}`;
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        // We can't easily send headers with WebSocket API in browser,
        // so we might need to send an auth message or use a query param.
        // However, the backend middleware expects X-Tenant-ID header.
        // The standard WebSocket API does NOT support custom headers.
        // Workaround: The backend `api_stream.go` uses `middleware.GetTenantFromContext`.
        // The middleware `tenant.go` reads from Header `X-Tenant-ID`.
        // This is a problem for standard browser WebSockets.
        // 
        // FIX: We need to modify the backend to ALSO accept a query param `tenant_id`
        // OR we relies on the cookie if we had one.
        // Since we can't change backend easily right now without going back,
        // let's stick to Polling for now if WS fails due to auth.
        // 
        // WAIT! I modified api_stream.go to use middleware.GetTenantFromContext.
        // Let's assume for now we use polling for the MVP dashboard
        // or we assume the browser environment might not work with custom headers.
        //
        // Actually, let's implemented Polling as a fallback or primary for now 
        // to ensure it works, as WS with custom headers is tricky in browsers.
        // Let's stick to REST polling every 5s for reliability first.

        // Changing strategy to Polling for robust MVP
        const interval = setInterval(async () => {
            try {
                const response = await fetch(`${import.meta.env.VITE_API_URL}/dashboard`, {
                    headers: {
                        'X-Tenant-ID': tenantID
                    }
                });
                if (response.ok) {
                    const data = await response.json();
                    setMetrics(data);
                    setConnected(true);
                } else {
                    setConnected(false);
                }
            } catch (err) {
                setConnected(false);
            }
        }, 2000); // 2 seconds poll

        return () => clearInterval(interval);
    }, [tenantID]);

    if (!metrics) {
        return <div className="p-8 text-center">Loading Dashboard for Tenant: {tenantID}...</div>;
    }

    return (
        <div className="p-6 max-w-7xl mx-auto space-y-6">
            {/* Header */}
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Operations Central</h1>
                    <p className="text-gray-500">Tenant: <span className="font-mono font-bold text-indigo-600">{tenantID}</span></p>
                </div>
                <div className="flex items-center space-x-2">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${connected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                        {connected ? 'Live' : 'Disconnected'}
                    </span>
                    <span className="text-xs text-gray-400">Last updated: {new Date(metrics.timestamp * 1000).toLocaleTimeString()}</span>
                </div>
            </div>

            {/* KPI Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <KPICard
                    title="Active Agents"
                    value={metrics.active_agents}
                    icon={<Server className="w-5 h-5 text-indigo-600" />}
                    change="Stable"
                />
                <KPICard
                    title="Pending States"
                    value={metrics.pending_states}
                    icon={<Clock className="w-5 h-5 text-amber-500" />}
                    change={metrics.pending_states > 0 ? "Processing" : "Idle"}
                />
                <KPICard
                    title="Drifted States"
                    value={metrics.drifted_states}
                    icon={<AlertTriangle className="w-5 h-5 text-red-500" />}
                    change={metrics.drifted_states > 0 ? "Action Required" : "Healthy"}
                />
                <KPICard
                    title="Active Tasks"
                    value={metrics.active_tasks}
                    icon={<Activity className="w-5 h-5 text-blue-500" />}
                    change={`${(metrics.worker_saturation * 100).toFixed(0)}% Saturation`}
                />
            </div>

            {/* Leadership & Cluster Status */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
                    <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
                        <Layers className="w-5 h-5 mr-2 text-gray-500" />
                        Cluster Status
                    </h3>
                    <div className="space-y-3">
                        <StatusRow label="Cluster ID" value={metrics.cluster_id} />
                        <StatusRow label="Role" value={metrics.cluster_role} />
                        <StatusRow label="Region" value={metrics.region} />
                        <StatusRow label="Admission Mode" value={metrics.admission_mode} />
                    </div>
                </div>

                <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
                    <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
                        <Database className="w-5 h-5 mr-2 text-gray-500" />
                        Leadership
                    </h3>
                    <div className="space-y-3">
                        <StatusRow label="Is Leader" value={metrics.is_leader ? "Yes" : "No"} highlight={metrics.is_leader} />
                        <StatusRow label="Node ID" value={metrics.node_id} />
                        <StatusRow label="Current Epoch" value={metrics.current_epoch.toString()} />
                        <StatusRow label="Transitions" value={metrics.leader_transitions.toString()} />
                    </div>
                </div>
            </div>
        </div>
    );
}

function KPICard({ title, value, icon, change }: { title: string, value: number | string, icon: React.ReactNode, change: string }) {
    return (
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100">
            <div className="flex justify-between items-start">
                <div>
                    <p className="text-sm font-medium text-gray-500">{title}</p>
                    <p className="mt-1 text-3xl font-semibold text-gray-900">{value}</p>
                </div>
                <div className="p-2 bg-gray-50 rounded-lg">
                    {icon}
                </div>
            </div>
            <div className="mt-4">
                <p className="text-sm text-gray-500">{change}</p>
            </div>
        </div>
    );
}

function StatusRow({ label, value, highlight = false }: { label: string, value: string, highlight?: boolean }) {
    return (
        <div className="flex justify-between items-center text-sm">
            <span className="text-gray-500">{label}</span>
            <span className={`font-medium ${highlight ? 'text-green-600' : 'text-gray-900'}`}>{value}</span>
        </div>
    );
}
