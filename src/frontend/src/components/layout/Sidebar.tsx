import { useMemo, useState, useEffect } from 'react';
import type { SessionMeta } from '../../hooks/useSessionHistory';

interface SidebarProps {
  collapsed: boolean;
  sessions: SessionMeta[];
  activeSessionId: string | null;
  onNewMission: () => void;
  onLoadSession: (id: string) => void;
  onToggleAnalytics: () => void;
  showAnalytics: boolean;
}

function timeAgo(isoString: string): string {
  const diff = Date.now() - new Date(isoString).getTime();
  const s = Math.floor(diff / 1000);
  const m = Math.floor(s / 60);
  const h = Math.floor(m / 60);
  const d = Math.floor(h / 24);
  if (d > 0) return `${d}d ago`;
  if (h > 0) return `${h}h ago`;
  if (m > 0) return `${m} min ago`;
  return 'just now';
}

const statusColors: Record<string, string> = {
  running: 'var(--warning)',
  done: 'var(--success)',
  error: 'var(--error)',
};

export function Sidebar({
  collapsed,
  sessions,
  activeSessionId,
  onNewMission,
  onLoadSession,
  onToggleAnalytics,
  showAnalytics,
}: SidebarProps) {
  const sessionCount = sessions.length;
  const [version, setVersion] = useState<string>('...');

  useEffect(() => {
    fetch('/version')
      .then((r) => r.json())
      .then((data: { version: string }) => setVersion(data.version))
      .catch(() => setVersion('?'));
  }, []);

  const sortedSessions = useMemo(
    () => [...sessions].sort((a, b) => b.startedAt.localeCompare(a.startedAt)),
    [sessions]
  );

  return (
    <div
      style={{
        width: collapsed ? '56px' : '240px',
        height: '100%',
        background: 'var(--bg-surface)',
        borderRight: '1px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        transition: 'width 200ms ease',
        overflow: 'hidden',
        flexShrink: 0,
      }}
    >
      {/* Logo area */}
      <div
        style={{
          padding: collapsed ? '16px 0' : '16px',
          borderBottom: '1px solid var(--border)',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          justifyContent: collapsed ? 'center' : 'flex-start',
          flexShrink: 0,
        }}
      >
        {/* Icon */}
        <div
          style={{
            width: '28px',
            height: '28px',
            borderRadius: '6px',
            background: 'linear-gradient(135deg, #1d4ed8, #58a6ff)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            position: 'relative',
          }}
        >
          <span style={{ fontSize: '14px' }}>🤖</span>
          {collapsed && sessionCount > 0 && (
            <div
              style={{
                position: 'absolute',
                top: '-4px',
                right: '-4px',
                width: '14px',
                height: '14px',
                borderRadius: '50%',
                background: 'var(--accent)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: '8px',
                fontWeight: 700,
                color: '#000',
                fontFamily: 'var(--font-mono)',
              }}
            >
              {sessionCount > 9 ? '9+' : sessionCount}
            </div>
          )}
        </div>

        {!collapsed && (
          <>
            <div>
              <div
                style={{
                  fontFamily: 'var(--font-ui)',
                  fontWeight: 700,
                  fontSize: '14px',
                  color: 'var(--text-primary)',
                  lineHeight: 1.2,
                }}
              >
                AI Engine
              </div>
            </div>
            <span
              style={{
                marginLeft: 'auto',
                fontFamily: 'var(--font-mono)',
                fontSize: '10px',
                color: 'var(--accent)',
                background: 'rgba(88,166,255,0.12)',
                padding: '2px 6px',
                borderRadius: '4px',
                flexShrink: 0,
              }}
            >
              {version}
            </span>
          </>
        )}
      </div>

      {/* Session history */}
      {!collapsed && (
        <div
          style={{
            flex: 1,
            overflowY: 'auto',
            padding: '8px',
          }}
        >
          {sortedSessions.length === 0 ? (
            <div
              style={{
                padding: '16px 8px',
                color: 'var(--text-muted)',
                fontFamily: 'var(--font-ui)',
                fontSize: '12px',
                textAlign: 'center',
              }}
            >
              No sessions yet
            </div>
          ) : (
            sortedSessions.map((session) => (
              <button
                key={session.id}
                onClick={() => onLoadSession(session.id)}
                style={{
                  width: '100%',
                  background: 'transparent',
                  border: 'none',
                  borderLeft:
                    session.id === activeSessionId
                      ? '3px solid var(--accent)'
                      : '3px solid transparent',
                  borderRadius: '0 6px 6px 0',
                  padding: '8px 10px',
                  cursor: 'pointer',
                  textAlign: 'left',
                  marginBottom: '2px',
                  transition: 'all 200ms ease',
                  background2: session.id === activeSessionId ? 'rgba(88,166,255,0.08)' : 'transparent',
                } as React.CSSProperties}
                onMouseEnter={(e) => {
                  if (session.id !== activeSessionId) {
                    (e.currentTarget as HTMLButtonElement).style.background = 'var(--bg-surface-2)';
                  }
                }}
                onMouseLeave={(e) => {
                  if (session.id !== activeSessionId) {
                    (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
                  }
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    marginBottom: '3px',
                  }}
                >
                  <div
                    style={{
                      width: '6px',
                      height: '6px',
                      borderRadius: '50%',
                      background: statusColors[session.status] ?? 'var(--text-muted)',
                      flexShrink: 0,
                    }}
                  />
                  <span
                    style={{
                      fontFamily: 'var(--font-mono)',
                      fontSize: '10px',
                      color: 'var(--text-muted)',
                    }}
                  >
                    {timeAgo(session.startedAt)}
                  </span>
                </div>
                <div
                  style={{
                    fontFamily: 'var(--font-ui)',
                    fontSize: '12px',
                    color: 'var(--text-primary)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {session.prompt.slice(0, 60)}
                  {session.prompt.length > 60 ? '…' : ''}
                </div>
              </button>
            ))
          )}
        </div>
      )}

      {/* Bottom actions */}
      <div
        style={{
          padding: collapsed ? '12px 8px' : '12px',
          borderTop: '1px solid var(--border)',
          flexShrink: 0,
        }}
      >
        <button
          onClick={onToggleAnalytics}
          style={{
            width: '100%',
            height: '36px',
            borderRadius: '6px',
            background: showAnalytics ? 'rgba(188,140,255,0.15)' : 'transparent',
            border: showAnalytics ? '1px solid rgba(188,140,255,0.4)' : '1px solid var(--border)',
            color: showAnalytics ? 'var(--purple)' : 'var(--text-muted)',
            fontFamily: 'var(--font-ui)',
            fontWeight: 600,
            fontSize: collapsed ? '14px' : '13px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '6px',
            marginBottom: '8px',
          }}
        >
          {collapsed ? '📊' : '📊 Analytics'}
        </button>
        <button
          onClick={onNewMission}
          style={{
            width: '100%',
            height: '36px',
            borderRadius: '6px',
            background: 'rgba(88,166,255,0.12)',
            border: '1px solid rgba(88,166,255,0.3)',
            color: 'var(--accent)',
            fontFamily: 'var(--font-ui)',
            fontWeight: 600,
            fontSize: collapsed ? '16px' : '13px',
            cursor: 'pointer',
            transition: 'all 200ms ease',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '6px',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(88,166,255,0.2)';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(88,166,255,0.12)';
          }}
        >
          {collapsed ? '+' : '+ New Mission'}
        </button>
      </div>
    </div>
  );
}
