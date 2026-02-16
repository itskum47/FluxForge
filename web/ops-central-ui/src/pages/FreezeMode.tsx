import React from 'react';
import {
    ShieldAlert,
    KeyRound,
    LayoutDashboard,
    Disc,
    Workflow,
    BarChart2,
    Settings,
    LockOpen,
    TimerOff,
    ArrowUp,
    Activity,
    Terminal,
    Filter,
    RotateCcw,
    Zap
} from 'lucide-react';

export const FreezeMode: React.FC = () => {
    return (
        <div className="bg-background-dark text-slate-200 font-display h-screen flex flex-col overflow-hidden selection:bg-critical selection:text-white">
            {/* Emergency Banner */}
            <div className="bg-critical text-white px-4 py-2 flex items-center justify-center font-bold text-sm tracking-wider uppercase shadow-lg shadow-critical/20 z-50 animate-pulse-slow">
                <ShieldAlert className="w-5 h-5 mr-2" />
                Emergency Freeze • All admissions blocked
                <span className="ml-4 bg-white/20 px-2 py-0.5 rounded text-xs">INC-2094</span>
            </div>

            <div className="flex flex-1 overflow-hidden relative">
                <div className="scanline"></div>
                {/* Sidebar */}
                <aside className="w-16 lg:w-20 bg-[#1e172d] border-r border-slate-800/60 flex flex-col items-center py-6 gap-6 z-10">
                    <div className="w-10 h-10 bg-primary/20 rounded-lg flex items-center justify-center text-primary mb-4">
                        <KeyRound className="w-6 h-6" />
                    </div>
                    <nav className="flex flex-col gap-4 w-full px-2">
                        <a className="w-full aspect-square rounded-lg flex items-center justify-center text-slate-400 hover:bg-slate-800 transition-colors" href="#"><LayoutDashboard className="w-6 h-6" /></a>
                        <a className="w-full aspect-square rounded-lg flex items-center justify-center bg-critical/10 text-critical border border-critical/50 shadow-[0_0_10px_rgba(255,59,59,0.3)]" href="#"><Disc className="w-6 h-6" /></a>
                        <a className="w-full aspect-square rounded-lg flex items-center justify-center text-slate-400 hover:bg-slate-800 transition-colors" href="#"><Workflow className="w-6 h-6" /></a>
                        <a className="w-full aspect-square rounded-lg flex items-center justify-center text-slate-400 hover:bg-slate-800 transition-colors" href="#"><BarChart2 className="w-6 h-6" /></a>
                    </nav>
                    <div className="mt-auto flex flex-col gap-4 w-full px-2">
                        <a className="w-full aspect-square rounded-lg flex items-center justify-center text-slate-400 hover:bg-slate-800 transition-colors" href="#"><Settings className="w-6 h-6" /></a>
                        <div className="w-8 h-8 rounded-full bg-slate-700 mx-auto"></div>
                    </div>
                </aside>

                {/* Content Area */}
                <main className="flex-1 overflow-y-auto w-full relative pb-20 custom-scrollbar">
                    <div className="absolute inset-0 z-0 opacity-[0.03] grid-bg"></div>

                    {/* Header */}
                    <header className="px-6 py-4 flex items-center justify-between border-b border-slate-800/60 bg-[#161121]/80 backdrop-blur-sm sticky top-0 z-20">
                        <div className="flex items-center gap-4">
                            <h1 className="text-xl font-bold tracking-tight text-white flex items-center gap-3">
                                Ops Central <span className="px-2 py-0.5 rounded text-xs font-mono bg-slate-800 text-slate-400 border border-slate-700">v4.2.0</span>
                            </h1>
                            <div className="h-6 w-px bg-slate-700 mx-2"></div>
                            <div className="flex items-center gap-2 text-critical font-mono text-sm font-bold animate-pulse">
                                <span className="w-2 h-2 rounded-full bg-critical"></span> MODE: FREEZE
                            </div>
                        </div>
                        <div className="flex items-center gap-4">
                            <button className="flex items-center gap-2 px-4 py-2 bg-critical text-white rounded font-medium text-sm hover:bg-critical-dark transition-colors shadow-[0_0_15px_rgba(255,59,59,0.4)]">
                                <LockOpen className="w-4 h-4" /> Override Freeze
                            </button>
                        </div>
                    </header>

                    <div className="p-6 grid grid-cols-1 md:grid-cols-12 gap-6 max-w-[1920px] mx-auto z-10 relative">
                        {/* ROW 1: Key Metrics */}
                        {/* Intent Age - Halted */}
                        <div className="col-span-12 md:col-span-4 lg:col-span-3 bg-[#1e172d] border-2 border-critical rounded-lg p-5 relative overflow-hidden group">
                            <div className="absolute inset-0 bg-critical/5 z-0"></div>
                            <div className="absolute -right-6 -top-6 w-24 h-24 bg-critical/20 rounded-full blur-2xl"></div>
                            <div className="relative z-10 flex justify-between items-start mb-4">
                                <div>
                                    <p className="text-slate-400 text-xs font-mono uppercase tracking-wider mb-1">Intent Age P99</p>
                                    <div className="flex items-center gap-1 text-critical text-xs">
                                        <ArrowUp className="w-4 h-4" /><span>+Infinity%</span>
                                    </div>
                                </div>
                                <TimerOff className="text-critical/80 w-5 h-5" />
                            </div>
                            <div className="relative z-10 mt-2 flex items-baseline gap-2">
                                <h2 className="text-5xl font-black text-critical tracking-tighter">HALTED</h2>
                            </div>
                            <div className="mt-4 pt-3 border-t border-critical/20 flex justify-between items-center text-xs font-mono text-slate-400">
                                <span>Last admission: 45m ago</span>
                                <span className="text-critical font-bold">STALLED</span>
                            </div>
                        </div>
                        {/* Reconciliation - Stalled */}
                        <div className="col-span-12 md:col-span-4 lg:col-span-3 bg-[#1e172d] border border-slate-700 rounded-lg p-5 relative overflow-hidden">
                            <div className="flex justify-between items-start mb-2">
                                <p className="text-slate-400 text-xs font-mono uppercase tracking-wider">Reconciliation Loop</p>
                                <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-slate-800 text-slate-400 border border-slate-600">IDLE</span>
                            </div>
                            <div className="mt-6 mb-2">
                                <div className="flex justify-between text-sm mb-1 font-mono">
                                    <span className="text-slate-300">Throughput</span>
                                    <span className="text-critical font-bold">0 ops/s</span>
                                </div>
                                <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden relative">
                                    <div className="absolute inset-0 w-full h-full" style={{ backgroundImage: "repeating-linear-gradient(45deg, #FF3B3B 0, #FF3B3B 10px, #8B0000 10px, #8B0000 20px)", opacity: 0.3 }}></div>
                                </div>
                            </div>
                            <div className="grid grid-cols-2 gap-2 mt-4">
                                <div className="bg-slate-900/50 p-2 rounded border border-slate-800">
                                    <span className="block text-[10px] text-slate-500 font-mono">Backlog</span>
                                    <span className="block text-lg font-mono text-white">4,291</span>
                                </div>
                                <div className="bg-slate-900/50 p-2 rounded border border-slate-800">
                                    <span className="block text-[10px] text-slate-500 font-mono">Drift</span>
                                    <span className="block text-lg font-mono text-critical">Critical</span>
                                </div>
                            </div>
                        </div>
                        {/* Worker Saturation - Maxed */}
                        <div className="col-span-12 md:col-span-4 lg:col-span-3 bg-[#1e172d] border border-critical/50 shadow-[0_0_15px_rgba(255,59,59,0.15)] rounded-lg p-5 flex flex-col items-center justify-center relative">
                            <p className="absolute top-5 left-5 text-slate-400 text-xs font-mono uppercase tracking-wider w-full">Worker Saturation</p>
                            <div className="relative w-32 h-32 mt-4">
                                <svg className="w-full h-full transform -rotate-90">
                                    <circle className="text-slate-800" cx="64" cy="64" fill="transparent" r="56" stroke="currentColor" strokeWidth="8"></circle>
                                    <circle className="text-critical drop-shadow-[0_0_8px_rgba(255,59,59,0.6)]" cx="64" cy="64" fill="transparent" r="56" stroke="currentColor" strokeDasharray="351.86" strokeDashoffset="0" strokeWidth="8"></circle>
                                </svg>
                                <div className="absolute inset-0 flex flex-col items-center justify-center">
                                    <span className="text-3xl font-bold text-white">100%</span>
                                    <span className="text-[10px] text-critical font-bold uppercase mt-1">Overload</span>
                                </div>
                            </div>
                            <div className="mt-2 text-center">
                                <span className="text-xs text-slate-400 font-mono">128/128 Nodes Unresponsive</span>
                            </div>
                        </div>
                        {/* AI Insights */}
                        <div className="col-span-12 lg:col-span-3 flex flex-col gap-4">
                            <div className="flex-1 bg-gradient-to-br from-critical/20 to-slate-900 border border-critical rounded-lg p-5 flex flex-col justify-between relative overflow-hidden">
                                <div className="absolute top-0 right-0 p-4 opacity-10"><Activity className="w-16 h-16 text-critical" /></div>
                                <div>
                                    <div className="flex items-center gap-2 mb-3">
                                        <span className="w-2 h-2 bg-critical rounded-full animate-ping"></span>
                                        <h3 className="text-critical font-bold text-sm uppercase tracking-wide">AI Insight: Critical</h3>
                                    </div>
                                    <p className="text-white text-sm font-medium leading-relaxed">Emergency freeze detected. Cascade failure predicted in <span className="font-mono text-critical">00:03:00</span> without intervention.</p>
                                </div>
                                <button className="mt-4 w-full py-2 bg-critical/10 hover:bg-critical/20 border border-critical text-critical text-xs font-bold uppercase rounded transition-all flex items-center justify-center gap-2">
                                    Execute Protocol 9 <Zap className="w-4 h-4" />
                                </button>
                            </div>
                        </div>

                        {/* Pending Intents Grid */}
                        <div className="col-span-12 lg:col-span-8 bg-[#1e172d] border border-slate-700 rounded-lg flex flex-col">
                            <div className="px-5 py-4 border-b border-slate-700 flex justify-between items-center bg-slate-800/30">
                                <div className="flex items-center gap-3">
                                    <h3 className="font-semibold text-slate-200">Pending Intents Queue</h3>
                                    <span className="px-2 py-0.5 rounded-full bg-slate-800 border border-slate-600 text-xs text-slate-400 font-mono">4,203 Locked</span>
                                </div>
                                <div className="flex gap-2">
                                    <button className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"><Filter className="w-4 h-4" /></button>
                                    <button className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded transition-colors"><RotateCcw className="w-4 h-4" /></button>
                                </div>
                            </div>
                            <div className="overflow-x-auto">
                                <table className="w-full text-left text-sm font-mono">
                                    <thead className="bg-slate-900/50 text-slate-500 uppercase text-xs">
                                        <tr>
                                            <th className="px-5 py-3 font-medium">ID</th>
                                            <th className="px-5 py-3 font-medium">Service</th>
                                            <th className="px-5 py-3 font-medium">Region</th>
                                            <th className="px-5 py-3 font-medium">Age</th>
                                            <th className="px-5 py-3 font-medium text-right">Status</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-slate-800/50">
                                        <tr className="hover:bg-slate-800/30 transition-colors group">
                                            <td className="px-5 py-3 text-slate-400">#INT-9921</td>
                                            <td className="px-5 py-3 text-slate-200">Payment-Gateway-v2</td>
                                            <td className="px-5 py-3 text-slate-400">us-east-1</td>
                                            <td className="px-5 py-3 text-critical">42m 10s</td>
                                            <td className="px-5 py-3 text-right">
                                                <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-critical/10 text-critical border border-critical/30 text-[10px] font-bold uppercase">
                                                    <Activity className="w-3 h-3" /> Frozen
                                                </span>
                                            </td>
                                        </tr>
                                        <tr className="hover:bg-slate-800/30 transition-colors group">
                                            <td className="px-5 py-3 text-slate-400">#INT-9920</td>
                                            <td className="px-5 py-3 text-slate-200">Auth-Service-Core</td>
                                            <td className="px-5 py-3 text-slate-400">eu-west-3</td>
                                            <td className="px-5 py-3 text-critical">41m 55s</td>
                                            <td className="px-5 py-3 text-right">
                                                <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-critical/10 text-critical border border-critical/30 text-[10px] font-bold uppercase">
                                                    <Activity className="w-3 h-3" /> Frozen
                                                </span>
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </div>
                        </div>

                        {/* Global Map Placeholder */}
                        <div className="col-span-12 lg:col-span-4 bg-[#1e172d] border border-slate-700 rounded-lg flex flex-col p-5 overflow-hidden">
                            <div className="flex justify-between items-center mb-4">
                                <h3 className="text-xs font-mono uppercase text-slate-400 tracking-wider">Global Impact</h3>
                                <div className="flex gap-2">
                                    <div className="w-2 h-2 rounded-full bg-critical animate-pulse"></div>
                                    <span className="text-[10px] text-critical font-bold">ALL ZONES DOWN</span>
                                </div>
                            </div>
                            <div className="flex-1 relative rounded-lg overflow-hidden bg-slate-900 border border-slate-800 group h-40">
                                <div className="absolute inset-0 opacity-40" style={{ backgroundImage: "radial-gradient(#334155 1px, transparent 1px)", backgroundSize: "20px 20px" }}></div>
                                <div className="absolute inset-0 flex items-center justify-center">
                                    <div className="bg-black/70 backdrop-blur px-4 py-2 border border-critical/50 rounded text-critical font-mono text-xs font-bold shadow-xl">CONNECTION LOST</div>
                                </div>
                            </div>
                        </div>

                        {/* Logs */}
                        <div className="col-span-12 bg-[#0d0a14] border border-slate-800 rounded-lg p-4 font-mono text-xs h-48 overflow-y-auto custom-scrollbar">
                            <div className="flex items-center gap-2 mb-2 sticky top-0 bg-[#0d0a14] pb-2 border-b border-slate-800 w-full z-10">
                                <Terminal className="w-4 h-4 text-slate-500" />
                                <span className="text-slate-400 font-bold">System Logs</span>
                                <span className="ml-auto text-critical text-[10px] uppercase animate-pulse">Live Stream</span>
                            </div>
                            <div className="flex flex-col gap-1">
                                <div className="text-slate-500"><span className="text-slate-600">[10:42:01]</span> <span className="text-blue-400">INFO</span> Initiating freeze protocol v3...</div>
                                <div className="text-slate-400"><span className="text-slate-600">[10:42:02]</span> <span className="text-yellow-500">WARN</span> Admission controller unresponsive. Force lock applied.</div>
                                <div className="text-critical"><span className="text-slate-600">[10:42:02]</span> <span className="bg-critical text-white px-1 font-bold">CRIT</span> RECONCILIATION LOOP BROKEN. PID 49202</div>
                                <div className="text-slate-300 opacity-50 border-l-2 border-slate-700 pl-2 ml-1 mt-1">&gt; Waiting for manual intervention... <span className="animate-pulse">_</span></div>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        </div>
    );
};
