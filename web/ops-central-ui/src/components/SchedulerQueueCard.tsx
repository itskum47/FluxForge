import React from 'react';

interface SchedulerQueueCardProps {
    depth: number;
    cap: number;
}

export const SchedulerQueueCard: React.FC<SchedulerQueueCardProps> = ({ depth, cap }) => {
    const percentage = Math.min((depth / cap) * 100, 100);

    // Color logic based on saturation
    let colorClass = 'bg-blue-500';
    if (percentage > 80) colorClass = 'bg-amber-500';
    if (percentage > 95) colorClass = 'bg-red-500';

    return (
        <div className="bg-ops-card border border-white/5 rounded-lg p-5 relative overflow-hidden group">
            <div className="absolute top-0 right-0 w-16 h-16 bg-gradient-to-br from-white/5 to-transparent rounded-bl-3xl"></div>
            <h3 className="text-gray-400 text-sm font-medium mb-1">Scheduler Queue</h3>
            <div className="flex items-baseline space-x-2 mb-4">
                <span className="text-3xl text-white font-mono font-bold">{depth}</span>
                <span className="text-xs text-gray-500">items</span>
            </div>
            <div className="w-full bg-gray-800 rounded-full h-1.5 mb-2 overflow-hidden">
                <div
                    className={`h-1.5 rounded-full transition-all duration-500 ${colorClass}`}
                    style={{ width: `${percentage}%` }}
                ></div>
            </div>
            <div className="flex justify-between text-[10px] text-gray-500 font-mono">
                <span>0</span>
                <span>Cap: {cap}</span>
            </div>
        </div>
    );
};
