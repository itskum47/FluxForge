import React, { useState } from 'react';
import {
    Activity,
    ArrowDown,
    Workflow,
    Sparkles,
    Server,
    CheckCircle
} from 'lucide-react';
import { useMetrics } from '../hooks/useMetrics';
import type { AdmissionMode } from '../types';
import { DashboardMetricsSection } from '../components/DashboardMetricsSection';

interface NormalModeProps {
    onSetMode: (mode: AdmissionMode) => void;
}

export const NormalMode: React.FC<NormalModeProps> = () => {
    const { metrics, isConnected, lastUpdated } = useMetrics();
    const [timeRange, setTimeRange] = useState('1h');

    // Helper for formatting large numbers
    const formatNumber = (num: number) => {
        return new Intl.NumberFormat('en-US', { notation: "compact", maximumFractionDigits: 1 }).format(num);
    };

    return (
        <div className="bg-ops-bg text-gray-300 min-h-screen grid-bg flex flex-col relative font-display">
            {/* Global Status Banner */}
            <div className="w-full bg-ops-success/10 border-b border-ops-success/20 py-1.5 px-6 flex items-center justify-center sticky top-0 z-50 backdrop-blur-sm">
                <div className="flex items-center space-x-2 text-ops-success text-xs font-mono font-medium tracking-wide">
                    <span className="relative flex h-2 w-2">
                        <span className={`animate-ping absolute inline-flex h-full w-full rounded-full bg-ops-success opacity-75 ${!isConnected ? 'hidden' : ''}`}></span>
                        <span className={`relative inline-flex rounded-full h-2 w-2 ${isConnected ? 'bg-ops-success' : 'bg-red-500'}`}></span>
                    </span>
                    <span className="uppercase">Normal: {isConnected ? 'Connected' : 'Disconnected'} · Stable</span>
                </div>
            </div>

            {/* Header */}
            <header className="px-6 py-4 border-b border-white/5 bg-ops-bg/90 backdrop-blur flex justify-between items-center z-40">
                <div className="flex items-center space-x-4">
                    <div className="flex items-center space-x-2">
                        <div className="w-8 h-8 rounded bg-gradient-to-br from-primary to-purple-800 flex items-center justify-center text-white font-bold text-lg">
                            F
                        </div>
                        <div className="flex flex-col">
                            <h1 className="text-white font-semibold leading-none tracking-tight">FluxForge</h1>
                            <span className="text-xs text-gray-500 font-mono tracking-wider uppercase">Ops Central</span>
                        </div>
                    </div>
                </div>
                {/* Middle: Time Filters */}
                <div className="bg-ops-card border border-white/5 rounded-lg p-1 flex space-x-1">
                    {['15m', '1h', '6h', '24h'].map((range) => (
                        <button
                            key={range}
                            onClick={() => setTimeRange(range)}
                            className={`px-3 py-1 text-xs font-medium rounded transition-colors ${timeRange === range ? 'bg-primary text-white shadow-sm shadow-primary/20' : 'text-gray-400 hover:text-white'}`}
                        >
                            {range}
                        </button>
                    ))}
                </div>
                {/* Right: Meta */}
                <div className="flex items-center space-x-4 text-xs font-mono text-gray-400">
                    <div className="flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-white/5 border border-white/5">
                        <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                        <span>Live Stream</span>
                    </div>
                    <span>Updated: {lastUpdated}</span>
                </div>
            </header>

            {/* Main Content Grid */}
            <main className="flex-1 overflow-y-auto p-6 space-y-6 pb-20 custom-scrollbar">
                {/* Dashboard Metrics Overview */}
                <section>
                    <div className="flex items-center space-x-2 mb-3">
                        <Server className="w-4 h-4 text-primary" />
                        <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">System Overview</h2>
                    </div>
                    <DashboardMetricsSection metrics={metrics} />
                </section>

                {/* ROW 1: Customer Impact */}
                <section>
                    <div className="flex items-center space-x-2 mb-3">
                        <Activity className="w-4 h-4 text-primary" />
                        <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Customer Impact</h2>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        {/* Intent Age P99 */}
                        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col justify-between shadow-lg shadow-black/20 group hover:border-white/10 transition-all">
                            <div className="flex justify-between items-start">
                                <div>
                                    <h3 className="text-gray-400 text-sm font-medium">Intent Age P99</h3>
                                    <div className="mt-2 flex items-baseline space-x-2">
                                        <span className="text-4xl text-white font-mono font-bold tracking-tight">{metrics.intentAgeP99}<span className="text-lg text-gray-500 font-normal">ms</span></span>
                                        <span className="text-xs text-emerald-500 flex items-center bg-emerald-500/10 px-1.5 py-0.5 rounded">
                                            <ArrowDown className="w-3 h-3 mr-0.5" /> 12%
                                        </span>
                                    </div>
                                </div>
                                <div className="text-right">
                                    <span className="text-xs font-mono text-gray-500 block">SLO</span>
                                    <span className="text-xs font-mono text-emerald-400 font-medium">&lt; 200ms</span>
                                </div>
                            </div>
                            <div className="mt-6 h-12 w-full flex items-end space-x-1">
                                <div className="w-full h-full flex items-end space-x-[2px] opacity-80">
                                    {[40, 55, 30, 65, 50, 45, 80, 60, 55, 40, 35, 45].map((h, i) => (
                                        <div key={i} className={`w-full ${i === 11 ? 'bg-primary hover:bg-primary/80' : 'bg-primary/20 hover:bg-primary/60'} transition-colors rounded-t-sm`} style={{ height: `${h}%` }}></div>
                                    ))}
                                </div>
                            </div>
                        </div>
                        {/* Reconciliation Success */}
                        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col justify-between shadow-lg shadow-black/20 group hover:border-white/10 transition-all">
                            <div className="flex justify-between items-start mb-4">
                                <h3 className="text-gray-400 text-sm font-medium">Reconciliation</h3>
                                <span className="text-xs font-mono text-gray-500">Last 1h</span>
                            </div>
                            <div className="flex items-center space-x-6">
                                <div className="relative w-24 h-24 flex-shrink-0">
                                    <div className="w-full h-full rounded-full" style={{ background: `conic-gradient(#10b981 0% ${metrics.reconciliationSuccessRate}%, #f59e0b ${metrics.reconciliationSuccessRate}% 98%, #ef4444 98% 100%)`, maskImage: "radial-gradient(transparent 60%, black 61%)", WebkitMaskImage: "radial-gradient(transparent 60%, black 61%)" }}></div>
                                    <div className="absolute inset-0 flex items-center justify-center flex-col">
                                        <span className="text-xl font-bold text-white font-mono">{metrics.reconciliationSuccessRate}</span>
                                        <span className="text-[10px] text-gray-500 uppercase">%</span>
                                    </div>
                                </div>
                                <div className="flex flex-col space-y-2 flex-1 relative">
                                    <div className="flex justify-between items-center text-xs">
                                        <div className="flex items-center"><span className="w-2 h-2 rounded-full bg-emerald-500 mr-2"></span>Success</div>
                                        <span className="font-mono text-white">{formatNumber(metrics.successfulReconciliations)}</span>
                                    </div>
                                    <div className="flex justify-between items-center text-xs">
                                        <div className="flex items-center"><span className="w-2 h-2 rounded-full bg-amber-500 mr-2"></span>Retrying</div>
                                        <span className="font-mono text-white">124</span>
                                    </div>
                                    <div className="flex justify-between items-center text-xs">
                                        <div className="flex items-center"><span className="w-2 h-2 rounded-full bg-red-500 mr-2"></span>Failed</div>
                                        <span className="font-mono text-white">{formatNumber(metrics.failedReconciliations)}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        {/* Pending Intent Volume */}
                        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col justify-between shadow-lg shadow-black/20 group hover:border-white/10 transition-all">
                            <div className="flex justify-between items-start">
                                <h3 className="text-gray-400 text-sm font-medium">Pending Intent Volume</h3>
                                <div className="bg-primary/10 text-primary text-[10px] px-2 py-0.5 rounded font-mono">HIGH THRUPUT</div>
                            </div>
                            <div className="flex-1 flex items-end mt-4 relative overflow-hidden rounded">
                                <div className="absolute inset-0 bg-gradient-to-t from-primary/20 to-transparent opacity-50"></div>
                                <svg className="w-full h-full" preserveAspectRatio="none" viewBox="0 0 100 40">
                                    <path d="M0 35 Q 10 30, 20 32 T 40 25 T 60 20 T 80 28 T 100 15" fill="none" stroke="#6324eb" strokeWidth="2"></path>
                                    <path d="M0 35 Q 10 30, 20 32 T 40 25 T 60 20 T 80 28 T 100 15 V 40 H 0 Z" fill="url(#grad1)" opacity="0.4"></path>
                                    <defs>
                                        <linearGradient id="grad1" x1="0%" x2="0%" y1="0%" y2="100%">
                                            <stop offset="0%" style={{ stopColor: "#6324eb", stopOpacity: 1 }}></stop>
                                            <stop offset="100%" style={{ stopColor: "#6324eb", stopOpacity: 0 }}></stop>
                                        </linearGradient>
                                    </defs>
                                </svg>
                                <div className="absolute bottom-2 left-2 text-white font-mono text-2xl font-bold">{formatNumber(metrics.activeIntents)}</div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ROW 2: Control Plane Health & AI Insights */}
                <section className="grid grid-cols-1 lg:grid-cols-4 gap-6">
                    {/* Control Plane Column (Spans 3) */}
                    <div className="lg:col-span-3 space-y-3">
                        <div className="flex items-center space-x-2">
                            <Workflow className="w-4 h-4 text-blue-400" />
                            <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Control Plane Health</h2>
                        </div>
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                            {/* Scheduler Queue Depth */}
                            <div className="bg-ops-card border border-white/5 rounded-lg p-5 relative overflow-hidden group">
                                <div className="absolute top-0 right-0 w-16 h-16 bg-gradient-to-br from-white/5 to-transparent rounded-bl-3xl"></div>
                                <h3 className="text-gray-400 text-sm font-medium mb-1">Scheduler Queue</h3>
                                <div className="flex items-baseline space-x-2 mb-4">
                                    <span className="text-3xl text-white font-mono font-bold">{metrics.queueDepth}</span>
                                    <span className="text-xs text-gray-500">items</span>
                                </div>
                                <div className="w-full bg-gray-800 rounded-full h-1.5 mb-2 overflow-hidden">
                                    <div className="bg-blue-500 h-1.5 rounded-full" style={{ width: `${Math.min(((metrics.schedulerQueueDepth || 0) / 500) * 100, 100)}%` }}></div>
                                </div>
                                <div className="flex justify-between text-[10px] text-gray-500 font-mono">
                                    <span>0</span>
                                    <span>Cap: 500</span>
                                </div>
                            </div>
                            {/* Worker Saturation */}
                            <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col items-center justify-center relative">
                                <h3 className="absolute top-4 left-5 text-gray-400 text-sm font-medium">Worker Saturation</h3>
                                <div className="relative w-32 h-16 mt-6 overflow-hidden">
                                    <div className="absolute top-0 left-0 w-full h-32 rounded-full border-[12px] border-gray-800 box-border"></div>
                                    <div
                                        className={`absolute top-0 left-0 w-full h-32 rounded-full border-[12px] box-border transition-all duration-1000 ${metrics.workerSaturation > 80 ? 'border-red-500' : 'border-emerald-500'}`}
                                        style={{
                                            clipPath: "polygon(0 0, 100% 0, 100% 50%, 0 50%)",
                                            transform: `rotate(${metrics.workerSaturation * 1.8 - 90}deg)`,
                                            borderLeftColor: "transparent",
                                            borderBottomColor: "transparent",
                                            borderRightColor: "transparent"
                                        }}
                                    ></div>
                                </div>
                                <div className="mt-[-10px] text-center">
                                    <div className="text-2xl font-mono font-bold text-white">{metrics.workerSaturation}%</div>
                                    <div className={`text-xs font-medium px-2 py-0.5 rounded-full mt-1 inline-block ${metrics.workerSaturation > 80 ? 'bg-red-500/10 text-red-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                                        {metrics.workerSaturation > 80 ? 'High Load' : 'Healthy'}
                                    </div>
                                </div>
                            </div>
                            {/* Circuit Breaker Timeline */}
                            <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col">
                                <h3 className="text-gray-400 text-sm font-medium mb-4">Circuit Breaker</h3>
                                <div className="flex-1 flex flex-col justify-center space-y-4">
                                    <div className="flex space-x-0.5 h-6 w-full">
                                        {[1, 2, 3, 4, 5, 6].map(i => (
                                            <div key={i} className={`flex-1 border-b-2 transition-colors ${i === 4 && metrics.circuitBreakerStatus === 'Open' ? 'bg-red-500/20 border-red-500' : 'bg-emerald-500/20 border-emerald-500'}`}></div>
                                        ))}
                                    </div>
                                    <div className="flex justify-between text-[10px] text-gray-500 font-mono uppercase">
                                        <div className="flex items-center"><span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${metrics.circuitBreakerStatus === 'Closed' ? 'bg-emerald-500 animate-pulse' : 'bg-gray-600'}`}></span>Closed</div>
                                        <div className="flex items-center"><span className="w-1.5 h-1.5 bg-amber-500 rounded-full mr-1.5 opacity-50"></span>Degraded</div>
                                        <div className="flex items-center"><span className={`w-1.5 h-1.5 rounded-full mr-1.5 ${metrics.circuitBreakerStatus === 'Open' ? 'bg-red-500 animate-pulse' : 'bg-gray-600'}`}></span>Open</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    {/* AI Insights Column */}
                    <div className="lg:col-span-1 flex flex-col space-y-3 h-full">
                        <div className="flex items-center space-x-2">
                            <Sparkles className="w-4 h-4 text-primary" />
                            <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">AI Insights</h2>
                        </div>
                        <div className="bg-ops-card border border-primary/20 rounded-lg p-0 flex-1 flex flex-col overflow-hidden relative">
                            <div className="bg-gradient-to-r from-primary/20 to-transparent px-4 py-3 border-b border-white/5">
                                <h3 className="text-primary text-xs font-bold uppercase tracking-wider flex items-center">
                                    <Activity className="w-4 h-4 mr-2 animate-pulse" /> Automated Analysis
                                </h3>
                            </div>
                            <div className="p-4 space-y-3 overflow-y-auto custom-scrollbar">
                                <div className="bg-white/5 rounded p-3 border-l-2 border-amber-500 hover:bg-white/10 transition-colors cursor-pointer">
                                    <div className="flex justify-between items-start mb-1">
                                        <span className="text-xs text-amber-400 font-medium">Anomaly Detected</span>
                                        <span className="text-[10px] text-gray-500 font-mono">12m ago</span>
                                    </div>
                                    <p className="text-xs text-gray-300 leading-snug">Unusual spike in DB connection pool retries in <span className="font-mono text-gray-400 bg-black/20 px-1 rounded">us-east-1</span>.</p>
                                </div>
                                <div className="bg-white/5 rounded p-3 border-l-2 border-blue-500 hover:bg-white/10 transition-colors cursor-pointer">
                                    <div className="flex justify-between items-start mb-1">
                                        <span className="text-xs text-blue-400 font-medium">Optimization</span>
                                        <span className="text-[10px] text-gray-500 font-mono">1h ago</span>
                                    </div>
                                    <p className="text-xs text-gray-300 leading-snug">Worker node utilization steady at 45%. Consider scaling down 2 nodes to optimize cost.</p>
                                    <div className="mt-2 flex space-x-2">
                                        <button className="text-[10px] bg-blue-500/20 text-blue-300 px-2 py-1 rounded hover:bg-blue-500/30">Apply</button>
                                        <button className="text-[10px] bg-transparent text-gray-500 px-2 py-1 hover:text-gray-300">Dismiss</button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>

                {/* ROW 3: Infrastructure Health */}
                <section>
                    <div className="flex items-center space-x-2 mb-3">
                        <Server className="w-4 h-4 text-gray-400" />
                        <h2 className="text-xs font-semibold text-gray-400 uppercase tracking-widest">Infrastructure Health</h2>
                    </div>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        {/* Mini Cards */}
                        <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                            <div>
                                <div className="text-[10px] text-gray-500 uppercase font-medium mb-1">Leader Transitions</div>
                                <div className="text-xl font-mono text-white font-medium">0</div>
                            </div>
                            <div className="h-8 w-16 flex items-end space-x-0.5">
                                {[1, 2, 1, 1, 1].map((h, i) => <div key={i} className={`w-1 bg-gray-700 h-${h === 2 ? '2' : '1'}`}></div>)}
                            </div>
                        </div>
                        <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                            <div>
                                <div className="text-[10px] text-gray-500 uppercase font-medium mb-1">Redis Latency</div>
                                <div className="text-xl font-mono text-white font-medium">0.8<span className="text-xs text-gray-500 ml-1">ms</span></div>
                            </div>
                            <div className="h-8 w-16 flex items-end space-x-0.5">
                                {[20, 30, 50, 80, 40, 20].map((h, i) => <div key={i} style={{ height: h + '%' }} className="w-1 bg-emerald-500/40"></div>)}
                            </div>
                        </div>
                        <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                            <div>
                                <div className="text-[10px] text-gray-500 uppercase font-medium mb-1">Lease Health</div>
                                <div className="text-xl font-mono text-emerald-400 font-medium flex items-center">
                                    <CheckCircle className="w-4 h-4 mr-1" /> OK
                                </div>
                            </div>
                            <div className="w-8 h-8 rounded-full border-2 border-emerald-500/20 border-t-emerald-500 animate-spin"></div>
                        </div>
                        <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                            <div>
                                <div className="text-[10px] text-gray-500 uppercase font-medium mb-1">Epoch Drift</div>
                                <div className="text-xl font-mono text-white font-medium">12<span className="text-xs text-gray-500 ml-1">ms</span></div>
                            </div>
                            <div className="h-8 w-16 flex items-end space-x-0.5 opacity-50">
                                {[100, 75, 50, 66, 100].map((h, i) => <div key={i} style={{ height: h + '%' }} className="w-1 bg-blue-500"></div>)}
                            </div>
                        </div>
                    </div>
                </section>
                <footer className="mt-8 border-t border-white/5 pt-4 text-center">
                    <p className="text-[10px] text-gray-600 font-mono">FluxForge Ops Central v4.2.0 · Region: us-west-2 · Normal Mode</p>
                </footer>
            </main>
        </div>
    );
};
