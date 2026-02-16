import React from 'react';
import { Activity, Database, CheckCircle, Clock } from 'lucide-react';

interface LeadershipPanelProps {
    transitions: number;
    redisLatency: number;
    epochDrift: number;
}

export const LeadershipPanel: React.FC<LeadershipPanelProps> = ({ transitions, redisLatency, epochDrift }) => {
    return (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Mini Card: Leader Transitions */}
            <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                <div>
                    <div className="text-[10px] text-gray-500 uppercase font-medium mb-1 flex items-center">
                        <Activity size={10} className="mr-1" /> Leader Transitions
                    </div>
                    <div className="text-xl font-mono text-white font-medium">{transitions}</div>
                </div>
                <div className="h-8 w-16 flex items-end space-x-0.5">
                    <div className="w-1 bg-gray-700 h-1"></div>
                    <div className="w-1 bg-gray-700 h-2"></div>
                    <div className="w-1 bg-gray-700 h-1"></div>
                    <div className="w-1 bg-gray-700 h-1"></div>
                    <div className="w-1 bg-gray-700 h-1"></div>
                </div>
            </div>

            {/* Mini Card: Redis Latency */}
            <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                <div>
                    <div className="text-[10px] text-gray-500 uppercase font-medium mb-1 flex items-center">
                        <Database size={10} className="mr-1" /> Redis Latency
                    </div>
                    <div className="text-xl font-mono text-white font-medium">
                        {(redisLatency * 1000).toFixed(1)}<span className="text-xs text-gray-500 ml-1">ms</span>
                    </div>
                </div>
                <div className="h-8 w-16 flex items-end space-x-0.5">
                    {/* Simulated Bars */}
                    <div className="w-1 bg-emerald-500/40 h-2"></div>
                    <div className="w-1 bg-emerald-500/40 h-3"></div>
                    <div className="w-1 bg-emerald-500/40 h-5"></div>
                    <div className="w-1 bg-emerald-500/60 h-8"></div>
                    <div className="w-1 bg-emerald-500/40 h-4"></div>
                    <div className="w-1 bg-emerald-500/40 h-2"></div>
                </div>
            </div>

            {/* Mini Card: Lease Health */}
            <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                <div>
                    <div className="text-[10px] text-gray-500 uppercase font-medium mb-1 flex items-center">
                        <CheckCircle size={10} className="mr-1" /> Lease Health
                    </div>
                    <div className="text-xl font-mono text-emerald-400 font-medium flex items-center">
                        OK
                    </div>
                </div>
                <div className="w-8 h-8 rounded-full border-2 border-emerald-500/20 border-t-emerald-500 animate-spin"></div>
            </div>

            {/* Mini Card: Epoch Drift */}
            <div className="bg-ops-card border border-white/5 rounded-lg p-4 flex items-center justify-between hover:border-white/10 transition-all">
                <div>
                    <div className="text-[10px] text-gray-500 uppercase font-medium mb-1 flex items-center">
                        <Clock size={10} className="mr-1" /> Epoch Drift
                    </div>
                    <div className="text-xl font-mono text-white font-medium">
                        {epochDrift}<span className="text-xs text-gray-500 ml-1">ms</span>
                    </div>
                </div>
                <div className="h-8 w-16 flex items-end space-x-0.5 opacity-50">
                    <div className="w-1 bg-blue-500 h-full"></div>
                    <div className="w-1 bg-blue-500 h-3/4"></div>
                    <div className="w-1 bg-blue-500 h-1/2"></div>
                    <div className="w-1 bg-blue-500 h-2/3"></div>
                    <div className="w-1 bg-blue-500 h-full"></div>
                </div>
            </div>
        </div>
    );
};
