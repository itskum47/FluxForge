import React from 'react';
import {
    TriangleAlert,
    X,
    Zap,
    LayoutDashboard,
    Server,
    LineChart,
    Terminal,
    Bell,
    Play,
    Briefcase,
    Layers,
    CircleAlert,
    Gauge,
    Shuffle,
    Lightbulb,
    Timer,
    Database,
    Cpu,
    Router,
    Clock,
    ShieldCheck
} from 'lucide-react';

export const DrainMode: React.FC = () => {
    return (
        <div className="bg-background-dark text-slate-200 font-display antialiased flex flex-col h-screen overflow-hidden selection:bg-amber-500/30 selection:text-amber-200">
            {/* Global Sticky Banner: Drain Mode */}
            <div className="w-full bg-amber-500 text-black font-semibold text-sm py-2 px-4 flex items-center justify-center shadow-lg relative z-50 animate-pulse">
                <TriangleAlert className="w-5 h-5 mr-2" />
                <span>SYSTEM IN DRAIN MODE • TRAFFIC IS BEING REDIRECTED TO SECONDARY CLUSTER</span>
                <button className="absolute right-4 top-1/2 -translate-y-1/2 bg-black/20 hover:bg-black/30 p-1 rounded transition-colors">
                    <X className="w-4 h-4" />
                </button>
            </div>

            <div className="flex flex-1 overflow-hidden">
                {/* Sidebar Navigation */}
                <aside className="w-16 lg:w-20 flex-shrink-0 bg-background-dark border-r border-white/5 flex flex-col items-center py-6 z-40">
                    <div className="mb-8 w-10 h-10 rounded-lg bg-primary flex items-center justify-center shadow-lg shadow-primary/20">
                        <Zap className="text-white w-6 h-6" />
                    </div>
                    <nav className="flex-1 flex flex-col gap-6 w-full items-center">
                        <a className="p-3 rounded-lg bg-white/5 text-white relative group" href="#"><LayoutDashboard className="w-6 h-6" /></a>
                        <a className="p-3 rounded-lg text-slate-400 hover:text-white hover:bg-white/5 transition-colors relative group" href="#"><Server className="w-6 h-6" /></a>
                        <a className="p-3 rounded-lg text-slate-400 hover:text-white hover:bg-white/5 transition-colors relative group" href="#"><LineChart className="w-6 h-6" /></a>
                        <a className="p-3 rounded-lg text-slate-400 hover:text-white hover:bg-white/5 transition-colors relative group" href="#"><Terminal className="w-6 h-6" /></a>
                    </nav>
                    <div className="flex flex-col gap-4 mb-2">
                        <a className="p-3 rounded-lg text-amber-500 hover:text-amber-400 hover:bg-amber-500/10 transition-colors" href="#"><Bell className="w-6 h-6" /></a>
                        <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-primary to-purple-400 border border-white/20"></div>
                    </div>
                </aside>

                {/* Main Content Area */}
                <main className="flex-1 flex flex-col min-w-0 bg-background-dark relative">
                    {/* Background Grid Pattern */}
                    <div className="absolute inset-0 z-0 opacity-[0.03] grid-bg"></div>

                    {/* Top Header */}
                    <header className="h-16 border-b border-white/5 flex items-center justify-between px-6 z-10 bg-background-dark/80 backdrop-blur-sm">
                        <div className="flex items-center gap-4">
                            <h1 className="text-lg font-medium text-white tracking-tight">FluxForge Ops Central</h1>
                            <span className="text-slate-600 text-sm font-light">/</span>
                            <div className="flex items-center gap-2 px-2 py-1 rounded bg-amber-500/10 border border-amber-500/20">
                                <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></span>
                                <span className="text-xs font-mono text-amber-500 uppercase tracking-wider">DRAINING US-EAST-1</span>
                            </div>
                        </div>
                        <div className="flex items-center gap-4">
                            <div className="flex items-center gap-2 text-xs text-slate-400 font-mono bg-surface-dark px-3 py-1.5 rounded border border-white/5">
                                <span className="text-green-400">●</span> API: 99.9%
                                <span className="mx-2 text-slate-700">|</span>
                                <span className="text-amber-400">●</span> DB: 45ms
                            </div>
                            <button className="bg-primary hover:bg-primary/90 text-white text-xs font-medium px-4 py-2 rounded transition-colors flex items-center gap-2">
                                <Play className="w-4 h-4" /> Resume Traffic
                            </button>
                        </div>
                    </header>

                    {/* Dashboard Content */}
                    <div className="flex-1 overflow-y-auto p-6 z-10 pb-20 custom-scrollbar">
                        {/* Status Row */}
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                            {/* Worker Saturation Card */}
                            <div className="glass-panel p-5 rounded-lg border border-amber-500/30 relative overflow-hidden group amber-glow">
                                <div className="absolute top-0 right-0 p-3 opacity-10">
                                    <Briefcase className="w-16 h-16 text-amber-500" />
                                </div>
                                <div className="flex justify-between items-start mb-4">
                                    <div className="flex flex-col">
                                        <span className="text-slate-400 text-xs font-medium uppercase tracking-wider">Worker Saturation</span>
                                        <span className="bg-amber-500 text-black text-[10px] font-bold px-1.5 py-0.5 rounded w-fit mt-1">DRAINING</span>
                                    </div>
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className="text-3xl font-mono text-white">12%</span>
                                    <span className="text-xs text-emerald-400 mb-1 flex items-center">
                                        <span className="text-sm mr-0.5">↓</span> 45%
                                    </span>
                                </div>
                                <div className="text-xs text-amber-500/80 mt-2 font-mono">Active connections closing...</div>
                                <div className="h-1 w-full bg-white/5 mt-4 rounded-full overflow-hidden">
                                    <div className="h-full bg-amber-500 w-[12%] rounded-full shadow-[0_0_10px_rgba(245,158,11,0.5)]"></div>
                                </div>
                            </div>

                            {/* Queue Depth Card */}
                            <div className="glass-panel p-5 rounded-lg relative overflow-hidden">
                                <div className="flex justify-between items-start mb-4">
                                    <span className="text-slate-400 text-xs font-medium uppercase tracking-wider">Queue Depth</span>
                                    <Layers className="text-slate-600 w-5 h-5" />
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className="text-3xl font-mono text-white">4,502</span>
                                </div>
                                <div className="text-xs text-slate-500 mt-2 flex items-center gap-1">
                                    <span className="w-2 h-2 rounded-full bg-amber-500 inline-block"></span> Items currently draining
                                </div>
                                <div className="mt-4 flex items-end gap-0.5 h-8 opacity-50">
                                    {[60, 50, 70, 80].map((h, i) => <div key={`q-${i}`} className="w-1 bg-slate-700 rounded-sm" style={{ height: h + '%' }}></div>)}
                                    {[60, 40, 20, 10].map((h, i) => <div key={`qd-${i}`} className="w-1 bg-amber-500/50 rounded-sm" style={{ height: h + '%' }}></div>)}
                                </div>
                            </div>

                            {/* Error Rate */}
                            <div className="glass-panel p-5 rounded-lg">
                                <div className="flex justify-between items-start mb-4">
                                    <span className="text-slate-400 text-xs font-medium uppercase tracking-wider">Error Rate</span>
                                    <CircleAlert className="text-slate-600 w-5 h-5" />
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className="text-3xl font-mono text-white">0.02%</span>
                                    <span className="text-xs text-slate-500 mb-1">Stable</span>
                                </div>
                                <div className="text-xs text-slate-500 mt-2">Within acceptable thresholds</div>
                                <div className="h-1 w-full bg-white/5 mt-4 rounded-full overflow-hidden">
                                    <div className="h-full bg-emerald-500 w-[2%] rounded-full"></div>
                                </div>
                            </div>

                            {/* Throughput */}
                            <div className="glass-panel p-5 rounded-lg">
                                <div className="flex justify-between items-start mb-4">
                                    <span className="text-slate-400 text-xs font-medium uppercase tracking-wider">Global RPS</span>
                                    <Gauge className="text-slate-600 w-5 h-5" />
                                </div>
                                <div className="flex items-end gap-2">
                                    <span className="text-3xl font-mono text-white">12.4k</span>
                                </div>
                                <div className="text-xs text-amber-500 mt-2 flex items-center gap-1">
                                    <Shuffle className="w-4 h-4" /> Redirecting to EU-WEST
                                </div>
                                <div className="h-8 mt-4 w-full flex items-end gap-px opacity-40">
                                    {[100, 90, 85].map((h, i) => <div key={i} className="flex-1 bg-slate-600 rounded-t-sm" style={{ height: h + '%' }}></div>)}
                                    {[60, 40, 20, 10].map((h, i) => <div key={i} className="flex-1 bg-amber-600 rounded-t-sm" style={{ height: h + '%' }}></div>)}
                                </div>
                            </div>
                        </div>

                        {/* Main Layout: Insights & Charts */}
                        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-auto lg:h-[400px]">
                            {/* Left: AI Insights Panel (Drain Specific) */}
                            <div className="glass-panel p-6 rounded-lg lg:col-span-1 flex flex-col border border-amber-500/20">
                                <div className="flex items-center gap-2 mb-4 border-b border-white/5 pb-4">
                                    <Lightbulb className="text-amber-400 w-5 h-5" />
                                    <h3 className="text-white font-medium">Ops AI Analysis</h3>
                                    <span className="ml-auto text-xs px-2 py-0.5 rounded-full bg-primary/20 text-primary border border-primary/20">LIVE</span>
                                </div>
                                <div className="flex-1 flex flex-col gap-4 overflow-y-auto pr-2 custom-scrollbar">
                                    <div className="bg-amber-500/10 border border-amber-500/30 p-4 rounded-lg">
                                        <h4 className="text-amber-400 text-sm font-semibold mb-1 flex items-center gap-2">
                                            <Timer className="w-4 h-4" /> Drain Estimation
                                        </h4>
                                        <p className="text-slate-300 text-sm leading-relaxed mb-3">
                                            Drain mode initiated by operator <span className="text-white font-mono bg-white/10 px-1 rounded text-xs">j.doe</span>. Estimated time until full clearance: <span className="text-amber-400 font-mono">4m 12s</span> based on current throughput decay.
                                        </p>
                                        <div className="w-full bg-black/40 h-1.5 rounded-full overflow-hidden">
                                            <div className="h-full bg-amber-500 w-[78%]"></div>
                                        </div>
                                    </div>
                                    <div className="bg-surface-dark border border-white/5 p-4 rounded-lg">
                                        <h4 className="text-slate-200 text-sm font-semibold mb-1">Load Balancer Logic</h4>
                                        <p className="text-slate-400 text-xs leading-relaxed">
                                            Automatic scale-in policies are paused during drain. New instances in <span className="text-emerald-400">Secondary (EU-West)</span> are scaling up to absorb traffic shift.
                                        </p>
                                    </div>
                                </div>
                            </div>

                            {/* Right: Traffic Volume Chart */}
                            <div className="glass-panel p-6 rounded-lg lg:col-span-2 flex flex-col">
                                <div className="flex justify-between items-center mb-6">
                                    <div>
                                        <h3 className="text-white font-medium">Traffic Distribution Shift</h3>
                                        <p className="text-xs text-slate-500 mt-1">Real-time request routing comparison</p>
                                    </div>
                                    <div className="flex items-center gap-4 text-xs">
                                        <div className="flex items-center gap-2">
                                            <span className="w-3 h-3 rounded bg-amber-500/50 border border-amber-500"></span>
                                            <span className="text-slate-300">US-East (Draining)</span>
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <span className="w-3 h-3 rounded bg-emerald-500/50 border border-emerald-500"></span>
                                            <span className="text-slate-300">EU-West (Active)</span>
                                        </div>
                                    </div>
                                </div>
                                {/* Chart Visualization */}
                                <div className="flex-1 w-full relative h-64 border-l border-b border-white/10">
                                    <div className="absolute inset-0 flex flex-col justify-between pointer-events-none">
                                        {[...Array(5)].map((_, i) => <div key={i} className="w-full h-px bg-white/5"></div>)}
                                    </div>
                                    <div className="absolute inset-0 flex items-end justify-between px-2 pb-px pt-4 gap-1">
                                        <div className="w-full h-full flex items-end gap-1">
                                            {[
                                                { e: 10, a: 80 }, { e: 15, a: 75 }, { e: 25, a: 70 }, { e: 40, a: 50 },
                                                { e: 55, a: 35 }, { e: 70, a: 20 }, { e: 85, a: 10 }, { e: 95, a: 5 }
                                            ].map((data, i) => (
                                                <div key={i} className="flex-1 flex flex-col justify-end h-full gap-0.5">
                                                    <div className="w-full bg-emerald-500/50 rounded-t-sm" style={{ height: data.e + '%' }}></div>
                                                    <div className="w-full bg-amber-500/50 rounded-b-sm" style={{ height: data.a + '%' }}></div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* System Health Grid */}
                        <div className="mt-6">
                            <h3 className="text-slate-400 text-xs font-medium uppercase tracking-wider mb-4">Service Mesh Health</h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
                                <div className="glass-panel p-4 rounded-lg flex items-center justify-between border-l-2 border-l-amber-500 bg-amber-500/5">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center border border-white/10"><Database className="text-amber-500 w-4 h-4" /></div>
                                        <div><div className="text-sm text-slate-200 font-medium">Postgres Primary</div><div className="text-xs text-amber-500">Connections Draining</div></div>
                                    </div>
                                    <span className="text-xs font-mono text-slate-400">12/500</span>
                                </div>
                                <div className="glass-panel p-4 rounded-lg flex items-center justify-between border-l-2 border-l-emerald-500">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center border border-white/10"><Cpu className="text-emerald-500 w-4 h-4" /></div>
                                        <div><div className="text-sm text-slate-200 font-medium">Redis Cache</div><div className="text-xs text-emerald-500">Healthy</div></div>
                                    </div>
                                    <span className="text-xs font-mono text-slate-400">1.2ms</span>
                                </div>
                                <div className="glass-panel p-4 rounded-lg flex items-center justify-between border-l-2 border-l-amber-500 bg-amber-500/5">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center border border-white/10"><Router className="text-amber-500 w-4 h-4" /></div>
                                        <div><div className="text-sm text-slate-200 font-medium">Ingress Gateway</div><div className="text-xs text-amber-500">Shedding Load</div></div>
                                    </div>
                                    <span className="text-xs font-mono text-slate-400">403s: 1.5%</span>
                                </div>
                                <div className="glass-panel p-4 rounded-lg flex items-center justify-between border-l-2 border-l-slate-600 opacity-60">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center border border-white/10"><Clock className="text-slate-400 w-4 h-4" /></div>
                                        <div><div className="text-sm text-slate-200 font-medium">Cron Workers</div><div className="text-xs text-slate-400">Suspended</div></div>
                                    </div>
                                    <span className="text-xs font-mono text-slate-400">--</span>
                                </div>
                                <div className="glass-panel p-4 rounded-lg flex items-center justify-between border-l-2 border-l-emerald-500">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded bg-surface-dark flex items-center justify-center border border-white/10"><ShieldCheck className="text-emerald-500 w-4 h-4" /></div>
                                        <div><div className="text-sm text-slate-200 font-medium">Auth Service</div><div className="text-xs text-emerald-500">Healthy</div></div>
                                    </div>
                                    <span className="text-xs font-mono text-slate-400">99.99%</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        </div>
    );
};
