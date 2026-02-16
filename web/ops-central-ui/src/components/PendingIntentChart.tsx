import React from 'react';
import { AreaChart, Area, ResponsiveContainer, XAxis, YAxis } from 'recharts';

interface PendingIntentChartProps {
    volume: number;
    history: { timestamp: number; value: number }[];
    isHighThroughput?: boolean;
}

export const PendingIntentChart: React.FC<PendingIntentChartProps> = ({ volume, history, isHighThroughput = false }) => {
    return (
        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col justify-between shadow-lg shadow-black/20 group hover:border-white/10 transition-all">
            <div className="flex justify-between items-start">
                <h3 className="text-gray-400 text-sm font-medium">Pending Intent Volume</h3>
                {isHighThroughput && (
                    <div className="bg-primary/10 text-primary text-[10px] px-2 py-0.5 rounded font-mono">HIGH THRUPUT</div>
                )}
            </div>

            <div className="flex-1 flex items-end mt-4 relative overflow-hidden rounded min-h-[100px]">
                <div className="absolute inset-0 w-full h-full">
                    <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={history}>
                            <defs>
                                <linearGradient id="volumeGradient" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="5%" stopColor="#6324eb" stopOpacity={0.5} />
                                    <stop offset="95%" stopColor="#6324eb" stopOpacity={0} />
                                </linearGradient>
                            </defs>
                            <XAxis dataKey="timestamp" hide />
                            <YAxis hide domain={['auto', 'auto']} />
                            <Area
                                type="basis"
                                dataKey="value"
                                stroke="#6324eb"
                                strokeWidth={2}
                                fillOpacity={1}
                                fill="url(#volumeGradient)"
                                isAnimationActive={false}
                            />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>
                <div className="absolute bottom-2 left-2 text-white font-mono text-2xl font-bold z-10 pointer-events-none">
                    {volume.toLocaleString()}
                </div>
            </div>
        </div>
    );
};
