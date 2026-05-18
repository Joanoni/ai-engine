import { useEffect, useRef, useCallback } from 'react';
import type { EngineEvent } from '../../types/events';
import { TerminalLine } from './TerminalLine';

interface LiveTerminalProps {
  events: EngineEvent[];
  onClear: () => void;
}

export function LiveTerminal({ events, onClear }: LiveTerminalProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [events]);

  const handleExport = useCallback(() => {
    const lines = events.map((e) => JSON.stringify(e)).join('\n');
    const blob = new Blob([lines], { type: 'application/jsonl' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `ai-engine-events-${Date.now()}.jsonl`;
    a.click();
    URL.revokeObjectURL(url);
  }, [events]);

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        background: '#050810',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Scanline overlay */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage:
            'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.03) 2px, rgba(0,0,0,0.03) 4px)',
          pointerEvents: 'none',
          zIndex: 1,
        }}
      />

      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '8px 16px',
          borderBottom: '1px solid var(--border)',
          background: 'rgba(13,17,23,0.6)',
          flexShrink: 0,
          zIndex: 2,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <div
            style={{
              width: '8px',
              height: '8px',
              borderRadius: '50%',
              background: 'var(--success)',
              boxShadow: '0 0 6px var(--success)',
            }}
          />
          <span
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              fontWeight: 500,
              color: 'var(--text-muted)',
              letterSpacing: '0.1em',
            }}
          >
            LIVE LOG
          </span>
          <span
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              color: 'var(--text-muted)',
            }}
          >
            ({events.length} events)
          </span>
        </div>

        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            onClick={handleExport}
            disabled={events.length === 0}
            style={{
              background: 'transparent',
              border: '1px solid var(--border)',
              borderRadius: '4px',
              color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              padding: '3px 8px',
              cursor: events.length === 0 ? 'not-allowed' : 'pointer',
              opacity: events.length === 0 ? 0.4 : 1,
              transition: 'all 200ms ease',
            }}
          >
            Export
          </button>
          <button
            onClick={onClear}
            style={{
              background: 'transparent',
              border: '1px solid var(--border)',
              borderRadius: '4px',
              color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              padding: '3px 8px',
              cursor: 'pointer',
              transition: 'all 200ms ease',
            }}
          >
            Clear
          </button>
        </div>
      </div>

      {/* Log content */}
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          overflowX: 'auto',
          padding: '12px 16px',
          zIndex: 2,
        }}
      >
        {events.length === 0 ? (
          <div
            style={{
              color: 'var(--text-muted)',
              fontFamily: 'var(--font-mono)',
              fontSize: '13px',
              opacity: 0.5,
            }}
          >
            Waiting for events...
          </div>
        ) : (
          events.map((event, i) => (
            <TerminalLine key={`${event.receivedAt}-${i}`} event={event} />
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
