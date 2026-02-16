import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts';

interface WorkerSaturationCardProps {
    saturation: number; // 0 to 100
}

export const WorkerSaturationCard: React.FC<WorkerSaturationCardProps> = ({ saturation }) => {
    const data = [
        { value: saturation },
        { value: 100 - saturation },
    ];

    let status = 'Healthy';
    let color = '#10b981'; // ops-success
    if (saturation > 70) {
        status = 'Busy';
        color = '#f59e0b'; // ops-warning
    }
    if (saturation > 90) {
        status = 'Overload';
        color = '#ef4444'; // ops-danger
    }

    // Custom rotation for gauge effect
    const startAngle = 180;
    const endAngle = 0;

    return (
        <div className="bg-ops-card border border-white/5 rounded-lg p-5 flex flex-col items-center justify-center relative">
            <h3 className="absolute top-4 left-5 text-gray-400 text-sm font-medium">Worker Saturation</h3>
            <div className="relative w-32 h-16 mt-6 overflow-hidden">
                <ResponsiveContainer width="100%" height="200%">
                    <PieChart>
                        <Pie
                            data={data}
                            cx="50%"
                            cy="50%"
                            startAngle={startAngle}
                            endAngle={endAngle}
                            innerRadius="70%"
                            outerRadius="100%"
                            stroke="none"
                            dataKey="value"
                        >
                            <Cell fill={color} />
                            <Cell fill="#1f2937" /> {/* gray-800 */}
                        </Pie>
                    </PieChart>
                </ResponsiveContainer>
            </div>
            <div className="mt-[-10px] text-center z-10">
                <div className="text-2xl font-mono font-bold text-white">{saturation.toFixed(0)}%</div>
                <div
                    className="text-xs font-medium px-2 py-0.5 rounded-full mt-1 inline-block"
                    style={{ color: color, backgroundColor: `${color}1A` }} // 10% opacity
                >
                    {status}
                </div>
            </div>
        </div>
    );
};
