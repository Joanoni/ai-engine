import { useState, useEffect, useRef, useCallback } from 'react';
import type { EngineEvent } from '../types/events';

interface UseWebSocketReturn {
  events: EngineEvent[];
  isConnected: boolean;
  isRunning: boolean;
  connectionStatus: string;
  latency: number | null;
  sessionStartTime: Date | null;
  sendMessage: (text: string) => void;
  clearEvents: () => void;
}

const WS_URL = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080/ws';
const MAX_RECONNECT_ATTEMPTS = 10;
const RECONNECT_INTERVAL_MS = 3000;

export function useWebSocket(): UseWebSocketReturn {
  const [events, setEvents] = useState<EngineEvent[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState('Connecting...');
  const [latency, setLatency] = useState<number | null>(null);
  const [sessionStartTime, setSessionStartTime] = useState<Date | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const pingTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const pingTimestamp = useRef<number | null>(null);
  const isMounted = useRef(true);

  const clearPingTimer = () => {
    if (pingTimer.current) {
      clearInterval(pingTimer.current);
      pingTimer.current = null;
    }
  };

  const clearReconnectTimer = () => {
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
  };

  const connect = useCallback(() => {
    if (!isMounted.current) return;

    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      if (!isMounted.current) return;
      reconnectAttempts.current = 0;
      setIsConnected(true);
      setConnectionStatus('Connected');

      // Start ping interval to measure latency
      clearPingTimer();
      pingTimer.current = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
          pingTimestamp.current = Date.now();
          try {
            ws.send(JSON.stringify({ event: 'ping' }));
          } catch {
            // ignore
          }
        }
      }, 5000);
    };

    ws.onclose = () => {
      if (!isMounted.current) return;
      setIsConnected(false);
      setIsRunning(false);
      clearPingTimer();

      if (reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts.current += 1;
        setConnectionStatus(
          `Reconnecting (attempt ${reconnectAttempts.current}/${MAX_RECONNECT_ATTEMPTS})...`
        );
        clearReconnectTimer();
        reconnectTimer.current = setTimeout(connect, RECONNECT_INTERVAL_MS);
      } else {
        setConnectionStatus('Disconnected');
      }
    };

    ws.onerror = () => {
      if (!isMounted.current) return;
      setIsConnected(false);
    };

    ws.onmessage = (event: MessageEvent) => {
      if (!isMounted.current) return;
      try {
        const raw = JSON.parse(event.data as string);

        // Handle pong for latency measurement
        if (raw.event === 'pong' || raw.type === 'pong') {
          if (pingTimestamp.current !== null) {
            setLatency(Date.now() - pingTimestamp.current);
            pingTimestamp.current = null;
          }
          return;
        }

        const engineEvent: EngineEvent = {
          ...raw,
          receivedAt: new Date().toISOString(),
        };

        // Measure latency from server timestamp if available
        if (raw.timestamp && pingTimestamp.current === null) {
          const serverTs = new Date(raw.timestamp as string).getTime();
          if (!isNaN(serverTs)) {
            setLatency(Math.abs(Date.now() - serverTs));
          }
        }

        if (engineEvent.type === 'session.started') {
          setIsRunning(true);
          setSessionStartTime(new Date());
        } else if (engineEvent.type === 'session.finished' || engineEvent.type === 'error') {
          setIsRunning(false);
        }

        setEvents((prev) => [...prev, engineEvent]);
      } catch {
        // ignore malformed messages
      }
    };
  }, []);

  useEffect(() => {
    isMounted.current = true;
    connect();

    return () => {
      isMounted.current = false;
      clearPingTimer();
      clearReconnectTimer();
      if (wsRef.current) {
        wsRef.current.onclose = null; // prevent reconnect on intentional close
        wsRef.current.close();
      }
    };
  }, [connect]);

  const sendMessage = useCallback((text: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ event: 'user.message', payload: { text } })
      );
    }
  }, []);

  const clearEvents = useCallback(() => {
    setEvents([]);
    setSessionStartTime(null);
  }, []);

  return {
    events,
    isConnected,
    isRunning,
    connectionStatus,
    latency,
    sessionStartTime,
    sendMessage,
    clearEvents,
  };
}
