import { useEffect, useState } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell,
} from 'recharts';
import type { SessionMeta } from '../../hooks/useSessionHistory';

interface ProjectViewProps {
  sessions: SessionMeta[];
  onSelectSession: (id: string) => void;
}

interface TokenData {
  input_tokens: number;
  output_tokens: number;
  estimated_cost_usd: number;
}

interface BarEntry {
  id: string;
  prompt: string;
  cost: number;
  status: string;
}

const cardStyle: React.CSSProperties = {
  background: 'var(--bg-surface)',
  border: '1px solid var(--border)',
  borderRadius: '8px',
  padding: '16px',
};

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div style={{ ...cardStyle, flex: 1, minWidth: '140px' }}>
      <div style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)', marginBottom: '6px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
        {label}
      </div>
      <div style={{ fontFamily: 'var(--font-mono)', fontSize: '22px', color: 'var(--text-primary)', fontWeight: 700 }}>
        {value}
      </div>
    </div>
  );
}

function formatDuration(startedAt: string, finishedAt?: string): string {
  if (!finishedAt) return 'running...';
  const ms = new Date(finishedAt).getTime() - new Date(startedAt).getTime();
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  if (m > 0) return `${m}m ${s % 60}s`;
  return `${s}s`;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString();
}

const STATUS_COLORS: Record<string, string> = {
  done: '#3fb950',
  error: '#f85149',
  running: '#d29922',
};

export function ProjectView({ sessions, onSelectSession }: ProjectViewProps) {
  const [projectTokens, setProjectTokens] = useState<TokenData | null>(null);
  const [sessionCosts, setSessionCosts] = useState<Record<string, number>>({});

  useEffect(() => {
    fetch('/tokens')
      .then((r) => r.json())
      .then((d: TokenData) => setProjectTokens(d))
      .catch(() => {});
  }, []);

  useEffect(() => {
    const fetchCosts = async () => {
      const results: Record<string, number> = {};
      await Promise.all(
        sessions.map(async (s) => {
          try {
            const r = await fetch(`/sessions/${s.id}/tokens`);
            const d: TokenData = await r.json();
            results[s.id] = d.estimated_cost_usd ?? 0;
          } catch {
            results[s.id] = 0;
          }
        })
      );
      setSessionCosts(results);
    };
    if (sessions.length > 0) fetchCosts();
  }, [sessions]);

  const doneCount = sessions.filter((s) => s.status === 'done').length;
  const errorCount = sessions.filter((s) => s.status === 'error').length;
  const runningCount = sessions.filter((s) => s.status === 'running').length;

  const barData: BarEntry[] = sessions
    .slice()
    .sort((a, b) => a.startedAt.localeCompare(b.startedAt))
    .map((s) => ({
      id: s.id,
      prompt: s.prompt.slice(0, 40) + (s.prompt.length > 40 ? '…' : ''),
      cost: Number(sessionCosts[s.id] ?? 0),
      status: s.status,
    }));

  const hasAnyCost = barData.some((d) => d.cost > 0);
  const barDomain: [number, number | string] = hasAnyCost ? [0, 'auto'] : [0, 0.001];

  const pieData = [
    { name: 'Done', value: doneCount, color: '#3fb950' },
    { name: 'Error', value: errorCount, color: '#f85149' },
    { name: 'Running', value: runningCount, color: '#d29922' },
  ].filter((d) => d.value > 0);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      {/* Stat cards */}
      <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap' }}>
        <StatCard label="Total Missions" value={sessions.length} />
        <StatCard label="Done" value={doneCount} />
        <StatCard label="Errors" value={errorCount} />
        <StatCard label="Input Tokens" value={projectTokens ? projectTokens.input_tokens.toLocaleString() : '—'} />
        <StatCard label="Output Tokens" value={projectTokens ? projectTokens.output_tokens.toLocaleString() : '—'} />
        <StatCard label="Total Cost" value={projectTokens?.estimated_cost_usd != null ? `$${projectTokens.estimated_cost_usd.toFixed(4)}` : '—'} />
      </div>

      {/* Charts row */}
      <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
        {/* Cost per mission bar chart */}
        <div style={{ ...cardStyle, flex: 2, minWidth: '320px' }}>
          <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
            Cost per Mission (USD)
          </div>
          {barData.length === 0 ? (
            <div style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-ui)', fontSize: '13px' }}>No sessions yet</div>
          ) : (
            <ResponsiveContainer width="100%" height={Math.max(120, barData.length * 36)}>
              <BarChart data={barData} layout="vertical" margin={{ left: 8, right: 24, top: 4, bottom: 4 }}>
                <XAxis type="number" domain={barDomain} allowDataOverflow tick={{ fill: 'var(--text-muted)', fontSize: 11, fontFamily: 'var(--font-mono)' }} tickFormatter={(v: unknown) => `$${Number(v ?? 0).toFixed(4)}`} />
                <YAxis type="category" dataKey="prompt" width={180} tick={{ fill: 'var(--text-muted)', fontSize: 11, fontFamily: 'var(--font-ui)' }} />
                <Tooltip
                    contentStyle={{ background: 'var(--bg-surface-2)', border: '1px solid var(--border)', borderRadius: '6px', fontFamily: 'var(--font-mono)', fontSize: '12px' }}
                    formatter={(value) => [`$${Number(value ?? 0).toFixed(6)}`, 'Cost']}
                  />
                <Bar
                  dataKey="cost"
                  radius={[0, 4, 4, 0]}
                  onClick={(data: unknown) => onSelectSession((data as BarEntry).id)}
                  style={{ cursor: 'pointer' }}
                >
                  {barData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={entry.status === 'error' ? '#f85149' : '#58a6ff'}
                    />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          )}
        </div>

        {/* Status donut */}
        <div style={{ ...cardStyle, flex: 1, minWidth: '200px', display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
          <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px', alignSelf: 'flex-start' }}>
            Status Distribution
          </div>
          {pieData.length === 0 ? (
            <div style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-ui)', fontSize: '13px' }}>No sessions yet</div>
          ) : (
            <>
              <ResponsiveContainer width="100%" height={160}>
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    innerRadius={45}
                    outerRadius={70}
                    dataKey="value"
                  >
                    {pieData.map((entry, index) => (
                      <Cell key={`pie-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ background: 'var(--bg-surface-2)', border: '1px solid var(--border)', borderRadius: '6px', fontFamily: 'var(--font-mono)', fontSize: '12px' }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap', justifyContent: 'center' }}>
                {pieData.map((d) => (
                  <div key={d.name} style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <div style={{ width: '8px', height: '8px', borderRadius: '50%', background: d.color }} />
                    <span style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)' }}>{d.name} ({d.value})</span>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>

      {/* Sessions table */}
      <div style={cardStyle}>
        <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
          All Sessions
        </div>
        {sessions.length === 0 ? (
          <div style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-ui)', fontSize: '13px' }}>No sessions yet</div>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontFamily: 'var(--font-ui)', fontSize: '12px' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border)' }}>
                  {['Prompt', 'Status', 'Started At', 'Duration', 'Cost'].map((h) => (
                    <th key={h} style={{ textAlign: 'left', padding: '8px 12px', color: 'var(--text-muted)', fontWeight: 600, fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {[...sessions].sort((a, b) => b.startedAt.localeCompare(a.startedAt)).map((s) => (
                  <tr
                    key={s.id}
                    onClick={() => onSelectSession(s.id)}
                    style={{ borderBottom: '1px solid var(--border)', cursor: 'pointer', transition: 'background 150ms' }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--bg-surface-2)')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                  >
                    <td style={{ padding: '10px 12px', color: 'var(--text-primary)', maxWidth: '300px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {s.prompt.slice(0, 80)}{s.prompt.length > 80 ? '…' : ''}
                    </td>
                    <td style={{ padding: '10px 12px' }}>
                      <span style={{ color: STATUS_COLORS[s.status] ?? 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 600 }}>
                        {s.status.toUpperCase()}
                      </span>
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px', whiteSpace: 'nowrap' }}>
                      {formatDate(s.startedAt)}
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>
                      {formatDuration(s.startedAt, s.finishedAt)}
                    </td>
                    <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>
                      {sessionCosts[s.id] !== undefined ? `$${sessionCosts[s.id].toFixed(6)}` : '—'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
