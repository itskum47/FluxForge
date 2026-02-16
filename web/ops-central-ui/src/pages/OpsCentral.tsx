import React, { useState } from 'react';
import { useMetrics } from '../hooks/useMetrics';
import { GlobalBanner } from '../components/GlobalBanner';
import { Header } from '../components/Header';
import { IntentAgeCard } from '../components/IntentAgeCard';
import { ReconciliationCard } from '../components/ReconciliationCard';
import { PendingIntentChart } from '../components/PendingIntentChart';
import { SchedulerQueueCard } from '../components/SchedulerQueueCard';
import { WorkerSaturationCard } from '../components/WorkerSaturationCard';
import { CircuitBreakerTimeline } from '../components/CircuitBreakerTimeline';
import { LeadershipPanel } from '../components/LeadershipPanel';
import { AIInsightsPanel } from '../components/AIInsightsPanel';
import type { AdmissionMode } from '../types';
import { setAdmissionMode } from '../api/fluxforge';
import { ShieldAlert, Zap, Lock, Play } from 'lucide-react';

export const OpsCentral: React.FC = () => {
    const { metrics, isConnected } = useMetrics();
    const [isChangingMode, setIsChangingMode] = useState(false);

    // Simulated Histories (would normally be time-series from backend)
    // For now we just append current value if we were building a history reducer,
    // but to keep it stateless visual we generate some "fake" history if empty
    const mockHistory = [
        { value: metrics.intentAgeP99 - 5 },
        { value: metrics.intentAgeP99 + 2 },
        { value: metrics.intentAgeP99 - 1 },
        { value: metrics.intentAgeP99 }
    ];
    const mockVolumeHistory = [
        { timestamp: 1, value: 2000 },
        { timestamp: 2, value: 2200 },
        { timestamp: 3, value: 2100 },
        { timestamp: 4, value: 2400 }
    ];

    const handleModeChange = async (mode: AdmissionMode) => {
        setIsChangingMode(true);
        await setAdmissionMode(mode);
        setIsChangingMode(false);
    };

    return (
        <>
            <GlobalBanner mode={metrics.admissionMode} isConnected={isConnected} />
            <Header />

            <main className="flex-1 overflow-y-auto p-6 space-y-6 max-w-[1920px] mx-auto w-full">

                {/* ROW 1: Customer Impact */}
                <section>
                    <div className="flex items-center space-x-2 mb-3">
                        <ShieldAlert size={14} className="text-primary" />
                        <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Customer Impact</h2>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        <IntentAgeCard
                            value={metrics.intentAgeP99}
                            trend={-12} // Static trend for now
                            history={mockHistory}
                        />
                        <ReconciliationCard
                            successRate={metrics.reconciliationSuccessRate}
                            successCount={42301} // Placeholder
                            retryCount={124}     // Placeholder
                            failCount={12}       // Placeholder
                        />
                        <PendingIntentChart
                            volume={2400} // Placeholder for volume aggregation
                            history={mockVolumeHistory}
                            isHighThroughput={true}
                        />
                    </div>
                </section>

                {/* ROW 2: Control Plane Health & AI Insights */}
                <section className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                    {/* Control Plane Column (Spans 3) */}
                    <div className="lg:col-span-3 space-y-3 flex flex-col">
                        <div className="flex items-center space-x-2">
                            <Zap size={14} className="text-blue-400" />
                            <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Control Plane Health</h2>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 flex-1">
                            <SchedulerQueueCard depth={metrics.schedulerQueueDepth} cap={500} />
                            <WorkerSaturationCard saturation={metrics.workerSaturation} />
                            <CircuitBreakerTimeline isOpen={metrics.circuitBreakerOpen === 1} />
                        </div>
                    </div>

                    {/* AI Insights Column */}
                    <div className="lg:col-span-1 flex flex-col space-y-3 h-full">
                        <div className="flex items-center space-x-2">
                            <span className="text-sm font-semibold text-gray-400 uppercase tracking-widest">AI Insights</span>
                        </div>
                        <AIInsightsPanel />
                    </div>
                </section>

                {/* ROW 3: Infrastructure Health */}
                <section>
                    <div className="flex items-center space-x-2 mb-3">
                        <Lock size={14} className="text-gray-400" />
                        <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Infrastructure Health</h2>
                    </div>
                    <LeadershipPanel
                        transitions={metrics.leaderTransitionsTotal}
                        redisLatency={metrics.redisLatency}
                        epochDrift={metrics.epochDrift}
                    />
                </section>

                {/* Footer / Status Bar - Debug Controls */}
                <footer className="mt-8 border-t border-white/5 pt-4 pb-8 flex flex-col md:flex-row justify-between items-center text-center md:text-left gap-4">
                    <p className="text-[10px] text-gray-600 font-mono">
                        FluxForge Ops Central v6.0.0 · Region: local · {metrics.runtimeMode.toUpperCase()} Mode
                    </p>

                    {/* Admin Controls */}
                    <div className="flex items-center gap-2 bg-black/20 p-1 rounded-lg border border-white/5">
                        <span className="text-[10px] text-gray-500 uppercase px-2 font-mono">Admin:</span>
                        <button
                            onClick={() => handleModeChange('normal')}
                            disabled={isChangingMode || metrics.admissionMode === 'normal'}
                            className={`p-1.5 rounded transition-colors ${metrics.admissionMode === 'normal' ? 'bg-emerald-500/20 text-emerald-400' : 'hover:bg-white/10 text-gray-400'}`}
                            title="Set Normal Mode"
                        >
                            <Play size={12} />
                        </button>
                        <button
                            onClick={() => handleModeChange('drain')}
                            disabled={isChangingMode || metrics.admissionMode === 'drain'}
                            className={`p-1.5 rounded transition-colors ${metrics.admissionMode === 'drain' ? 'bg-amber-500/20 text-amber-400' : 'hover:bg-white/10 text-gray-400'}`}
                            title="Set Drain Mode"
                        >
                            <ShieldAlert size={12} />
                        </button>
                        <button
                            onClick={() => handleModeChange('freeze')}
                            disabled={isChangingMode || metrics.admissionMode === 'freeze'}
                            className={`p-1.5 rounded transition-colors ${metrics.admissionMode === 'freeze' ? 'bg-red-500/20 text-red-400' : 'hover:bg-white/10 text-gray-400'}`}
                            title="Set Freeze Mode"
                        >
                            <Lock size={12} />
                        </button>
                    </div>
                </footer>

            </main>
        </>
    );
};
