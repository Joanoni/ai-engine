import { useState, useEffect, useRef, useCallback } from 'react';
import type { EngineEvent } from '../types/events';

interface UseWebSocketReturn {
  events: EngineEvent[];
  isConnected: boolean;
  isRunning: boolean;
  sendMessage: (text: string) => void;
  clearEvents: () => void;
}

const WS_URL = import.meta.env.VITE_WS_URL ?? 'ws://localhost:8080/ws';

export function useWebSocket(): UseWebSocketReturn {
  const [events, setEvents] = useState<EngineEvent[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
    };

    ws.onclose = () => {
      setIsConnected(false);
      setIsRunning(false);
    };

    ws.onerror = () => {
      setIsConnected(false);
      setIsRunning(false);
    };

    ws.onmessage = (event: MessageEvent) => {
      try {
        const raw = JSON.parse(event.data as string);
        const engineEvent: EngineEvent = {
          ...raw,
          receivedAt: new Date().toISOString(),
        };

        if (engineEvent.type === 'session.started') {
          setIsRunning(true);
        } else if (engineEvent.type === 'session.finished' || engineEvent.type === 'error') {
          setIsRunning(false);
        }

        setEvents((prev) => [...prev, engineEvent]);
      } catch {
        // ignore malformed messages
      }
    };

    return () => {
      ws.close();
    };
  }, []);

  const sendMessage = useCallback((text: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ event: 'user.message', payload: { text } })
      );
    }
  }, []);

  const clearEvents = useCallback(() => {
    setEvents([]);
  }, []);

  return { events, isConnected, isRunning, sendMessage, clearEvents };
}
