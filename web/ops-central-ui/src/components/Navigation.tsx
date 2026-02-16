import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

export const Navigation: React.FC = () => {
    const navigate = useNavigate();
    const location = useLocation();

    const modes = [
        { path: "/", label: "Dashboard", color: "bg-indigo-500", border: "border-indigo-500", text: "text-indigo-500" },
        { path: "/normal", label: "Normal", color: "bg-emerald-500", border: "border-emerald-500", text: "text-emerald-500" },
        { path: "/drain", label: "Drain", color: "bg-amber-500", border: "border-amber-500", text: "text-amber-500" },
        { path: "/freeze", label: "Freeze", color: "bg-red-500", border: "border-red-500", text: "text-red-500" },
        { path: "/recovery", label: "Recovery", color: "bg-teal-500", border: "border-teal-500", text: "text-teal-500" },
    ];

    return (
        <div className="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-[100] bg-gray-900/90 backdrop-blur-md border border-white/10 rounded-full px-4 py-2 shadow-2xl flex space-x-2">
            {modes.map((mode) => {
                const isActive = location.pathname === mode.path;
                return (
                    <button
                        key={mode.path}
                        onClick={() => navigate(mode.path)}
                        className={`px-4 py-1.5 rounded-full text-xs font-mono font-bold transition-all duration-300 ${isActive
                            ? `${mode.color} text-black shadow-[0_0_15px_rgba(0,0,0,0.5)] scale-105`
                            : `bg-transparent text-gray-400 hover:text-white hover:bg-white/5`
                            }`}
                    >
                        {mode.label}
                    </button>
                );
            })}
        </div>
    );
};
