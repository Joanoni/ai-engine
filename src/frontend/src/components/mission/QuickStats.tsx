import { useMemo, useState, useEffect } from 'react';
import type { EngineEvent } from '../../types/events';

interface QuickStatsProps {
  events: EngineEvent[];
  startTime: Date | null;
}

interface TokenStats {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  estimated_cost_usd: number;
}

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  const h = Math.floor(m / 60);
  if (h > 0) return `${h}h ${m % 60}m`;
  if (m > 0) return `${m}m ${s % 60}s`;
  return `${s}s`;
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}k`;
  return String(n);
}

function formatCost(usd: number): string {
  if (usd === 0) return '';
  if (usd < 0.001) return `≈ $${usd.toFixed(5)}`;
  if (usd < 0.01) return `≈ $${usd.toFixed(4)}`;
  if (usd < 1) return `≈ $${usd.toFixed(3)}`;
  return `≈ $${usd.toFixed(2)}`;
}

export function QuickStats({ events, startTime }: QuickStatsProps) {
  const [now, setNow] = useState(Date.now());
  const [sessionTokens, setSessionTokens] = useState<TokenStats | null>(null);
  const [projectTokens, setProjectTokens] = useState<TokenStats | null>(null);

  const isRunning = useMemo(() => {
    const last = [...events].reverse().find(
      (e) => e.type === 'session.started' || e.type === 'session.finished'
    );
    return last?.type === 'session.started';
  }, [events]);

  useEffect(() => {
    if (!isRunning) return;
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, [isRunning]);

  // Fetch project tokens on mount.
  useEffect(() => {
    fetch('/tokens')
      .then((r) => r.json())
      .then((data: TokenStats) => setProjectTokens(data))
      .catch(() => {/* non-fatal */});
  }, []);

  // Fetch session tokens and refresh project tokens when session finishes.
  useEffect(() => {
    const finished = [...events].reverse().find((e) => e.type === 'session.finished');
    if (!finished) return;

    const sessionId = finished.session_id;
    if (!sessionId) return;

    fetch(`/sessions/${sessionId}/tokens`)
      .then((r) => r.json())
      .then((data: TokenStats) => setSessionTokens(data))
      .catch(() => {/* non-fatal */});

    fetch('/tokens')
      .then((r) => r.json())
      .then((data: TokenStats) => setProjectTokens(data))
      .catch(() => {/* non-fatal */});
  }, [events]);

  const stats = useMemo(() => {
    const toolCalls = events.filter((e) => e.type === 'tool.called').length;
    const agentNames = new Set(
      events.filter((e) => e.agent_name).map((e) => e.agent_name!)
    );
    return { toolCalls, agentCount: agentNames.size };
  }, [events]);

  const duration = startTime
    ? formatDuration(now - startTime.getTime())
    : '—';

  const statCards = [
    { label: 'Tool Calls', value: String(stats.toolCalls), color: 'var(--warning)', sub: null },
    { label: 'Agents', value: String(stats.agentCount), color: 'var(--purple)', sub: null },
    { label: 'Duration', value: duration, color: 'var(--accent)', sub: null },
    {
      label: 'Session Tokens',
      value: sessionTokens ? formatTokens(sessionTokens.total_tokens) : '—',
      color: 'var(--success, #4ade80)',
      sub: sessionTokens && sessionTokens.estimated_cost_usd > 0
        ? formatCost(sessionTokens.estimated_cost_usd)
        : null,
    },
    {
      label: 'Project Tokens',
      value: projectTokens ? formatTokens(projectTokens.total_tokens) : '—',
      color: 'var(--info, #60a5fa)',
      sub: projectTokens && projectTokens.estimated_cost_usd > 0
        ? formatCost(projectTokens.estimated_cost_usd)
        : null,
    },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
      <div
        style={{
          fontFamily: 'var(--font-ui)',
          fontSize: '11px',
          fontWeight: 600,
          color: 'var(--text-muted)',
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
        }}
      >
        Quick Stats
      </div>

      <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
        {statCards.map((card) => (
          <div
            key={card.label}
            style={{
              flex: '1 1 calc(33% - 8px)',
              minWidth: '72px',
              background: 'var(--bg-surface-2)',
              border: '1px solid var(--border)',
              borderRadius: '8px',
              padding: '10px 8px',
              textAlign: 'center',
            }}
          >
            <div
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '18px',
                fontWeight: 600,
                color: card.color,
                lineHeight: 1.2,
              }}
            >
              {card.value}
            </div>
            {card.sub && (
              <div
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '10px',
                  color: card.color,
                  opacity: 0.75,
                  marginTop: '2px',
                  lineHeight: 1.2,
                }}
              >
                {card.sub}
              </div>
            )}
            <div
              style={{
                fontFamily: 'var(--font-ui)',
                fontSize: '10px',
                color: 'var(--text-muted)',
                marginTop: '4px',
              }}
            >
              {card.label}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
