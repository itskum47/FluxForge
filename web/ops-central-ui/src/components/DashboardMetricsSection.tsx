import React from 'react';
import { Server, Clock, AlertTriangle, Activity, Layers, Database } from 'lucide-react';
import { useTenant } from '../contexts/TenantContext';
import type { DashboardMetrics } from '../types';

interface DashboardMetricsSectionProps {
    metrics: DashboardMetrics;
}

function KPICard({ title, value, icon, change }: { title: string, value: number | string, icon: React.ReactNode, change: string }) {
    return (
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-100 hover:shadow-md transition-shadow">
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

export function DashboardMetricsSection({ metrics }: DashboardMetricsSectionProps) {
    const { tenantID } = useTenant();

    return (
        <div className="space-y-4">
            {/* Tenant Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                    <Database className="w-4 h-4 text-indigo-600" />
                    <span className="text-sm text-gray-500">
                        Tenant: <span className="font-mono font-semibold text-indigo-600">{tenantID}</span>
                    </span>
                </div>
            </div>

            {/* KPI Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                <KPICard
                    title="Active Agents"
                    value={metrics.activeAgents ?? metrics.active_agents ?? 0}
                    icon={<Server className="w-5 h-5 text-indigo-600" />}
                    change="Stable"
                />
                <KPICard
                    title="Pending States"
                    value={metrics.pendingStates ?? metrics.pending_states ?? 0}
                    icon={<Clock className="w-5 h-5 text-amber-500" />}
                    change={(metrics.pendingStates ?? metrics.pending_states ?? 0) > 0 ? "Processing" : "Idle"}
                />
                <KPICard
                    title="Drifted States"
                    value={metrics.driftedStates ?? metrics.drifted_states ?? 0}
                    icon={<AlertTriangle className="w-5 h-5 text-red-500" />}
                    change={(metrics.driftedStates ?? metrics.drifted_states ?? 0) > 0 ? "Action Required" : "Healthy"}
                />
                <KPICard
                    title="Active Tasks"
                    value={metrics.activeTasks ?? metrics.active_tasks ?? 0}
                    icon={<Activity className="w-5 h-5 text-blue-500" />}
                    change={`${Math.round(metrics.workerSaturation ?? metrics.worker_saturation ?? 0)}% Saturation`}
                />
            </div>

            {/* Cluster Status Bar */}
            <div className="bg-gradient-to-r from-indigo-50 to-blue-50 border border-indigo-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                        <Layers className="w-4 h-4 text-indigo-600" />
                        <span className="text-sm font-medium text-gray-700">Cluster Status</span>
                    </div>
                    <div className="flex items-center space-x-6 text-sm">
                        <div>
                            <span className="text-gray-500">Queue:</span>
                            <span className="ml-1 font-semibold text-gray-900">{metrics.queueDepth ?? metrics.queue_depth ?? 0}</span>
                        </div>
                        <div>
                            <span className="text-gray-500">Mode:</span>
                            <span className="ml-1 font-semibold text-gray-900 capitalize">{metrics.admissionMode ?? metrics.admission_mode ?? 'normal'}</span>
                        </div>
                        <div>
                            <span className="text-gray-500">Circuit:</span>
                            <span className={`ml-1 font-semibold ${(metrics.circuitBreakerStatus ?? metrics.circuit_breaker_state) === 'Closed' ? 'text-green-600' : 'text-red-600'}`}>
                                {metrics.circuitBreakerStatus ?? metrics.circuit_breaker_state ?? 'Closed'}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
