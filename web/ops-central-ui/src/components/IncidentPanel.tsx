import React, { useState, useEffect } from 'react';
import { AlertCircle, Download, Play } from 'lucide-react';
import { dashboardService, IncidentSnapshot } from '../services/dashboardService';

export function IncidentPanel() {
    const [incidents, setIncidents] = useState<IncidentSnapshot[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [replayingId, setReplayingId] = useState<string | null>(null);

    useEffect(() => {
        loadIncidents();
    }, []);

    const loadIncidents = async () => {
        try {
            setLoading(true);
            const data = await dashboardService.getIncidents();
            setIncidents(data);
            setError(null);
        } catch (err) {
            setError((err as Error).message);
        } finally {
            setLoading(false);
        }
    };

    const handleReplay = async (incidentId: string) => {
        try {
            setReplayingId(incidentId);
            const replay = await dashboardService.replayIncident(incidentId);
            console.log('Replay result:', replay);
            alert(`Incident ${incidentId} replayed successfully`);
        } catch (err) {
            alert(`Failed to replay incident: ${(err as Error).message}`);
        } finally {
            setReplayingId(null);
        }
    };

    const handleDownload = (incident: IncidentSnapshot) => {
        const dataStr = JSON.stringify(incident, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `incident-${incident.incident_id}.json`;
        link.click();
        URL.revokeObjectURL(url);
    };

    if (loading) {
        return (
            <div className="glass-panel p-6 rounded-xl">
                <div className="flex items-center justify-center h-48">
                    <div className="text-gray-400">Loading incidents...</div>
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
            <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold text-white">Incident Replay</h2>
                <button
                    onClick={loadIncidents}
                    className="px-3 py-1 text-sm bg-white/10 hover:bg-white/20 rounded-lg transition-colors"
                >
                    Refresh
                </button>
            </div>

            {incidents.length === 0 ? (
                <div className="text-center py-12 text-gray-400">
                    <AlertCircle className="w-12 h-12 mx-auto mb-3 opacity-50" />
                    <p>No incidents captured yet</p>
                    <p className="text-sm mt-2">Incidents will appear here when failures occur</p>
                </div>
            ) : (
                <div className="space-y-3">
                    {incidents.map((incident) => (
                        <div
                            key={incident.incident_id}
                            className="bg-white/5 hover:bg-white/10 p-4 rounded-lg transition-colors"
                        >
                            <div className="flex items-start justify-between">
                                <div className="flex-1">
                                    <div className="flex items-center gap-2 mb-2">
                                        <span className="font-mono text-sm text-white">
                                            {incident.incident_id}
                                        </span>
                                        <span className="text-xs text-gray-400">
                                            {new Date(incident.timestamp * 1000).toLocaleString()}
                                        </span>
                                    </div>
                                    <div className="text-sm text-gray-300">
                                        <div>State: {incident.state_id}</div>
                                        <div>Node: {incident.node_id}</div>
                                        <div className="text-red-400 mt-1">{incident.failure_reason}</div>
                                    </div>
                                </div>
                                <div className="flex gap-2">
                                    <button
                                        onClick={() => handleReplay(incident.incident_id)}
                                        disabled={replayingId === incident.incident_id}
                                        className="p-2 bg-teal-500/20 hover:bg-teal-500/30 rounded-lg transition-colors disabled:opacity-50"
                                        title="Replay incident"
                                    >
                                        <Play className="w-4 h-4 text-teal-400" />
                                    </button>
                                    <button
                                        onClick={() => handleDownload(incident)}
                                        className="p-2 bg-white/10 hover:bg-white/20 rounded-lg transition-colors"
                                        title="Download snapshot"
                                    >
                                        <Download className="w-4 h-4 text-white" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
