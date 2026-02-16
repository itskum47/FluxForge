import { useEffect, useState, useCallback, useRef } from 'react';
import type { DashboardMetrics } from '../services/dashboardService';
import { dashboardService } from '../services/dashboardService';

interface UseDashboardStreamOptions {
    autoConnect?: boolean;
    reconnectInterval?: number;
    maxReconnectAttempts?: number;
}

export function useDashboardStream(options: UseDashboardStreamOptions = {}) {
    const {
        autoConnect = true,
        reconnectInterval = 3000,
        maxReconnectAttempts = 10,
    } = options;

    const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const [error, setError] = useState<Error | null>(null);
    const [reconnectAttempts, setReconnectAttempts] = useState(0);

    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<number | null>(null);
    const shouldReconnectRef = useRef(true);

    const connect = useCallback(() => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
            return;
        }

        try {
            const ws = dashboardService.createWebSocketConnection();
            wsRef.current = ws;

            ws.onopen = () => {
                console.log('Dashboard WebSocket connected');
                setIsConnected(true);
                setError(null);
                setReconnectAttempts(0);
            };

            ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    setMetrics(data);
                } catch (err) {
                    console.error('Failed to parse WebSocket message:', err);
                }
            };

            ws.onerror = (event) => {
                console.error('WebSocket error:', event);
                setError(new Error('WebSocket connection error'));
            };

            ws.onclose = () => {
                console.log('Dashboard WebSocket disconnected');
                setIsConnected(false);
                wsRef.current = null;

                // Attempt reconnection
                if (shouldReconnectRef.current && reconnectAttempts < maxReconnectAttempts) {
                    reconnectTimeoutRef.current = setTimeout(() => {
                        console.log(`Reconnecting... (attempt ${reconnectAttempts + 1}/${maxReconnectAttempts})`);
                        setReconnectAttempts((prev) => prev + 1);
                        connect();
                    }, reconnectInterval);
                } else if (reconnectAttempts >= maxReconnectAttempts) {
                    setError(new Error('Max reconnection attempts reached'));
                }
            };
        } catch (err) {
            console.error('Failed to create WebSocket connection:', err);
            setError(err as Error);
        }
    }, [reconnectAttempts, reconnectInterval, maxReconnectAttempts]);

    const disconnect = useCallback(() => {
        shouldReconnectRef.current = false;
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
        }
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
        setIsConnected(false);
    }, []);

    const reconnect = useCallback(() => {
        disconnect();
        shouldReconnectRef.current = true;
        setReconnectAttempts(0);
        connect();
    }, [connect, disconnect]);

    useEffect(() => {
        if (autoConnect) {
            connect();
        }

        return () => {
            shouldReconnectRef.current = false;
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [autoConnect, connect]);

    return {
        metrics,
        isConnected,
        error,
        reconnectAttempts,
        connect,
        disconnect,
        reconnect,
    };
}
