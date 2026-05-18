import { useMemo } from 'react';
import type { EngineEvent } from '../../types/events';

interface TaskItem {
  text: string;
  state: 'done' | 'in-progress' | 'pending';
}

function parseTaskList(content: string): TaskItem[] {
  const lines = content.split('\n');
  const items: TaskItem[] = [];

  for (const line of lines) {
    const trimmed = line.trim();
    if (trimmed.startsWith('- [x]') || trimmed.startsWith('- [X]')) {
      items.push({ text: trimmed.slice(6).trim(), state: 'done' });
    } else if (trimmed.startsWith('- [-]')) {
      items.push({ text: trimmed.slice(6).trim(), state: 'in-progress' });
    } else if (trimmed.startsWith('- [ ]')) {
      items.push({ text: trimmed.slice(6).trim(), state: 'pending' });
    }
  }

  return items;
}

interface TaskProgressProps {
  events: EngineEvent[];
}

export function TaskProgress({ events }: TaskProgressProps) {
  const tasks = useMemo(() => {
    // Find the latest tasks.updated event
    const latest = [...events].reverse().find((e) => e.type === 'tasks.updated');
    if (!latest?.payload?.content) return [];
    return parseTaskList(String(latest.payload.content));
  }, [events]);

  if (tasks.length === 0) return null;

  const done = tasks.filter((t) => t.state === 'done').length;
  const total = tasks.length;
  const pct = total > 0 ? (done / total) * 100 : 0;

  const stateIcon: Record<string, string> = {
    done: '✓',
    'in-progress': '⟳',
    pending: '○',
  };

  const stateColor: Record<string, string> = {
    done: 'var(--success)',
    'in-progress': 'var(--warning)',
    pending: 'var(--text-muted)',
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
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
        Task Progress
      </div>

      {/* Progress bar */}
      <div>
        <div
          style={{
            height: '4px',
            background: 'var(--bg-surface-2)',
            borderRadius: '2px',
            overflow: 'hidden',
            marginBottom: '6px',
          }}
        >
          <div
            style={{
              height: '100%',
              width: `${pct}%`,
              background: 'var(--success)',
              borderRadius: '2px',
              transition: 'width 400ms ease',
            }}
          />
        </div>
        <div
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '11px',
            color: 'var(--text-muted)',
            textAlign: 'right',
          }}
        >
          {done}/{total} completed
        </div>
      </div>

      {/* Task list */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        {tasks.map((task, i) => (
          <div
            key={i}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              fontFamily: 'var(--font-ui)',
              fontSize: '12px',
            }}
          >
            <span
              style={{
                color: stateColor[task.state],
                flexShrink: 0,
                width: '14px',
                textAlign: 'center',
              }}
            >
              {stateIcon[task.state]}
            </span>
            <span
              style={{
                color: task.state === 'done' ? 'var(--text-muted)' : 'var(--text-primary)',
                textDecoration: task.state === 'done' ? 'line-through' : 'none',
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
              title={task.text}
            >
              {task.text.slice(0, 40)}
              {task.text.length > 40 ? '…' : ''}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
