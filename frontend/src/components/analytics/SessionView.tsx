import { useEffect, useState } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell,
} from 'recharts';
import type { SessionMeta } from '../../hooks/useSessionHistory';
import type { SessionLogs, LogEntry } from '../../types/logs';

interface SessionViewProps {
  sessionId: string;
  sessions: SessionMeta[];
  logs: SessionLogs | null;
  loadingLogs: boolean;
  onSelectAgent: (agentName: string) => void;
  onBack: () => void;
}

interface TokenData {
  input_tokens: number;
  output_tokens: number;
  estimated_cost_usd: number;
}

const cardStyle: React.CSSProperties = {
  background: 'var(--bg-surface)',
  border: '1px solid var(--border)',
  borderRadius: '8px',
  padding: '16px',
};

const STATUS_COLORS: Record<string, string> = {
  done: '#3fb950',
  error: '#f85149',
  running: '#d29922',
};

function formatDuration(startedAt: string, finishedAt?: string): string {
  if (!finishedAt) return 'running...';
  const ms = new Date(finishedAt).getTime() - new Date(startedAt).getTime();
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  if (m > 0) return `${m}m ${s % 60}s`;
  return `${s}s`;
}

interface AgentStats {
  name: string;
  agentType: string;
  model: string;
  turns: number;
  toolCalls: number;
  inputTokens: number;
  outputTokens: number;
  cost: number;
  avgToolDuration: number;
  startTs: string;
  endTs: string;
}

function computeAgentStats(agentName: string, entries: LogEntry[]): AgentStats {
  const initEntry = entries.find((e) => e.role === 'agent_init');
  const finishEntry = entries.find((e) => e.role === 'finish');
  const llmRequests = entries.filter((e) => e.role === 'llm_request');
  const llmResponses = entries.filter((e) => e.role === 'llm_response');
  const toolResults = entries.filter((e) => e.role === 'tool_result');

  const inputTokens = llmResponses.reduce((sum, e) => sum + (e.input_tokens ?? 0), 0);
  const outputTokens = llmResponses.reduce((sum, e) => sum + (e.output_tokens ?? 0), 0);
  const cost = inputTokens * 3 / 1_000_000 + outputTokens * 15 / 1_000_000;

  const toolDurations = toolResults.map((e) => e.duration_ms ?? 0).filter((d) => d > 0);
  const avgToolDuration = toolDurations.length > 0
    ? toolDurations.reduce((a, b) => a + b, 0) / toolDurations.length
    : 0;

  return {
    name: agentName,
    agentType: initEntry?.agent_type ?? '?',
    model: initEntry?.model ?? '?',
    turns: llmRequests.length,
    toolCalls: toolResults.length,
    inputTokens,
    outputTokens,
    cost,
    avgToolDuration,
    startTs: initEntry?.ts ?? entries[0]?.ts ?? '',
    endTs: finishEntry?.ts ?? entries[entries.length - 1]?.ts ?? '',
  };
}

interface ToolUsage {
  name: string;
  total: number;
  success: number;
  error: number;
}

function computeToolUsage(logs: SessionLogs): ToolUsage[] {
  const map: Record<string, ToolUsage> = {};
  for (const entries of Object.values(logs.agents)) {
    for (const e of entries) {
      if (e.role === 'tool_result' && e.tool) {
        if (!map[e.tool]) map[e.tool] = { name: e.tool, total: 0, success: 0, error: 0 };
        map[e.tool].total++;
        if (e.success) map[e.tool].success++;
        else map[e.tool].error++;
      }
    }
  }
  return Object.values(map).sort((a, b) => b.total - a.total);
}

function toolBarColor(usage: ToolUsage): string {
  if (usage.total === 0) return '#7d8590';
  const rate = usage.success / usage.total;
  if (rate >= 0.9) return '#3fb950';
  if (rate >= 0.5) return '#d29922';
  return '#f85149';
}

export function SessionView({ sessionId, sessions, logs, loadingLogs, onSelectAgent, onBack }: SessionViewProps) {
  const [sessionCost, setSessionCost] = useState<TokenData | null>(null);

  const session = sessions.find((s) => s.id === sessionId);

  useEffect(() => {
    fetch(`/sessions/${sessionId}/tokens`)
      .then((r) => r.json())
      .then((d: TokenData) => setSessionCost(d))
      .catch(() => {});
  }, [sessionId]);

  if (loadingLogs) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '200px', color: 'var(--text-muted)', fontFamily: 'var(--font-ui)', fontSize: '14px' }}>
        Loading logs...
      </div>
    );
  }

  const agentStatsList: AgentStats[] = logs
    ? Object.entries(logs.agents).map(([name, entries]) => computeAgentStats(name, entries))
    : [];

  const toolUsage = logs ? computeToolUsage(logs) : [];

  // Swimlane timeline
  const allTimestamps = agentStatsList.flatMap((a) => [a.startTs, a.endTs]).filter(Boolean);
  const minTs = allTimestamps.length > 0 ? Math.min(...allTimestamps.map((t) => new Date(t).getTime())) : 0;
  const maxTs = allTimestamps.length > 0 ? Math.max(...allTimestamps.map((t) => new Date(t).getTime())) : 1;
  const totalSpan = maxTs - minTs || 1;

  const agentColors = ['#58a6ff', '#bc8cff', '#3fb950', '#d29922', '#ff7b72', '#79c0ff'];

  // Cost waterfall
  const totalCost = agentStatsList.reduce((sum, a) => sum + a.cost, 0);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      {/* Header */}
      <div style={cardStyle}>
        <div style={{ display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
          <button
            onClick={onBack}
            style={{ background: 'transparent', border: '1px solid var(--border)', borderRadius: '6px', color: 'var(--text-muted)', cursor: 'pointer', padding: '6px 10px', fontFamily: 'var(--font-ui)', fontSize: '13px', flexShrink: 0 }}
          >
            ← Back
          </button>
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'var(--font-ui)', fontSize: '14px', color: 'var(--text-primary)', marginBottom: '8px', lineHeight: 1.5 }}>
              {session?.prompt ?? sessionId}
            </div>
            <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap', alignItems: 'center' }}>
              {session && (
                <span style={{ color: STATUS_COLORS[session.status] ?? 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 600 }}>
                  {session.status.toUpperCase()}
                </span>
              )}
              {session && (
                <span style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>
                  {formatDuration(session.startedAt, session.finishedAt)}
                </span>
              )}
              {sessionCost?.estimated_cost_usd != null && (
                <span style={{ color: 'var(--accent)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>
                  ${sessionCost.estimated_cost_usd.toFixed(6)} total
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {!logs ? (
        <div style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-ui)', fontSize: '13px', padding: '16px' }}>
          Failed to load logs for this session.
        </div>
      ) : (
        <>
          {/* Agent swimlane timeline */}
          {agentStatsList.length > 0 && (
            <div style={cardStyle}>
              <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
                Agent Timeline
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                {agentStatsList.map((agent, idx) => {
                  const startPct = ((new Date(agent.startTs).getTime() - minTs) / totalSpan) * 100;
                  const endPct = ((new Date(agent.endTs).getTime() - minTs) / totalSpan) * 100;
                  const widthPct = Math.max(endPct - startPct, 1);
                  const color = agentColors[idx % agentColors.length];
                  return (
                    <div key={agent.name} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <div style={{ width: '120px', fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)', textAlign: 'right', flexShrink: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {agent.name}
                      </div>
                      <div style={{ flex: 1, height: '20px', background: 'var(--bg-surface-2)', borderRadius: '4px', position: 'relative' }}>
                        <div
                          style={{
                            position: 'absolute',
                            left: `${startPct}%`,
                            width: `${widthPct}%`,
                            height: '100%',
                            background: color,
                            borderRadius: '4px',
                            opacity: 0.8,
                            display: 'flex',
                            alignItems: 'center',
                            paddingLeft: '6px',
                            overflow: 'hidden',
                          }}
                        >
                          <span style={{ fontFamily: 'var(--font-ui)', fontSize: '10px', color: '#000', fontWeight: 600, whiteSpace: 'nowrap' }}>
                            {agent.name}
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Agent summary table */}
          <div style={cardStyle}>
            <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
              Agent Summary
            </div>
            <div style={{ overflowX: 'auto' }}>
              <table style={{ width: '100%', borderCollapse: 'collapse', fontFamily: 'var(--font-ui)', fontSize: '12px' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--border)' }}>
                    {['Agent', 'Type', 'LLM Turns', 'Tool Calls', 'Input Tok', 'Output Tok', 'Cost', 'Avg Tool ms'].map((h) => (
                      <th key={h} style={{ textAlign: 'left', padding: '8px 12px', color: 'var(--text-muted)', fontWeight: 600, fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.05em', whiteSpace: 'nowrap' }}>{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {agentStatsList.map((agent) => (
                    <tr
                      key={agent.name}
                      onClick={() => onSelectAgent(agent.name)}
                      style={{ borderBottom: '1px solid var(--border)', cursor: 'pointer', transition: 'background 150ms' }}
                      onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--bg-surface-2)')}
                      onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                    >
                      <td style={{ padding: '10px 12px', color: 'var(--text-primary)', fontWeight: 600 }}>{agent.name}</td>
                      <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>
                        <span style={{ background: agent.agentType === 'leader' ? 'rgba(188,140,255,0.15)' : 'rgba(88,166,255,0.15)', color: agent.agentType === 'leader' ? 'var(--purple)' : 'var(--accent)', padding: '2px 6px', borderRadius: '4px', fontWeight: 600 }}>
                          {agent.agentType === 'leader' ? 'L' : 'E'}
                        </span>
                      </td>
                      <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>{agent.turns}</td>
                      <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>{agent.toolCalls}</td>
                      <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>{agent.inputTokens.toLocaleString()}</td>
                      <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>{agent.outputTokens.toLocaleString()}</td>
                      <td style={{ padding: '10px 12px', color: 'var(--accent)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>${agent.cost.toFixed(6)}</td>
                      <td style={{ padding: '10px 12px', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)', fontSize: '11px' }}>{agent.avgToolDuration.toFixed(0)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Tool usage bar chart */}
          {toolUsage.length > 0 && (
            <div style={cardStyle}>
              <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
                Tool Usage
              </div>
              <ResponsiveContainer width="100%" height={Math.max(120, toolUsage.length * 36)}>
                <BarChart data={toolUsage} layout="vertical" margin={{ left: 8, right: 24, top: 4, bottom: 4 }}>
                  <XAxis type="number" tick={{ fill: 'var(--text-muted)', fontSize: 11, fontFamily: 'var(--font-mono)' }} />
                  <YAxis type="category" dataKey="name" width={160} tick={{ fill: 'var(--text-muted)', fontSize: 11, fontFamily: 'var(--font-mono)' }} />
                  <Tooltip
                    contentStyle={{ background: 'var(--bg-surface-2)', border: '1px solid var(--border)', borderRadius: '6px', fontFamily: 'var(--font-mono)', fontSize: '12px' }}
                  />
                  <Bar dataKey="total" radius={[0, 4, 4, 0]}>
                    {toolUsage.map((entry, index) => (
                      <Cell key={`tool-${index}`} fill={toolBarColor(entry)} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}

          {/* Cost waterfall */}
          {agentStatsList.length > 0 && totalCost > 0 && (
            <div style={cardStyle}>
              <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px' }}>
                Cost Breakdown by Agent
              </div>
              <div style={{ display: 'flex', height: '32px', borderRadius: '6px', overflow: 'hidden', gap: '1px' }}>
                {agentStatsList.map((agent, idx) => {
                  const pct = totalCost > 0 ? (agent.cost / totalCost) * 100 : 0;
                  const color = agentColors[idx % agentColors.length];
                  return (
                    <div
                      key={agent.name}
                      title={`${agent.name}: $${agent.cost.toFixed(6)} (${pct.toFixed(1)}%)`}
                      style={{ width: `${pct}%`, background: color, display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden', minWidth: pct > 5 ? undefined : '0' }}
                    >
                      {pct > 8 && (
                        <span style={{ fontFamily: 'var(--font-ui)', fontSize: '10px', color: '#000', fontWeight: 600, whiteSpace: 'nowrap', padding: '0 4px' }}>
                          {agent.name}
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
              <div style={{ display: 'flex', gap: '12px', flexWrap: 'wrap', marginTop: '8px' }}>
                {agentStatsList.map((agent, idx) => (
                  <div key={agent.name} style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                    <div style={{ width: '8px', height: '8px', borderRadius: '2px', background: agentColors[idx % agentColors.length] }} />
                    <span style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)' }}>
                      {agent.name} (${agent.cost.toFixed(6)})
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
