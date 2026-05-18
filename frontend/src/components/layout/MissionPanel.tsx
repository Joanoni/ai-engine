import type { EngineEvent } from '../../types/events';
import { PromptEditor } from '../mission/PromptEditor';
import { LaunchButton } from '../mission/LaunchButton';
import { TaskProgress } from '../mission/TaskProgress';
import { AgentRoster } from '../mission/AgentRoster';
import { QuickStats } from '../mission/QuickStats';

interface MissionPanelProps {
  events: EngineEvent[];
  isConnected: boolean;
  isRunning: boolean;
  connectionStatus: string;
  latency: number | null;
  sessionStartTime: Date | null;
  onSend: (text: string) => void;
}

export function MissionPanel({
  events,
  isConnected,
  isRunning,
  connectionStatus,
  latency,
  sessionStartTime,
  onSend,
}: MissionPanelProps) {
  const hasAgents = events.some((e) => e.agent_name);
  const hasTaskUpdates = events.some((e) => e.type === 'tasks.updated');
  const sessionStarted = events.some((e) => e.type === 'session.started');

  return (
    <div
      style={{
        width: '360px',
        height: '100%',
        background: 'var(--bg-surface)',
        borderLeft: '1px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        flexShrink: 0,
        overflow: 'hidden',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '14px 16px',
          borderBottom: '1px solid var(--border)',
          flexShrink: 0,
        }}
      >
        <div
          style={{
            fontFamily: 'var(--font-ui)',
            fontWeight: 700,
            fontSize: '13px',
            color: 'var(--text-primary)',
            letterSpacing: '0.05em',
            textTransform: 'uppercase',
            marginBottom: '10px',
          }}
        >
          Mission Control
        </div>

        {/* Connection status */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              background: isConnected ? 'var(--success)' : 'var(--error)',
              boxShadow: isConnected ? '0 0 6px var(--success)' : 'none',
              flexShrink: 0,
            }}
          />
          <span
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              color: isConnected ? 'var(--success)' : 'var(--error)',
            }}
          >
            {connectionStatus}
          </span>
          {latency !== null && isConnected && (
            <span
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '10px',
                color: 'var(--text-muted)',
                marginLeft: 'auto',
              }}
            >
              {latency}ms
            </span>
          )}
        </div>
      </div>

      {/* Scrollable content */}
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '16px',
          display: 'flex',
          flexDirection: 'column',
          gap: '20px',
        }}
      >
        {/* Prompt Editor */}
        <PromptEditor onSend={onSend} disabled={isRunning || !isConnected} />

        {/* Launch Button */}
        <LaunchButton
          onClick={() => {
            // The actual send is handled by PromptEditor's Ctrl+Enter
            // This button triggers the same action via the textarea value
            const textarea = document.querySelector<HTMLTextAreaElement>('textarea');
            const value = textarea?.value.trim();
            if (value) {
              onSend(value);
              if (textarea) {
                textarea.value = '';
                textarea.style.height = 'auto';
              }
            }
          }}
          disabled={!isConnected}
          isRunning={isRunning}
        />

        {/* Divider */}
        <div style={{ height: '1px', background: 'var(--border)' }} />

        {/* Task Progress — only when tasks exist */}
        {hasTaskUpdates && <TaskProgress events={events} />}

        {/* Agent Roster — only when agents exist */}
        {hasAgents && <AgentRoster events={events} />}

        {/* Quick Stats — only when session started */}
        {sessionStarted && (
          <QuickStats events={events} startTime={sessionStartTime} />
        )}
      </div>
    </div>
  );
}
