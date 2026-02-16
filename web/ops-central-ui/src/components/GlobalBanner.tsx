import React from 'react';
import type { AdmissionMode } from '../types';
import clsx from 'clsx';

interface GlobalBannerProps {
    mode: AdmissionMode;
    isConnected: boolean;
}

export const GlobalBanner: React.FC<GlobalBannerProps> = ({ mode, isConnected }) => {
    const getBannerConfig = () => {
        if (!isConnected) {
            return {
                bg: 'bg-gray-800/50',
                border: 'border-gray-700',
                text: 'text-gray-400',
                dot: 'bg-gray-500',
                label: 'Disconnected: Reconnecting...',
            };
        }

        switch (mode.toLowerCase()) {
            case 'normal':
            case 'pilot':
                return {
                    bg: 'bg-ops-success/10',
                    border: 'border-ops-success/20',
                    text: 'text-ops-success',
                    dot: 'bg-ops-success',
                    label: 'Normal: Connected · Stable',
                };
            case 'drain':
                return {
                    bg: 'bg-ops-warning/10',
                    border: 'border-ops-warning/20',
                    text: 'text-ops-warning',
                    dot: 'bg-ops-warning',
                    label: 'Drain Mode: Rejecting New Work',
                };
            case 'freeze':
                return {
                    bg: 'bg-ops-danger/10',
                    border: 'border-ops-danger/20',
                    text: 'text-ops-danger',
                    dot: 'bg-ops-danger',
                    label: 'Freeze Mode: System Halted',
                };
            default:
                return {
                    bg: 'bg-gray-800',
                    border: 'border-gray-700',
                    text: 'text-gray-400',
                    dot: 'bg-gray-400',
                    label: 'Unknown State',
                };
        }
    };

    const config = getBannerConfig();

    return (
        <div className={clsx(
            'w-full py-1.5 px-6 flex items-center justify-center sticky top-0 z-50 backdrop-blur-sm transition-colors duration-500',
            config.bg,
            'border-b',
            config.border
        )}>
            <div className={clsx(
                'flex items-center space-x-2 text-xs font-mono font-medium tracking-wide',
                config.text
            )}>
                <span className="relative flex h-2 w-2">
                    {isConnected && (
                        <span className={clsx(
                            'animate-ping absolute inline-flex h-full w-full rounded-full opacity-75',
                            config.dot
                        )}></span>
                    )}
                    <span className={clsx(
                        'relative inline-flex rounded-full h-2 w-2',
                        config.dot
                    )}></span>
                </span>
                <span className="uppercase">{config.label}</span>
            </div>
        </div>
    );
};
