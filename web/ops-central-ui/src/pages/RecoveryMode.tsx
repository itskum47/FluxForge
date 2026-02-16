import React from 'react';
import {
    ShieldCheck,
    Infinity,
    LayoutDashboard,
    HeartPulse,
    Sparkles,
    Zap,
    ArrowUp,
    CheckCircle,
    Timer
} from 'lucide-react';

export const RecoveryMode: React.FC = () => {
    return (
        <div className="bg-background-dark text-slate-200 font-display antialiased selection:bg-recovery-500 selection:text-white h-screen flex flex-col overflow-hidden">
            {/* Global Sticky Banner */}
            <div className="w-full bg-recovery-900/30 border-b border-recovery-500/30 backdrop-blur-md sticky top-0 z-50">
                <div className="max-w-[1920px] mx-auto px-4 h-10 flex items-center justify-between text-xs font-medium">
                    <div className="flex items-center gap-3">
                        <span className="flex h-2 w-2 relative">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-recovery-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-recovery-500"></span>
                        </span>
                        <span className="text-recovery-400 uppercase tracking-wider font-bold">Recovery in Progress — Phased Re-enable</span>
                        <span className="text-neutral-slate-400 px-2 border-l border-neutral-slate-700 font-mono">INC-2023-994</span>
                    </div>
                    <div className="flex items-center gap-6">
                        <div className="flex items-center gap-2">
                            <span className="text-neutral-slate-400">Traffic Ramp:</span>
                            <div className="w-24 h-1.5 bg-neutral-slate-700 rounded-full overflow-hidden">
                                <div className="h-full bg-recovery-500 w-[45%] relative overflow-hidden">
                                    <div className="absolute inset-0 bg-white/20 w-full h-full animate-pulse"></div>
                                </div>
                            </div>
                            <span className="text-recovery-400 font-mono">45%</span>
                        </div>
                        <div className="flex items-center gap-2 text-neutral-slate-400">
                            <Timer className="w-4 h-4" />
                            <span className="font-mono">T+00:14:22</span>
                        </div>
                    </div>
                </div>
            </div>

            <div className="flex flex-1 overflow-hidden">
                {/* Sidebar */}
                <aside className="w-16 bg-surface-darker border-r border-neutral-slate-700 flex flex-col items-center py-6 gap-6 z-40">
                    <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center shadow-lg shadow-primary/20 mb-4">
                        <Infinity className="text-white w-6 h-6" />
                    </div>
                    <nav className="flex flex-col gap-4 w-full px-2">
                        <button className="w-10 h-10 mx-auto rounded-lg bg-neutral-slate-800 text-neutral-slate-400 hover:text-white hover:bg-neutral-slate-700 transition-colors flex items-center justify-center group relative"><LayoutDashboard className="w-5 h-5" /></button>
                        <button className="w-10 h-10 mx-auto rounded-lg bg-recovery-900/50 text-recovery-400 border border-recovery-500/30 flex items-center justify-center relative shadow-[0_0_15px_rgba(20,184,166,0.1)]"><HeartPulse className="w-5 h-5" /><span className="absolute top-1 right-1 w-2 h-2 bg-recovery-500 rounded-full"></span></button>
                        <button className="w-10 h-10 mx-auto rounded-lg bg-neutral-slate-800 text-neutral-slate-400 hover:text-white hover:bg-neutral-slate-700 transition-colors flex items-center justify-center"><Sparkles className="w-5 h-5" /></button>
                    </nav>
                </aside>

                {/* Main Content */}
                <main className="flex-1 overflow-y-auto bg-background-dark relative pb-20 custom-scrollbar">
                    <div className="absolute inset-0 opacity-[0.03] pointer-events-none grid-bg"></div>
                    <div className="max-w-[1920px] mx-auto p-4 lg:p-6 space-y-4">
                        {/* Header */}
                        <div className="flex items-end justify-between mb-2">
                            <div>
                                <h1 className="text-2xl font-bold text-white tracking-tight font-display">FluxForge Ops Central</h1>
                                <p className="text-recovery-400 text-sm flex items-center gap-2 mt-1">
                                    <ShieldCheck className="w-4 h-4" /> System Health: Recovering
                                </p>
                            </div>
                            <div className="flex gap-3">
                                <button className="bg-surface-dark border border-neutral-slate-700 text-sm text-neutral-slate-400 px-4 py-2 rounded-lg hover:bg-neutral-slate-700 transition-colors">View Logs</button>
                                <button className="bg-primary/10 border border-primary/40 text-sm text-primary-300 px-4 py-2 rounded-lg hover:bg-primary/20 transition-colors flex items-center gap-2"><Zap className="w-4 h-4" /> Runbook Actions</button>
                            </div>
                        </div>

                        {/* Row 1: Metrics */}
                        <div className="grid grid-cols-1 lg:grid-cols-4 gap-4 h-auto lg:h-48">
                            {/* Intent Age */}
                            <div className="bg-surface-dark border border-recovery-500/20 rounded-xl p-5 flex flex-col justify-between relative overflow-hidden group hover:border-recovery-500/40 transition-colors">
                                <div className="absolute -right-6 -top-6 w-24 h-24 bg-recovery-500/10 rounded-full blur-2xl group-hover:bg-recovery-500/20 transition-all"></div>
                                <div className="flex justify-between items-start z-10">
                                    <h3 className="text-neutral-slate-400 text-xs font-semibold uppercase tracking-wider">Intent Age</h3>
                                    <span className="bg-recovery-900/40 text-recovery-400 text-[10px] font-bold px-2 py-0.5 rounded border border-recovery-500/20">STABILIZING</span>
                                </div>
                                <div className="z-10 mt-2">
                                    <div className="flex items-baseline gap-2">
                                        <span className="text-4xl font-mono font-bold text-white">124<span className="text-lg text-neutral-slate-400">ms</span></span>
                                        <span className="text-recovery-400 text-xs font-mono flex items-center"><ArrowUp className="w-3 h-3 rotate-180" /> 40%</span>
                                    </div>
                                    <p className="text-neutral-slate-500 text-xs mt-1">Global p99 Latency</p>
                                </div>
                                <div className="h-12 w-full mt-4 flex items-end gap-1">
                                    {[80, 90, 100, 70, 50, 40, 35, 30, 25, 22].map((h, i) => (
                                        <div key={i} className={`w-1 rounded-sm ${i < 4 ? 'bg-red-500/20' : i === 4 ? 'bg-neutral-slate-700' : 'bg-recovery-500/' + (30 + (i - 5) * 10)}`} style={{ height: h + '%' }}></div>
                                    ))}
                                    <div className="flex-1 relative h-full ml-2">
                                        <svg className="w-full h-full overflow-visible" preserveAspectRatio="none">
                                            <path d="M0 10 L10 5 L20 40 L30 35 L40 45 L50 48" fill="none" stroke="#2dd4bf" strokeLinecap="round" strokeWidth="2" vectorEffect="non-scaling-stroke"></path>
                                            <circle className="animate-pulse" cx="50" cy="48" fill="#2dd4bf" r="3"></circle>
                                        </svg>
                                    </div>
                                </div>
                            </div>

                            {/* Reconciliation Chart */}
                            <div className="bg-surface-dark border border-neutral-slate-700 rounded-xl p-5 flex flex-col justify-between col-span-1 lg:col-span-2 relative">
                                <div className="flex justify-between items-start mb-2">
                                    <div>
                                        <h3 className="text-neutral-slate-400 text-xs font-semibold uppercase tracking-wider">Reconciliation Success</h3>
                                        <p className="text-2xl font-mono font-bold text-white mt-1">99.2% <span className="text-sm font-normal text-recovery-400 ml-2">(Recovering)</span></p>
                                    </div>
                                </div>
                                <div className="flex-1 w-full bg-gradient-to-t from-recovery-900/10 to-transparent rounded-lg relative overflow-hidden flex items-end px-2 pt-6">
                                    <div className="absolute inset-0 flex flex-col justify-between py-2 pointer-events-none opacity-20">
                                        <div className="w-full border-t border-dashed border-neutral-slate-500"></div>
                                        <div className="w-full border-t border-dashed border-neutral-slate-500"></div>
                                        <div className="w-full border-t border-dashed border-neutral-slate-500"></div>
                                    </div>
                                    <svg className="w-full h-[80%] overflow-visible" preserveAspectRatio="none" viewBox="0 0 100 50">
                                        <defs>
                                            <linearGradient id="recoveryGradient" x1="0" x2="0" y1="0" y2="1">
                                                <stop offset="0%" stopColor="#2dd4bf" stopOpacity="0.5"></stop>
                                                <stop offset="100%" stopColor="#2dd4bf" stopOpacity="0"></stop>
                                            </linearGradient>
                                        </defs>
                                        <path d="M0 50 L10 45 L20 48 L30 40 L40 30 L50 35 L60 20 L70 15 L80 10 L90 5 L100 2 V50 H0" fill="url(#recoveryGradient)"></path>
                                        <path d="M0 50 L10 45 L20 48 L30 40 L40 30 L50 35 L60 20 L70 15 L80 10 L90 5 L100 2" fill="none" stroke="#2dd4bf" strokeWidth="2" vectorEffect="non-scaling-stroke"></path>
                                    </svg>
                                </div>
                            </div>

                            {/* Error Budget */}
                            <div className="bg-surface-dark border border-neutral-slate-700 rounded-xl p-5 flex flex-col justify-between">
                                <h3 className="text-neutral-slate-400 text-xs font-semibold uppercase tracking-wider">Error Budget Burn</h3>
                                <div className="flex items-center justify-center my-2 relative">
                                    <svg className="w-24 h-24 transform -rotate-90">
                                        <circle cx="48" cy="48" fill="none" r="40" stroke="#2d2a3d" strokeWidth="8"></circle>
                                        <circle cx="48" cy="48" fill="none" r="40" stroke="#fbbf24" strokeDasharray="251.2" strokeDashoffset="100" strokeWidth="8"></circle>
                                    </svg>
                                    <div className="absolute text-center">
                                        <div className="text-xl font-bold text-white font-mono">38%</div>
                                        <div className="text-[10px] text-yellow-500 uppercase">Warning</div>
                                    </div>
                                </div>
                                <div className="text-xs text-neutral-slate-400 text-center">Remaining budget is stabilizing.</div>
                            </div>
                        </div>

                        {/* Row 2: Circuit Breaker & Insights */}
                        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 lg:h-64">
                            <div className="col-span-1 lg:col-span-2 bg-surface-dark border border-neutral-slate-700 rounded-xl p-6 flex flex-col">
                                <div className="flex justify-between items-center mb-6">
                                    <h3 className="text-white font-medium flex items-center gap-2"><Zap className="text-recovery-400 w-5 h-5" /> Circuit Breaker State</h3>
                                    <span className="px-2 py-1 rounded bg-recovery-900/50 text-recovery-400 text-xs border border-recovery-500/30 font-mono">HALF-OPEN / PROBING</span>
                                </div>
                                <div className="flex-1 flex items-center relative px-4">
                                    <div className="absolute left-4 right-4 h-0.5 bg-neutral-slate-700 top-1/2 -translate-y-1/2 z-0"></div>
                                    <div className="w-full flex justify-between items-center relative z-10">
                                        <div className="flex flex-col items-center gap-3">
                                            <div className="w-3 h-3 bg-red-500 rounded-full shadow-[0_0_10px_rgba(239,68,68,0.5)]"></div>
                                            <div className="text-center"><div className="text-xs text-neutral-slate-300 font-mono">14:05</div><div className="text-[10px] text-red-400 uppercase font-bold mt-1">Tripped</div></div>
                                        </div>
                                        <div className="flex flex-col items-center gap-3">
                                            <div className="w-3 h-3 bg-neutral-slate-500 rounded-full border border-neutral-slate-800"></div>
                                            <div className="text-center"><div className="text-xs text-neutral-slate-500 font-mono">14:08</div><div className="text-[10px] text-neutral-slate-500 uppercase font-bold mt-1">Cool Down</div></div>
                                        </div>
                                        <div className="flex flex-col items-center gap-3">
                                            <div className="relative">
                                                <div className="w-4 h-4 bg-recovery-500 rounded-full shadow-[0_0_15px_rgba(45,212,191,0.6)] z-20 relative"></div>
                                                <div className="absolute inset-0 bg-recovery-400 rounded-full animate-ping opacity-75"></div>
                                            </div>
                                            <div className="text-center"><div className="text-xs text-white font-mono">14:22</div><div className="text-[10px] text-recovery-400 uppercase font-bold mt-1">Probing</div></div>
                                        </div>
                                        <div className="flex flex-col items-center gap-3 opacity-30">
                                            <div className="w-3 h-3 bg-neutral-slate-700 rounded-full border-2 border-neutral-slate-600"></div>
                                            <div className="text-center"><div className="text-xs text-neutral-slate-500 font-mono">--:--</div><div className="text-[10px] text-neutral-slate-500 uppercase font-bold mt-1">Closed</div></div>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div className="bg-gradient-to-br from-surface-dark to-surface-darker border border-neutral-slate-700 rounded-xl p-5 flex flex-col relative overflow-hidden">
                                <div className="absolute top-0 right-0 w-32 h-32 bg-primary/10 blur-3xl rounded-full pointer-events-none"></div>
                                <div className="flex items-center gap-2 mb-4">
                                    <Sparkles className="text-primary-300 w-5 h-5" />
                                    <h3 className="text-white font-medium">Auto-Remediation Insights</h3>
                                </div>
                                <div className="flex-1 overflow-y-auto space-y-3 pr-2 custom-scrollbar">
                                    <div className="bg-neutral-slate-800/50 p-3 rounded-lg border-l-2 border-recovery-500">
                                        <p className="text-xs text-neutral-slate-300 leading-relaxed"><span className="font-bold text-recovery-400 block mb-1">Status Update</span>Traffic shaping active. Recommendation: Hold traffic ramp at 50% until DB sync completes.</p>
                                    </div>
                                    <div className="bg-neutral-slate-800/50 p-3 rounded-lg border-l-2 border-neutral-slate-600">
                                        <p className="text-xs text-neutral-slate-400 leading-relaxed"><span className="font-bold text-neutral-slate-300 block mb-1">Previous Action</span>Phase 2 complete. Initiated cache warming.</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Row 3: Zones */}
                        <div>
                            <h3 className="text-neutral-slate-400 text-xs font-semibold uppercase tracking-wider mb-3 px-1">Zone Synchronization Status</h3>
                            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
                                <div className="bg-surface-dark border border-green-500/30 rounded-lg p-4 relative group">
                                    <div className="flex justify-between items-start mb-2"><div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-green-500"></div><span className="text-white font-mono text-sm">US-East-1a</span></div><span className="text-[10px] bg-green-900/30 text-green-400 px-1.5 py-0.5 rounded border border-green-500/20">LEADER</span></div>
                                    <div className="w-full bg-neutral-slate-700 h-1 rounded-full mt-3 overflow-hidden"><div className="bg-green-500 h-full w-full"></div></div>
                                </div>
                                <div className="bg-surface-dark border border-recovery-500 rounded-lg p-4 relative shadow-[0_0_15px_rgba(20,184,166,0.15)]">
                                    <div className="flex justify-between items-start mb-2"><div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-recovery-500 animate-pulse"></div><span className="text-white font-mono text-sm">US-East-1b</span></div><span className="text-[10px] bg-recovery-900/30 text-recovery-400 px-1.5 py-0.5 rounded border border-recovery-500/20">SYNCING</span></div>
                                    <div className="w-full bg-neutral-slate-700 h-1 rounded-full mt-3 overflow-hidden"><div className="bg-recovery-500 h-full w-[65%] relative overflow-hidden"><div className="absolute inset-0 bg-white/30 w-full animate-[shimmer_1s_infinite]"></div></div></div>
                                </div>
                                <div className="bg-surface-dark border border-neutral-slate-700 rounded-lg p-4 opacity-75">
                                    <div className="flex justify-between items-start mb-2"><div className="flex items-center gap-2"><div className="w-2 h-2 rounded-full bg-neutral-slate-500"></div><span className="text-neutral-slate-300 font-mono text-sm">US-East-1c</span></div><span className="text-[10px] bg-neutral-slate-800 text-neutral-slate-400 px-1.5 py-0.5 rounded border border-neutral-slate-600">FOLLOWER</span></div>
                                    <div className="w-full bg-neutral-slate-700 h-1 rounded-full mt-3"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        </div>
    );
};
