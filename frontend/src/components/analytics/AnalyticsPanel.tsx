import { useState } from 'react';
import type { SessionMeta } from '../../hooks/useSessionHistory';
import type { SessionLogs } from '../../types/logs';
import { ProjectView } from './ProjectView';
import { SessionView } from './SessionView';
import { AgentView } from './AgentView';

type AnalyticsView =
  | { type: 'project' }
  | { type: 'session'; sessionId: string }
  | { type: 'agent'; sessionId: string; agentName: string };

interface AnalyticsPanelProps {
  sessions: SessionMeta[];
  loadSessionLogs: (id: string) => Promise<SessionLogs | null>;
  sessionLogs: SessionLogs | null;
  loadingLogs: boolean;
}

export function AnalyticsPanel({ sessions, loadSessionLogs, sessionLogs, loadingLogs }: AnalyticsPanelProps) {
  const [view, setView] = useState<AnalyticsView>({ type: 'project' });

  const handleSelectSession = async (id: string) => {
    await loadSessionLogs(id);
    setView({ type: 'session', sessionId: id });
  };

  const handleSelectAgent = (agentName: string) => {
    if (view.type === 'session') {
      setView({ type: 'agent', sessionId: view.sessionId, agentName });
    }
  };

  const handleBackToProject = () => {
    setView({ type: 'project' });
  };

  const handleBackToSession = () => {
    if (view.type === 'agent') {
      setView({ type: 'session', sessionId: view.sessionId });
    }
  };

  // Breadcrumb
  const breadcrumb = () => {
    if (view.type === 'project') return 'Analytics / Project';
    if (view.type === 'session') return `Analytics / Session`;
    if (view.type === 'agent') return `Analytics / Session / ${view.agentName}`;
    return 'Analytics';
  };

  return (
    <div
      style={{
        flex: 1,
        height: '100%',
        background: 'var(--bg-base)',
        overflowY: 'auto',
        padding: '24px',
        display: 'flex',
        flexDirection: 'column',
        gap: '0',
      }}
    >
      {/* Header bar */}
      <div style={{ marginBottom: '20px', display: 'flex', alignItems: 'center', gap: '12px' }}>
        <div style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)', letterSpacing: '0.05em' }}>
          {breadcrumb()}
        </div>
        <div style={{ marginLeft: 'auto', display: 'flex', gap: '8px' }}>
          {view.type !== 'project' && (
            <button
              onClick={handleBackToProject}
              style={{ background: 'transparent', border: '1px solid var(--border)', borderRadius: '6px', color: 'var(--text-muted)', cursor: 'pointer', padding: '4px 10px', fontFamily: 'var(--font-ui)', fontSize: '12px' }}
            >
              ⌂ Project
            </button>
          )}
        </div>
      </div>

      {/* View content */}
      {view.type === 'project' && (
        <ProjectView sessions={sessions} onSelectSession={handleSelectSession} />
      )}

      {view.type === 'session' && (
        <SessionView
          sessionId={view.sessionId}
          sessions={sessions}
          logs={sessionLogs}
          loadingLogs={loadingLogs}
          onSelectAgent={handleSelectAgent}
          onBack={handleBackToProject}
        />
      )}

      {view.type === 'agent' && (() => {
        const agentLogs = sessionLogs?.agents[view.agentName] ?? [];
        return (
          <AgentView
            agentName={view.agentName}
            sessionId={view.sessionId}
            logs={agentLogs}
            onBack={handleBackToSession}
          />
        );
      })()}
    </div>
  );
}
