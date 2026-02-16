import React, { useState, useEffect } from 'react';
import { Server, CheckCircle, AlertCircle, Activity } from 'lucide-react';
import { dashboardService, ClusterInfo } from '../services/dashboardService';

export function ClusterOverview() {
    const [clusters, setClusters] = useState<ClusterInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        loadClusters();
        const interval = setInterval(loadClusters, 10000); // Refresh every 10s
        return () => clearInterval(interval);
    }, []);

    const loadClusters = async () => {
        try {
            const data = await dashboardService.getClusters();
            setClusters(data);
            setError(null);
        } catch (err) {
            setError((err as Error).message);
        } finally {
            setLoading(false);
        }
    };

    const getRoleColor = (role: string) => {
        switch (role) {
            case 'leader':
                return 'text-green-400 bg-green-500/20';
            case 'follower':
                return 'text-blue-400 bg-blue-500/20';
            case 'standby':
                return 'text-gray-400 bg-gray-500/20';
            default:
                return 'text-gray-400 bg-gray-500/20';
        }
    };

    const getHealthColor = (score: number) => {
        if (score >= 0.9) return 'text-green-400';
        if (score >= 0.7) return 'text-yellow-400';
        return 'text-red-400';
    };

    if (loading) {
        return (
            <div className="glass-panel p-6 rounded-xl">
                <div className="flex items-center justify-center h-48">
                    <div className="text-gray-400">Loading cluster info...</div>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="glass-panel p-6 rounded-xl border-red-500/30">
                <div className="flex items-center gap-2 text-red-400">
                    <AlertCircle className="w-5 h-5" />
                    <span>Error: {error}</span>
                </div>
            </div>
        );
    }

    return (
        <div className="glass-panel p-6 rounded-xl">
            <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-white flex items-center gap-2">
                    <Server className="w-5 h-5" />
                    Cluster Topology
                </h2>
                <div className="text-sm text-gray-400">
                    {clusters.length} {clusters.length === 1 ? 'node' : 'nodes'}
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {clusters.map((cluster) => (
                    <div
                        key={cluster.cluster_id}
                        className={`bg-white/5 p-4 rounded-lg border ${cluster.is_leader ? 'border-green-500/30' : 'border-white/10'
                            }`}
                    >
                        <div className="flex items-start justify-between mb-3">
                            <div>
                                <div className="font-mono text-sm text-white mb-1">
                                    {cluster.cluster_id}
                                </div>
                                <div className="text-xs text-gray-400">{cluster.region}</div>
                            </div>
                            <span
                                className={`px-2 py-1 rounded text-xs font-medium ${getRoleColor(
                                    cluster.role
                                )}`}
                            >
                                {cluster.role}
                            </span>
                        </div>

                        <div className="space-y-2">
                            <div className="flex items-center justify-between text-sm">
                                <span className="text-gray-400">Health</span>
                                <span className={`font-medium ${getHealthColor(cluster.health_score)}`}>
                                    {(cluster.health_score * 100).toFixed(1)}%
                                </span>
                            </div>

                            <div className="flex items-center justify-between text-sm">
                                <span className="text-gray-400">Agents</span>
                                <span className="text-white font-medium">{cluster.agent_count}</span>
                            </div>

                            <div className="flex items-center justify-between text-sm">
                                <span className="text-gray-400">Status</span>
                                <div className="flex items-center gap-1">
                                    {cluster.is_leader ? (
                                        <>
                                            <CheckCircle className="w-4 h-4 text-green-400" />
                                            <span className="text-green-400">Active</span>
                                        </>
                                    ) : (
                                        <>
                                            <Activity className="w-4 h-4 text-blue-400" />
                                            <span className="text-blue-400">Standby</span>
                                        </>
                                    )}
                                </div>
                            </div>
                        </div>

                        <div className="mt-3 pt-3 border-t border-white/10">
                            <a
                                href={cluster.endpoint}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-xs text-gray-400 hover:text-white transition-colors"
                            >
                                {cluster.endpoint}
                            </a>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}
