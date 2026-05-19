import { useState, useCallback, useMemo, useEffect } from 'react';
import { ReactFlowProvider } from '@xyflow/react';
import type { EngineEvent } from './types/events';
import { useWebSocket } from './hooks/useWebSocket';
import { useAgentGraph } from './hooks/useAgentGraph';
import { useSessionHistory } from './hooks/useSessionHistory';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useAnalytics } from './hooks/useAnalytics';
import { Sidebar } from './components/layout/Sidebar';
import { CockpitArea } from './components/layout/CockpitArea';
import { MissionPanel } from './components/layout/MissionPanel';
import { AgentDetailDrawer } from './components/drawers/AgentDetailDrawer';
import { AnalyticsPanel } from './components/analytics/AnalyticsPanel';
import './App.css';

export default function App() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [replayEvents, setReplayEvents] = useState<EngineEvent[] | null>(null);
  const [showAnalytics, setShowAnalytics] = useState(false);
  const [terminalClearedAt, setTerminalClearedAt] = useState<number | null>(null);

  const {
    events: wsEvents,
    isConnected,
    isRunning,
    connectionStatus,
    latency,
    sessionStartTime,
    sendMessage,
    clearEvents,
  } = useWebSocket();

  const {
    sessions,
    activeSessionId,
    setActiveSessionId,
    loadSessionEvents,
    refreshSessions,
  } = useSessionHistory();

  const { sessionLogs, loadingLogs, loadSessionLogs, clearSessionLogs } = useAnalytics();

  // Use replay events if a past session is loaded, otherwise live events
  const events: EngineEvent[] = replayEvents ?? wsEvents;

  const { nodes, edges } = useAgentGraph(events);

  // Refresh session list on terminal WebSocket events
  useEffect(() => {
    const last = wsEvents[wsEvents.length - 1];
    if (!last) return;
    if (last.type === 'session.started') {
      setTimeout(refreshSessions, 300);
    } else if (last.type === 'session.finished' || last.type === 'error') {
      setTimeout(refreshSessions, 500);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsEvents.length]);

  const handleSend = useCallback(
    (text: string) => {
      setReplayEvents(null); // exit replay mode
      sendMessage(text);
      setTimeout(refreshSessions, 500);
    },
    [sendMessage, refreshSessions]
  );

  const handleNewMission = useCallback(() => {
    clearEvents();
    setActiveSessionId(null);
    setReplayEvents(null);
    setSelectedAgent(null);
    setShowAnalytics(false);
    setTerminalClearedAt(null);
  }, [clearEvents, setActiveSessionId]);

  const handleLoadSession = useCallback(async (id: string) => {
    const evts = await loadSessionEvents(id);
    setReplayEvents(evts);
    setActiveSessionId(id);
    setSelectedAgent(null);
    setShowAnalytics(false);
    setTerminalClearedAt(null);
  }, [loadSessionEvents, setActiveSessionId]);

  const handleNodeClick = useCallback((agentName: string) => {
    setSelectedAgent(agentName);
  }, []);

  const handleCloseDrawer = useCallback(() => {
    setSelectedAgent(null);
  }, []);

  const handleClearEvents = useCallback(() => {
    setTerminalClearedAt(Date.now());
  }, []);

  const terminalEvents = useMemo(
    () =>
      terminalClearedAt
        ? events.filter(
            (e) =>
              new Date(e.receivedAt ?? e.timestamp ?? 0).getTime() > terminalClearedAt
          )
        : events,
    [events, terminalClearedAt]
  );

  const handleToggleAnalytics = useCallback(() => {
    setShowAnalytics((prev) => {
      if (prev) clearSessionLogs();
      return !prev;
    });
  }, [clearSessionLogs]);

  const shortcuts = useMemo(
    () => ({
      'ctrl+b': () => setSidebarCollapsed((prev) => !prev),
      'ctrl+shift+a': () => handleToggleAnalytics(),
      escape: () => setSelectedAgent(null),
    }),
    [handleToggleAnalytics]
  );

  useKeyboardShortcuts(shortcuts);

  return (
    <ReactFlowProvider>
      <div className="app-layout">
        <Sidebar
          collapsed={sidebarCollapsed}
          sessions={sessions}
          activeSessionId={activeSessionId}
          onNewMission={handleNewMission}
          onLoadSession={handleLoadSession}
          onToggleAnalytics={handleToggleAnalytics}
          showAnalytics={showAnalytics}
        />

        {showAnalytics ? (
          <AnalyticsPanel
            sessions={sessions}
            loadSessionLogs={loadSessionLogs}
            sessionLogs={sessionLogs}
            loadingLogs={loadingLogs}
          />
        ) : (
          <CockpitArea
            nodes={nodes}
            edges={edges}
            events={events}
            terminalEvents={terminalEvents}
            onNodeClick={handleNodeClick}
            onClearEvents={handleClearEvents}
          />
        )}

        <MissionPanel
          events={events}
          isConnected={isConnected}
          isRunning={isRunning}
          connectionStatus={connectionStatus}
          latency={latency}
          sessionStartTime={sessionStartTime}
          onSend={handleSend}
        />

        <AgentDetailDrawer
          agent={selectedAgent}
          events={events}
          onClose={handleCloseDrawer}
        />
      </div>
    </ReactFlowProvider>
  );
}
