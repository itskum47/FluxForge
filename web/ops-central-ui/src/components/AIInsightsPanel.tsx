import React from 'react';
import { Sparkles, CheckCircle } from 'lucide-react';

export const AIInsightsPanel: React.FC = () => {
    return (
        <div className="bg-ops-card border border-primary/20 rounded-lg p-0 flex-1 flex flex-col overflow-hidden relative h-full">
            {/* Header with Gradient */}
            <div className="bg-gradient-to-r from-primary/20 to-transparent px-4 py-3 border-b border-white/5">
                <h3 className="text-primary text-xs font-bold uppercase tracking-wider flex items-center">
                    <Sparkles size={14} className="mr-2 animate-pulse" /> Automated Analysis
                </h3>
            </div>

            <div className="p-4 space-y-3 overflow-y-auto flex-1 flex flex-col items-center justify-center text-center">
                {/* Empty State / Healthy State */}
                <div className="bg-white/5 rounded-full p-3 mb-2">
                    <CheckCircle size={24} className="text-emerald-500" />
                </div>
                <p className="text-sm font-medium text-gray-300">System Healthy</p>
                <p className="text-xs text-gray-500 max-w-[200px]">
                    FluxAI is monitoring real-time signals. No anomalies detected in the last 15 minutes.
                </p>

                {/* 
        Example of how an anomaly would look (commented out for now):
        <div className="w-full bg-white/5 rounded p-3 border-l-2 border-amber-500 hover:bg-white/10 transition-colors cursor-pointer text-left">
          <div className="flex justify-between items-start mb-1">
            <span className="text-xs text-amber-400 font-medium">Anomaly Detected</span>
            <span className="text-[10px] text-gray-500 font-mono">12m ago</span>
          </div>
          <p className="text-xs text-gray-300 leading-snug">Unusual spike in DB connection pool retries in <span className="font-mono text-gray-400 bg-black/20 px-1 rounded">us-east-1</span>.</p>
        </div>
        */}
            </div>

            <div className="px-4 py-2 border-t border-white/5 bg-black/20">
                <div className="flex justify-between text-[10px] text-gray-500 font-mono">
                    <span>Last scan: just now</span>
                    <span>FluxAI v2.4</span>
                </div>
            </div>
        </div>
    );
};
