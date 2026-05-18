import { useState, useCallback } from 'react';
import type { SessionLogs } from '../types/logs';

export function useAnalytics() {
  const [sessionLogs, setSessionLogs] = useState<SessionLogs | null>(null);
  const [loadingLogs, setLoadingLogs] = useState(false);

  const loadSessionLogs = useCallback(async (sessionId: string): Promise<SessionLogs | null> => {
    setLoadingLogs(true);
    try {
      const res = await fetch(`/sessions/${sessionId}/logs`);
      if (!res.ok) return null;
      const data: SessionLogs = await res.json();
      setSessionLogs(data);
      return data;
    } catch {
      return null;
    } finally {
      setLoadingLogs(false);
    }
  }, []);

  const clearSessionLogs = useCallback(() => {
    setSessionLogs(null);
  }, []);

  return { sessionLogs, loadingLogs, loadSessionLogs, clearSessionLogs };
}
