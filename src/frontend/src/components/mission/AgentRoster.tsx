import { useMemo } from 'react';
import type { EngineEvent } from '../../types/events';

interface AgentInfo {
  name: string;
  type: 'leader' | 'executor';
  status: 'idle' | 'running' | 'done' | 'error';
  toolCallCount: number;
}

function inferType(name: string): 'leader' | 'executor' {
  const lower = name.toLowerCase();
  if (
    lower.includes('leader') ||
    lower.includes('orchestrat') ||
    lower.includes('manager') ||
    lower.includes('swarmito')
  ) {
    return 'leader';
  }
  return 'executor';
}

interface AgentRosterProps {
  events: EngineEvent[];
}

export function AgentRoster({ events }: AgentRosterProps) {
  const agents = useMemo(() => {
    const map = new Map<string, AgentInfo>();

    for (const event of events) {
      const name = event.agent_name;
      if (!name) continue;

      if (!map.has(name)) {
        map.set(name, {
          name,
          type: inferType(name),
          status: 'idle',
          toolCallCount: 0,
        });
      }

      const agent = map.get(name)!;

      switch (event.type) {
        case 'agent.started':
          map.set(name, { ...agent, status: 'running' });
          break;
        case 'agent.finished':
          map.set(name, { ...agent, status: 'done' });
          break;
        case 'tool.called':
          map.set(name, { ...agent, toolCallCount: agent.toolCallCount + 1 });
          break;
        case 'error':
          if (event.payload?.agent === name) {
            map.set(name, { ...agent, status: 'error' });
          }
          break;
      }
    }

    return Array.from(map.values());
  }, [events]);

  if (agents.length === 0) return null;

  const statusColor: Record<string, string> = {
    idle: 'var(--text-muted)',
    running: 'var(--accent)',
    done: 'var(--success)',
    error: 'var(--error)',
  };

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
        Agent Roster
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        {agents.map((agent) => (
          <div
            key={agent.name}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '6px 8px',
              borderRadius: '6px',
              background:
                agent.status === 'running'
                  ? 'rgba(88,166,255,0.08)'
                  : 'transparent',
              transition: 'background 200ms ease',
            }}
          >
            {/* Status dot */}
            <div
              style={{
                width: '7px',
                height: '7px',
                borderRadius: '50%',
                background: statusColor[agent.status],
                flexShrink: 0,
                boxShadow:
                  agent.status === 'running'
                    ? '0 0 6px var(--accent)'
                    : 'none',
              }}
            />

            {/* Name */}
            <span
              style={{
                fontFamily: 'var(--font-ui)',
                fontSize: '12px',
                color: 'var(--text-primary)',
                flex: 1,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {agent.name}
            </span>

            {/* Type badge */}
            <span
              style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '10px',
                fontWeight: 600,
                color: agent.type === 'leader' ? 'var(--purple)' : 'var(--text-muted)',
                background:
                  agent.type === 'leader'
                    ? 'rgba(188,140,255,0.12)'
                    : 'var(--bg-surface-2)',
                padding: '1px 5px',
                borderRadius: '3px',
                flexShrink: 0,
              }}
            >
              {agent.type === 'leader' ? 'L' : 'E'}
            </span>

            {/* Tool call count */}
            {agent.toolCallCount > 0 && (
              <span
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: '11px',
                  color: 'var(--warning)',
                  flexShrink: 0,
                }}
              >
                {agent.toolCallCount}🔧
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
