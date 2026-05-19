import { useState } from 'react';
import {
  LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer,
  BarChart, Bar, Cell,
} from 'recharts';
import type { LogEntry, ToolCallEntry } from '../../types/logs';

interface AgentViewProps {
  agentName: string;
  sessionId: string;
  logs: LogEntry[];
  onBack: () => void;
}

const cardStyle: React.CSSProperties = {
  background: 'var(--bg-surface)',
  border: '1px solid var(--border)',
  borderRadius: '8px',
  padding: '16px',
};

const preStyle: React.CSSProperties = {
  fontFamily: 'var(--font-mono)',
  fontSize: '11px',
  color: 'var(--text-primary)',
  background: 'var(--bg-base)',
  border: '1px solid var(--border)',
  borderRadius: '6px',
  padding: '10px',
  overflowY: 'auto',
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
  margin: 0,
};

interface TurnGroup {
  turn: number;
  llmRequest: LogEntry | null;
  llmResponse: LogEntry | null;
  toolResults: LogEntry[];
}

function groupByTurn(logs: LogEntry[]): TurnGroup[] {
  const map = new Map<number, TurnGroup>();
  for (const entry of logs) {
    if (entry.role === 'agent_init' || entry.role === 'user' || entry.role === 'finish' || entry.role === 'error') continue;
    const t = entry.turn;
    if (!map.has(t)) map.set(t, { turn: t, llmRequest: null, llmResponse: null, toolResults: [] });
    const group = map.get(t)!;
    if (entry.role === 'llm_request') group.llmRequest = entry;
    else if (entry.role === 'llm_response') group.llmResponse = entry;
    else if (entry.role === 'tool_result') group.toolResults.push(entry);
  }
  return Array.from(map.values()).sort((a, b) => a.turn - b.turn);
}

function SystemPromptTabs({ layers }: { layers: NonNullable<LogEntry['system_layers']> }) {
  const [activeTab, setActiveTab] = useState(0);
  const tabs = [
    { label: 'Engine Context', content: layers.engine_context },
    { label: 'Workspace Tree (L4)', content: layers.dynamic_context },
    { label: 'Agent Role', content: layers.agent_role },
    { label: 'Task Context', content: layers.task_context },
  ];

  return (
    <div>
      <div style={{ display: 'flex', gap: '4px', marginBottom: '8px', flexWrap: 'wrap' }}>
        {tabs.map((tab, i) => (
          <button
            key={tab.label}
            onClick={() => setActiveTab(i)}
            style={{
              background: activeTab === i ? 'rgba(88,166,255,0.15)' : 'transparent',
              border: activeTab === i ? '1px solid rgba(88,166,255,0.4)' : '1px solid var(--border)',
              borderRadius: '4px',
              color: activeTab === i ? 'var(--accent)' : 'var(--text-muted)',
              fontFamily: 'var(--font-ui)',
              fontSize: '11px',
              padding: '4px 8px',
              cursor: 'pointer',
            }}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <pre style={{ ...preStyle, maxHeight: '300px' }}>
        {tabs[activeTab].content
          ? tabs[activeTab].content
          : <span style={{ color: 'var(--text-muted)' }}>(empty)</span>}
      </pre>
    </div>
  );
}

function MessageHistorySection({ messages }: { messages: NonNullable<LogEntry['messages']> }) {
  const [expanded, setExpanded] = useState(false);
  const [expandedIdx, setExpandedIdx] = useState<Set<number>>(new Set());

  return (
    <div>
      <button
        onClick={() => setExpanded((p) => !p)}
        style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontFamily: 'var(--font-ui)', fontSize: '12px', padding: 0, marginBottom: '6px' }}
      >
        {expanded ? '▼' : '▶'} {messages.length} messages in history
      </button>
      {expanded && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
          {messages.map((msg, i) => {
            const isExpanded = expandedIdx.has(i);
            const preview = msg.content.map((c) => c.text ?? c.content ?? c.name ?? c.type).join(' ').slice(0, 200);
            return (
              <div key={i} style={{ background: 'var(--bg-base)', border: '1px solid var(--border)', borderRadius: '4px', padding: '6px 10px' }}>
                <div
                  style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}
                  onClick={() => setExpandedIdx((prev) => {
                    const next = new Set(prev);
                    if (next.has(i)) next.delete(i); else next.add(i);
                    return next;
                  })}
                >
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: msg.role === 'user' ? 'var(--accent)' : 'var(--purple)', fontWeight: 600, minWidth: '60px' }}>
                    {msg.role}
                  </span>
                  <span style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {preview}{preview.length >= 200 ? '…' : ''}
                  </span>
                </div>
                {isExpanded && (
                  <pre style={{ ...preStyle, marginTop: '6px', maxHeight: '200px' }}>
                    {msg.content.map((c, ci) => (
                      <div key={ci}>
                        {c.type === 'text' && c.text}
                        {c.type === 'tool_use' && `[tool_use] ${c.name}\n${JSON.stringify(c.input, null, 2)}`}
                        {c.type === 'tool_result' && `[tool_result] ${c.tool_use_id}\n${c.content}`}
                      </div>
                    ))}
                  </pre>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function ToolExecutionSection({ toolResults, toolCalls }: { toolResults: LogEntry[]; toolCalls: ToolCallEntry[] }) {
  const [expandedTools, setExpandedTools] = useState<Set<string>>(new Set());

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
      {toolResults.map((tr, i) => {
        const matchingCall = toolCalls.find((tc) => tc.id === tr.tool_use_id);
        const key = `${tr.tool_use_id ?? i}`;
        const isExpanded = expandedTools.has(key);
        return (
          <div key={key} style={{ background: 'var(--bg-base)', border: `1px solid ${tr.success ? 'rgba(63,185,80,0.3)' : 'rgba(248,81,73,0.3)'}`, borderRadius: '6px', padding: '8px 12px' }}>
            <div
              style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}
              onClick={() => setExpandedTools((prev) => {
                const next = new Set(prev);
                if (next.has(key)) next.delete(key); else next.add(key);
                return next;
              })}
            >
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-primary)', fontWeight: 600 }}>{tr.tool}</span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: tr.success ? 'var(--success)' : 'var(--error)', fontWeight: 600 }}>
                {tr.success ? '✓ OK' : '✗ ERR'}
              </span>
              {tr.duration_ms !== undefined && (
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-muted)' }}>{tr.duration_ms}ms</span>
              )}
              <span style={{ marginLeft: 'auto', color: 'var(--text-muted)', fontSize: '11px' }}>{isExpanded ? '▲' : '▼'}</span>
            </div>
            {isExpanded && (
              <div style={{ marginTop: '8px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
                {matchingCall && (
                  <div>
                    <div style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)', marginBottom: '4px' }}>Input:</div>
                    <pre style={{ ...preStyle, maxHeight: '200px' }}>{JSON.stringify(matchingCall.input, null, 2)}</pre>
                  </div>
                )}
                <div>
                  <div style={{ fontFamily: 'var(--font-ui)', fontSize: '11px', color: 'var(--text-muted)', marginBottom: '4px' }}>Output:</div>
                  <pre style={{ ...preStyle, maxHeight: '200px' }}>{tr.output ?? '(no output)'}</pre>
                </div>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

function TurnAccordion({ group, expandedTurns, onToggle }: { group: TurnGroup; expandedTurns: Set<number>; onToggle: (t: number) => void }) {
  const [activeSection, setActiveSection] = useState<string | null>(null);
  const isExpanded = expandedTurns.has(group.turn);
  const req = group.llmRequest;
  const res = group.llmResponse;

  const toolCallsInResponse: ToolCallEntry[] = res?.tool_calls ?? [];

  const toggleSection = (s: string) => setActiveSection((prev) => (prev === s ? null : s));

  return (
    <div style={{ border: '1px solid var(--border)', borderRadius: '8px', overflow: 'hidden' }}>
      {/* Turn header */}
      <div
        onClick={() => onToggle(group.turn)}
        style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '10px 14px', background: 'var(--bg-surface-2)', cursor: 'pointer', userSelect: 'none' }}
      >
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--accent)', fontWeight: 700, minWidth: '60px' }}>
          Turn {group.turn}
        </span>
        {res?.stop_reason && (
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-muted)', background: 'var(--bg-base)', padding: '2px 6px', borderRadius: '4px' }}>
            {res.stop_reason}
          </span>
        )}
        {res && (
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-muted)' }}>
            {(res.input_tokens ?? 0).toLocaleString()} in / {(res.output_tokens ?? 0).toLocaleString()} out
          </span>
        )}
        {group.toolResults.length > 0 && (
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--warning)' }}>
            {group.toolResults.length} tool calls
          </span>
        )}
        {req?.consecutive_errors !== undefined && req.consecutive_errors > 0 && (
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--error)' }}>
            {req.consecutive_errors} consec. errors
          </span>
        )}
        <span style={{ marginLeft: 'auto', color: 'var(--text-muted)', fontSize: '12px' }}>{isExpanded ? '▲' : '▼'}</span>
      </div>

      {isExpanded && (
        <div style={{ padding: '12px 14px', display: 'flex', flexDirection: 'column', gap: '10px', background: 'var(--bg-surface)' }}>
          {/* System Prompt */}
          {req?.system_layers && (
            <div>
              <button
                onClick={() => toggleSection('system')}
                style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontFamily: 'var(--font-ui)', fontSize: '12px', padding: 0, marginBottom: '6px' }}
              >
                {activeSection === 'system' ? '▼' : '▶'} 📋 System Prompt
              </button>
              {activeSection === 'system' && <SystemPromptTabs layers={req.system_layers} />}
            </div>
          )}

          {/* Message History */}
          {req?.messages && req.messages.length > 0 && (
            <div>
              <div style={{ fontFamily: 'var(--font-ui)', fontSize: '12px', color: 'var(--text-muted)', marginBottom: '4px' }}>💬 Message History</div>
              <MessageHistorySection messages={req.messages} />
            </div>
          )}

          {/* Tools Available */}
          {req?.tools && req.tools.length > 0 && (
            <div>
              <button
                onClick={() => toggleSection('tools')}
                style={{ background: 'transparent', border: 'none', color: 'var(--text-muted)', cursor: 'pointer', fontFamily: 'var(--font-ui)', fontSize: '12px', padding: 0, marginBottom: '6px' }}
              >
                {activeSection === 'tools' ? '▼' : '▶'} 🔧 Tools Available ({req.tools.length})
              </button>
              {activeSection === 'tools' && (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
                  {req.tools.map((t) => (
                    <span key={t.name} style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--warning)', background: 'rgba(210,153,34,0.1)', padding: '2px 6px', borderRadius: '4px' }}>
                      {t.name}
                    </span>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* LLM Response */}
          {res && (
            <div>
              <div style={{ fontFamily: 'var(--font-ui)', fontSize: '12px', color: 'var(--text-muted)', marginBottom: '6px' }}>🤖 LLM Response</div>
              <div style={{ display: 'flex', gap: '8px', alignItems: 'center', marginBottom: '6px', flexWrap: 'wrap' }}>
                {res.stop_reason && (
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-muted)', background: 'var(--bg-base)', padding: '2px 6px', borderRadius: '4px' }}>
                    stop: {res.stop_reason}
                  </span>
                )}
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '10px', color: 'var(--text-muted)' }}>
                  {(res.input_tokens ?? 0).toLocaleString()} in / {(res.output_tokens ?? 0).toLocaleString()} out tokens
                </span>
              </div>
              {res.text && (
                <pre style={{ ...preStyle, maxHeight: '200px' }}>{res.text}</pre>
              )}
            </div>
          )}

          {/* Tool Executions */}
          {group.toolResults.length > 0 && (
            <div>
              <div style={{ fontFamily: 'var(--font-ui)', fontSize: '12px', color: 'var(--text-muted)', marginBottom: '6px' }}>⚡ Tool Executions</div>
              <ToolExecutionSection toolResults={group.toolResults} toolCalls={toolCallsInResponse} />
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export function AgentView({ agentName, logs, onBack }: AgentViewProps) {
  const [expandedTurns, setExpandedTurns] = useState<Set<number>>(new Set());

  const initEntry = logs.find((e) => e.role === 'agent_init');
  const llmResponses = logs.filter((e) => e.role === 'llm_response');
  const toolResults = logs.filter((e) => e.role === 'tool_result');

  const totalInputTokens = llmResponses.reduce((sum, e) => sum + (e.input_tokens ?? 0), 0);
  const totalOutputTokens = llmResponses.reduce((sum, e) => sum + (e.output_tokens ?? 0), 0);
  const totalCost = totalInputTokens * 3 / 1_000_000 + totalOutputTokens * 15 / 1_000_000;

  const turns = groupByTurn(logs);

  const toggleTurn = (t: number) => {
    setExpandedTurns((prev) => {
      const next = new Set(prev);
      if (next.has(t)) next.delete(t); else next.add(t);
      return next;
    });
  };

  // Sparkline data: tokens per turn
  const sparklineData = turns.map((g) => ({
    turn: g.turn,
    input: g.llmResponse?.input_tokens ?? 0,
    output: g.llmResponse?.output_tokens ?? 0,
  }));

  // Tool duration bar chart
  const toolDurationMap: Record<string, { name: string; total: number; count: number }> = {};
  for (const e of toolResults) {
    if (e.tool && e.duration_ms !== undefined) {
      if (!toolDurationMap[e.tool]) toolDurationMap[e.tool] = { name: e.tool, total: 0, count: 0 };
      toolDurationMap[e.tool].total += e.duration_ms;
      toolDurationMap[e.tool].count++;
    }
  }
  const toolDurationChartData = Object.values(toolDurationMap).map((d) => ({
    name: d.name,
    avgMs: Math.round(d.total / d.count),
  }));

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
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '6px', flexWrap: 'wrap' }}>
              <span style={{ fontFamily: 'var(--font-ui)', fontSize: '16px', fontWeight: 700, color: 'var(--text-primary)' }}>{agentName}</span>
              {initEntry?.agent_type && (
                <span style={{
                  background: initEntry.agent_type === 'leader' ? 'rgba(188,140,255,0.15)' : 'rgba(88,166,255,0.15)',
                  color: initEntry.agent_type === 'leader' ? 'var(--purple)' : 'var(--accent)',
                  fontFamily: 'var(--font-mono)', fontSize: '11px', fontWeight: 700, padding: '2px 8px', borderRadius: '4px',
                }}>
                  {initEntry.agent_type === 'leader' ? 'LEADER' : 'EXECUTOR'}
                </span>
              )}
              {initEntry?.model && (
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', color: 'var(--text-muted)', background: 'var(--bg-surface-2)', padding: '2px 6px', borderRadius: '4px' }}>
                  {initEntry.model}
                </span>
              )}
            </div>
            <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-muted)' }}>
                <span style={{ color: 'var(--text-primary)' }}>{turns.length}</span> turns
              </span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-muted)' }}>
                <span style={{ color: 'var(--text-primary)' }}>{toolResults.length}</span> tool calls
              </span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-muted)' }}>
                <span style={{ color: 'var(--text-primary)' }}>{totalInputTokens.toLocaleString()}</span> in / <span style={{ color: 'var(--text-primary)' }}>{totalOutputTokens.toLocaleString()}</span> out tokens
              </span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--accent)' }}>
                ${totalCost.toFixed(6)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Metrics panel */}
      <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
        {sparklineData.length > 0 && (
          <div style={{ ...cardStyle, flex: 1, minWidth: '240px' }}>
            <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '8px' }}>
              Tokens per Turn
            </div>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={sparklineData} margin={{ left: 0, right: 8, top: 4, bottom: 4 }}>
                <XAxis dataKey="turn" tick={{ fill: 'var(--text-muted)', fontSize: 10, fontFamily: 'var(--font-mono)' }} />
                <YAxis tick={{ fill: 'var(--text-muted)', fontSize: 10, fontFamily: 'var(--font-mono)' }} />
                <Tooltip
                  contentStyle={{ background: 'var(--bg-surface-2)', border: '1px solid var(--border)', borderRadius: '6px', fontFamily: 'var(--font-mono)', fontSize: '11px' }}
                />
                <Line type="monotone" dataKey="input" stroke="#58a6ff" strokeWidth={2} dot={false} name="Input" />
                <Line type="monotone" dataKey="output" stroke="#bc8cff" strokeWidth={2} dot={false} name="Output" />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}

        {toolDurationChartData.length > 0 && (
          <div style={{ ...cardStyle, flex: 1, minWidth: '240px' }}>
            <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '8px' }}>
              Avg Tool Duration (ms)
            </div>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={toolDurationChartData} layout="vertical" margin={{ left: 8, right: 16, top: 4, bottom: 4 }}>
                <XAxis type="number" tick={{ fill: 'var(--text-muted)', fontSize: 10, fontFamily: 'var(--font-mono)' }} />
                <YAxis type="category" dataKey="name" width={120} tick={{ fill: 'var(--text-muted)', fontSize: 10, fontFamily: 'var(--font-mono)' }} />
                <Tooltip
                  contentStyle={{ background: 'var(--bg-surface-2)', border: '1px solid var(--border)', borderRadius: '6px', fontFamily: 'var(--font-mono)', fontSize: '11px' }}
                />
                <Bar dataKey="avgMs" radius={[0, 4, 4, 0]}>
                  {toolDurationChartData.map((_, i) => (
                    <Cell key={i} fill="#d29922" />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      {/* Turn-by-turn accordion */}
      <div style={cardStyle}>
        <div style={{ fontFamily: 'var(--font-ui)', fontSize: '13px', fontWeight: 600, color: 'var(--text-primary)', marginBottom: '12px', display: 'flex', alignItems: 'center', gap: '8px' }}>
          Turn-by-Turn Detail
          <button
            onClick={() => {
              if (expandedTurns.size === turns.length) {
                setExpandedTurns(new Set());
              } else {
                setExpandedTurns(new Set(turns.map((t) => t.turn)));
              }
            }}
            style={{ background: 'transparent', border: '1px solid var(--border)', borderRadius: '4px', color: 'var(--text-muted)', cursor: 'pointer', fontFamily: 'var(--font-ui)', fontSize: '11px', padding: '2px 8px' }}
          >
            {expandedTurns.size === turns.length ? 'Collapse All' : 'Expand All'}
          </button>
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
          {turns.length === 0 ? (
            <div style={{ color: 'var(--text-muted)', fontFamily: 'var(--font-ui)', fontSize: '13px' }}>No LLM turns found in logs.</div>
          ) : (
            turns.map((group) => (
              <TurnAccordion
                key={group.turn}
                group={group}
                expandedTurns={expandedTurns}
                onToggle={toggleTurn}
              />
            ))
          )}
        </div>
      </div>
    </div>
  );
}
