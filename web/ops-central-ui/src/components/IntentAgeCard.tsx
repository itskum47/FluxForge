import React from 'react';
import { AreaChart, Area, ResponsiveContainer } from 'recharts';
import { ArrowDownRight, ArrowUpRight } from 'lucide-react';

interface IntentAgeCardProps {
    value: number; // in milliseconds
    trend: number; // percentage
    history: { value: number }[];
}

export const IntentAgeCard: React.FC<IntentAgeCardProps> = ({ value, trend, history }) => {
    const isImprovement = trend <= 0;

    return (
        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col justify-between shadow-lg shadow-black/20 group hover:border-white/10 transition-all">
            <div className="flex justify-between items-start">
                <div>
                    <h3 className="text-gray-400 text-sm font-medium">Intent Age P99</h3>
                    <div className="mt-2 flex items-baseline space-x-2">
                        <span className="text-4xl text-white font-mono font-bold tracking-tight">
                            {value.toFixed(0)}<span className="text-lg text-gray-500 font-normal">ms</span>
                        </span>
                        <span className={`text-xs flex items-center px-1.5 py-0.5 rounded ${isImprovement ? 'text-emerald-500 bg-emerald-500/10' : 'text-rose-500 bg-rose-500/10'}`}>
                            {isImprovement ? <ArrowDownRight size={10} className="mr-0.5" /> : <ArrowUpRight size={10} className="mr-0.5" />}
                            {Math.abs(trend)}%
                        </span>
                    </div>
                </div>
                <div className="text-right">
                    <span className="text-xs font-mono text-gray-500 block">SLO</span>
                    <span className={`text-xs font-mono font-medium ${value < 200 ? 'text-emerald-400' : 'text-rose-400'}`}>&lt; 200ms</span>
                </div>
            </div>

            <div className="mt-6 h-12 w-full">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={history}>
                        <defs>
                            <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#6324eb" stopOpacity={0.3} />
                                <stop offset="95%" stopColor="#6324eb" stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <Area
                            type="monotone"
                            dataKey="value"
                            stroke="#6324eb"
                            fillOpacity={1}
                            fill="url(#latencyGradient)"
                            strokeWidth={2}
                            isAnimationActive={false}
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};
