import { useState, useEffect, useCallback } from 'react';
import type { EngineEvent } from '../types/events';

export interface SessionMeta {
  id: string;
  prompt: string;
  startedAt: string;
  finishedAt?: string;
  status: 'running' | 'done' | 'error';
}

interface UseSessionHistoryReturn {
  sessions: SessionMeta[];
  activeSessionId: string | null;
  setActiveSessionId: (id: string | null) => void;
  loadSessionEvents: (id: string) => Promise<EngineEvent[]>;
  refreshSessions: () => void;
}

const API_BASE = window.location.origin;

export function useSessionHistory(): UseSessionHistoryReturn {
  const [sessions, setSessions] = useState<SessionMeta[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);

  const refreshSessions = useCallback(() => {
    fetch(`${API_BASE}/sessions`)
      .then((r) => r.json())
      .then((data: SessionMeta[]) => setSessions(data ?? []))
      .catch(() => setSessions([]));
  }, []);

  useEffect(() => {
    refreshSessions();
  }, [refreshSessions]);

  const loadSessionEvents = useCallback(async (id: string): Promise<EngineEvent[]> => {
    try {
      const r = await fetch(`${API_BASE}/sessions/${id}/events`);
      const data = await r.json() as EngineEvent[];
      // Add receivedAt fallback for historical events that have timestamp but no receivedAt
      return (data ?? []).map(e => ({
        ...e,
        receivedAt: e.receivedAt || (e as unknown as { timestamp?: string }).timestamp || '',
      }));
    } catch {
      return [];
    }
  }, []);

  return {
    sessions,
    activeSessionId,
    setActiveSessionId,
    loadSessionEvents,
    refreshSessions,
  };
}
