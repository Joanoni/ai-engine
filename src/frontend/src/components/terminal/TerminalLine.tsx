import type { EngineEvent } from '../../types/events';

interface TerminalLineProps {
  event: EngineEvent;
}

const typeColors: Record<string, string> = {
  'session.started':  'var(--accent)',
  'session.finished': 'var(--accent)',
  'agent.started':    'var(--purple)',
  'agent.finished':   'var(--purple)',
  'tool.called':      'var(--warning)',
  'tool.result':      'var(--warning)',
  'tasks.updated':    '#2dd4bf',
  'warning':          '#f0a500',
  'error':            'var(--error)',
};

const typeLabels: Record<string, string> = {
  'session.started':  'SESSION',
  'session.finished': 'SESSION',
  'agent.started':    'AGENT  ',
  'agent.finished':   'AGENT  ',
  'tool.called':      'TOOL   ',
  'tool.result':      'RESULT ',
  'tasks.updated':    'TASKS  ',
  'warning':          'WARN   ',
  'error':            'ERROR  ',
};

function formatTime(isoString: string): string {
  try {
    const d = new Date(isoString);
    const hh = String(d.getHours()).padStart(2, '0');
    const mm = String(d.getMinutes()).padStart(2, '0');
    const ss = String(d.getSeconds()).padStart(2, '0');
    return `${hh}:${mm}:${ss}`;
  } catch {
    return '??:??:??';
  }
}

function buildMessage(event: EngineEvent): string {
  const p = event.payload;
  if (!p) return event.type;

  switch (event.type) {
    case 'session.started':
      return `Session started${p.prompt ? ` — ${String(p.prompt).slice(0, 80)}` : ''}`;
    case 'session.finished':
      return `Session finished${p.status ? ` [${p.status}]` : ''}`;
    case 'agent.started':
      return `Started${p.triggered_by ? ` (triggered by ${p.triggered_by})` : ''}`;
    case 'agent.finished':
      return `Finished${p.result ? ` — ${String(p.result).slice(0, 80)}` : ''}`;
    case 'tool.called':
      return `${p.tool ?? 'unknown'}${p.call_id ? ` [${p.call_id}]` : ''}`;
    case 'tool.result':
      return `${p.call_id ? `[${p.call_id}] ` : ''}${String(p.result ?? '').slice(0, 100)}`;
    case 'tasks.updated':
      return String(p.content ?? '').replace(/\n/g, ' ').slice(0, 100);
    case 'warning':
      return String(p.warning ?? p.message ?? 'Unknown warning').slice(0, 100);
    case 'error':
      return String(p.message ?? p.error ?? 'Unknown error').slice(0, 100);
    default:
      return JSON.stringify(p).slice(0, 100);
  }
}

export function TerminalLine({ event }: TerminalLineProps) {
  const color = typeColors[event.type] ?? 'var(--text-muted)';
  const label = typeLabels[event.type] ?? event.type.padEnd(7);
  const time = formatTime(event.receivedAt ?? event.timestamp ?? '');
  const agentName = event.agent_name ?? '—';
  const message = buildMessage(event);

  return (
    <div
      style={{
        display: 'flex',
        gap: '12px',
        padding: '2px 0',
        fontFamily: 'var(--font-mono)',
        fontSize: '13px',
        lineHeight: '1.6',
        whiteSpace: 'nowrap',
        color: 'var(--text-primary)',
      }}
    >
      <span style={{ color: 'var(--text-muted)', flexShrink: 0 }}>{time}</span>
      <span style={{ color, flexShrink: 0, fontWeight: 500 }}>[{label}]</span>
      <span style={{ color: 'var(--text-muted)', flexShrink: 0 }}>{agentName}</span>
      <span style={{ color: 'var(--text-muted)', flexShrink: 0 }}>›</span>
      <span style={{ color }}>{message}</span>
    </div>
  );
}
