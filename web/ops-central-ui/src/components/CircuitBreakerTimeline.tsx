import React from 'react';

interface CircuitBreakerTimelineProps {
    isOpen: boolean; // true = Open (Red), false = Closed (Green)
    history?: boolean[]; // true = Open, false = Closed
}

export const CircuitBreakerTimeline: React.FC<CircuitBreakerTimelineProps> = ({ isOpen }) => {
    // Simulate history based on current state (in a real app, this would come from a time-series metric)
    // We'll just show "Closed" history for now, with the current tip reflecting state

    const historySegments = Array(20).fill(false); // 20 segments of "Closed"

    return (
        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col">
            <h3 className="text-gray-400 text-sm font-medium mb-4">Circuit Breaker</h3>
            <div className="flex-1 flex flex-col justify-center space-y-4">
                {/* Timeline Visualization */}
                <div className="flex space-x-0.5 h-6 w-full">
                    {historySegments.map((wasOpen, i) => (
                        <div
                            key={i}
                            className={`flex-1 transition-colors ${wasOpen ? 'bg-red-500/20 border-b-2 border-red-500' : 'bg-emerald-500/20 border-b-2 border-emerald-500'}`}
                        ></div>
                    ))}
                    {/* Current State Tip */}
                    <div
                        className={`w-8 transition-colors ${isOpen ? 'bg-red-500/60 border-b-2 border-red-500 animate-pulse' : 'bg-emerald-500/60 border-b-2 border-emerald-500'}`}
                        title={isOpen ? "Open" : "Closed"}
                    ></div>
                </div>

                <div className="flex justify-between text-[10px] text-gray-500 font-mono uppercase">
                    <div className="flex items-center"><span className="w-1.5 h-1.5 bg-emerald-500 rounded-full mr-1.5"></span>Closed</div>
                    <div className="flex items-center"><span className="w-1.5 h-1.5 bg-amber-500 rounded-full mr-1.5"></span>Degraded</div>
                    <div className="flex items-center"><span className="w-1.5 h-1.5 bg-red-500 rounded-full mr-1.5"></span>Open</div>
                </div>
            </div>
        </div>
    );
};
