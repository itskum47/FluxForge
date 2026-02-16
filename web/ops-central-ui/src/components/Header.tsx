import React, { useState, useEffect } from 'react';
import { Clock } from 'lucide-react';

export const Header: React.FC = () => {
    const [time, setTime] = useState(new Date());

    useEffect(() => {
        const timer = setInterval(() => setTime(new Date()), 1000);
        return () => clearInterval(timer);
    }, []);

    return (
        <header className="px-6 py-4 border-b border-white/5 bg-ops-bg/90 backdrop-blur flex justify-between items-center z-40 sticky top-8">
            <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                    <div className="w-8 h-8 rounded bg-gradient-to-br from-primary to-purple-800 flex items-center justify-center text-white font-bold text-lg shadow-lg shadow-primary/20">
                        F
                    </div>
                    <div className="flex flex-col">
                        <h1 className="text-white font-semibold leading-none tracking-tight">FluxForge</h1>
                        <span className="text-[10px] text-gray-500 font-mono tracking-wider uppercase mt-0.5">Ops Central</span>
                    </div>
                </div>
            </div>

            {/* Time Filters */}
            <div className="bg-ops-card border border-white/5 rounded-lg p-1 flex space-x-1">
                <button className="px-3 py-1 text-xs font-medium rounded hover:text-white transition-colors text-gray-400 hover:bg-white/5">15m</button>
                <button className="px-3 py-1 text-xs font-medium rounded bg-primary text-white shadow-sm shadow-primary/20">1h</button>
                <button className="px-3 py-1 text-xs font-medium rounded hover:text-white transition-colors text-gray-400 hover:bg-white/5">6h</button>
                <button className="px-3 py-1 text-xs font-medium rounded hover:text-white transition-colors text-gray-400 hover:bg-white/5">24h</button>
            </div>

            {/* Meta Info */}
            <div className="flex items-center space-x-4 text-xs font-mono text-gray-400">
                <div className="flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-white/5 border border-white/5">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                    <span>Live Stream</span>
                </div>
                <div className="flex items-center space-x-1.5">
                    <Clock size={12} />
                    <span>{time.toUTCString().split(' ')[4]} UTC</span>
                </div>
            </div>
        </header>
    );
};
