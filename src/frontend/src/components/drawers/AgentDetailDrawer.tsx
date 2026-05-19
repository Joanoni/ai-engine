import type { EngineEvent } from '../../types/events';

interface AgentDetailDrawerProps {
  agent: string | null;
  events: EngineEvent[];
  onClose: () => void;
}

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

function inferType(name: string): 'L' | 'E' {
  const lower = name.toLowerCase();
  if (
    lower.includes('leader') ||
    lower.includes('orchestrat') ||
    lower.includes('manager') ||
    lower.includes('swarmito')
  ) {
    return 'L';
  }
  return 'E';
}

interface EventItemProps {
  event: EngineEvent;
}

function EventItem({ event }: EventItemProps) {
  const time = formatTime(event.receivedAt ?? event.timestamp ?? '');
  const p = event.payload;

  let color = 'var(--text-muted)';
  let title: string = event.type;
  let detail = '';

  switch (event.type) {
    case 'agent.started':
      color = 'var(--purple)';
      title = 'Agent Started';
      detail = p?.triggered_by ? `triggered by ${p.triggered_by}` : '';
      break;
    case 'agent.finished':
      color = 'var(--success)';
      title = 'Agent Finished';
      detail = p?.result ? String(p.result).slice(0, 200) : '';
      break;
    case 'tool.called':
      color = 'var(--warning)';
      title = `Tool: ${p?.tool ?? 'unknown'}`;
      detail = p?.call_id ? `call_id: ${p.call_id}` : '';
      break;
    case 'tool.result':
      color = 'var(--warning)';
      title = 'Tool Result';
      detail = p?.call_id ? `[${p.call_id}] ${String(p.result ?? '').slice(0, 200)}` : String(p?.result ?? '').slice(0, 200);
      break;
    case 'error':
      color = 'var(--error)';
      title = 'Error occurred';
      detail = String(p?.message ?? p?.error ?? '');
      break;
    default:
      detail = p ? JSON.stringify(p).slice(0, 100) : '';
  }

  return (
    <div
      style={{
        padding: '10px 0',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        gap: '12px',
      }}
    >
      <span
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          color: 'var(--text-muted)',
          flexShrink: 0,
          paddingTop: '2px',
        }}
      >
        {time}
      </span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            fontFamily: 'var(--font-ui)',
            fontSize: '13px',
            fontWeight: 600,
            color,
            marginBottom: detail ? '4px' : 0,
          }}
        >
          {title}
        </div>
        {detail && (
          <div
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              color: 'var(--text-muted)',
              wordBreak: 'break-word',
              lineHeight: 1.5,
            }}
          >
            {detail}
          </div>
        )}
      </div>
    </div>
  );
}

export function AgentDetailDrawer({ agent, events, onClose }: AgentDetailDrawerProps) {
  if (!agent) return null;

  const agentEvents = events.filter((e) => e.agent_name === agent);
  const typeLabel = inferType(agent);
  const typeColor = typeLabel === 'L' ? 'var(--purple)' : 'var(--accent)';

  return (
    <div
      style={{
        position: 'fixed',
        top: 0,
        right: 0,
        bottom: 0,
        width: '420px',
        background: 'var(--bg-surface)',
        borderLeft: '1px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        animation: 'slide-in-right 250ms ease',
        zIndex: 100,
        boxShadow: '-8px 0 32px rgba(0,0,0,0.4)',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '16px',
          borderBottom: '1px solid var(--border)',
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          flexShrink: 0,
        }}
      >
        {/* Type badge */}
        <div
          style={{
            width: '28px',
            height: '28px',
            borderRadius: typeLabel === 'L' ? '6px' : '50%',
            background: typeColor,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          <span
            style={{
              color: '#000',
              fontSize: '12px',
              fontWeight: 700,
              fontFamily: 'var(--font-ui)',
            }}
          >
            {typeLabel}
          </span>
        </div>

        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            style={{
              fontFamily: 'var(--font-ui)',
              fontWeight: 700,
              fontSize: '14px',
              color: 'var(--text-primary)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {agent}
          </div>
          <div
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              color: 'var(--text-muted)',
            }}
          >
            {agentEvents.length} events
          </div>
        </div>

        <button
          onClick={onClose}
          style={{
            background: 'transparent',
            border: '1px solid var(--border)',
            borderRadius: '6px',
            color: 'var(--text-muted)',
            fontFamily: 'var(--font-ui)',
            fontSize: '16px',
            width: '32px',
            height: '32px',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            transition: 'all 200ms ease',
          }}
          onMouseEnter={(e) => {
            (e.currentTarget as HTMLButtonElement).style.borderColor = 'var(--error)';
            (e.currentTarget as HTMLButtonElement).style.color = 'var(--error)';
          }}
          onMouseLeave={(e) => {
            (e.currentTarget as HTMLButtonElement).style.borderColor = 'var(--border)';
            (e.currentTarget as HTMLButtonElement).style.color = 'var(--text-muted)';
          }}
        >
          ×
        </button>
      </div>

      {/* Event list */}
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '0 16px',
        }}
      >
        {agentEvents.length === 0 ? (
          <div
            style={{
              padding: '24px 0',
              color: 'var(--text-muted)',
              fontFamily: 'var(--font-ui)',
              fontSize: '13px',
              textAlign: 'center',
            }}
          >
            No events for this agent
          </div>
        ) : (
          agentEvents.map((event, i) => (
            <EventItem key={`${event.receivedAt ?? event.timestamp}-${i}`} event={event} />
          ))
        )}
      </div>
    </div>
  );
}
