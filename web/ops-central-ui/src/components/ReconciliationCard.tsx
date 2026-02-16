import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';

interface ReconciliationCardProps {
    successRate: number;
    successCount: number;
    retryCount: number;
    failCount: number;
}

export const ReconciliationCard: React.FC<ReconciliationCardProps> = ({
    successRate, successCount, retryCount, failCount
}) => {
    const data = [
        { name: 'Success', value: successCount, color: '#10b981' },
        { name: 'Retrying', value: retryCount, color: '#f59e0b' },
        { name: 'Failed', value: failCount, color: '#ef4444' },
    ];

    return (
        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col justify-between shadow-lg shadow-black/20 group hover:border-white/10 transition-all">
            <div className="flex justify-between items-start mb-4">
                <h3 className="text-gray-400 text-sm font-medium">Reconciliation</h3>
                <span className="text-xs font-mono text-gray-500">Last 1h</span>
            </div>

            <div className="flex items-center space-x-6">
                <div className="relative w-24 h-24 flex-shrink-0">
                    <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                            <Pie
                                data={data}
                                cx="50%"
                                cy="50%"
                                innerRadius={38}
                                outerRadius={45}
                                startAngle={90}
                                endAngle={-270}
                                dataKey="value"
                                stroke="none"
                            >
                                {data.map((entry, index) => (
                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                ))}
                            </Pie>
                        </PieChart>
                    </ResponsiveContainer>
                    <div className="absolute inset-0 flex items-center justify-center flex-col">
                        <span className="text-xl font-bold text-white font-mono">{successRate.toFixed(1)}</span>
                        <span className="text-[10px] text-gray-500 uppercase">%</span>
                    </div>
                </div>

                <div className="flex flex-col space-y-2 flex-1">
                    {data.map((item) => (
                        <div key={item.name} className="flex justify-between items-center text-xs">
                            <div className="flex items-center">
                                <span style={{ backgroundColor: item.color }} className="w-2 h-2 rounded-full mr-2"></span>
                                {item.name}
                            </div>
                            <span className="font-mono text-white">{item.value.toLocaleString()}</span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};
