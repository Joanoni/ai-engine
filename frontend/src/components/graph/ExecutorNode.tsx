import { useState } from 'react';
import { Handle, Position, type NodeProps, type Node } from '@xyflow/react';
import type { AgentNodeData } from '../../types/graph';

type ExecutorNodeType = Node<AgentNodeData & { toolCallCount?: number }, 'executorNode'>;

const statusBorderColor: Record<string, string> = {
  idle:    'rgba(88,166,255,0.25)',
  running: 'rgba(88,166,255,0.9)',
  done:    'rgba(63,185,80,0.7)',
  error:   'rgba(248,81,73,0.7)',
};

const statusGlow: Record<string, string> = {
  idle:    'none',
  running: '0 0 20px rgba(88,166,255,0.5), 0 0 40px rgba(88,166,255,0.15)',
  done:    '0 0 12px rgba(63,185,80,0.3)',
  error:   '0 0 12px rgba(248,81,73,0.3)',
};

const statusBarColor: Record<string, string> = {
  idle:    'rgba(125,133,144,0.3)',
  running: 'var(--accent)',
  done:    'var(--success)',
  error:   'var(--error)',
};

const statusBarShadow: Record<string, string> = {
  idle:    'none',
  running: '0 0 8px rgba(88,166,255,0.7)',
  done:    '0 0 6px rgba(63,185,80,0.5)',
  error:   '0 0 6px rgba(248,81,73,0.5)',
};

const statusLabelColor: Record<string, string> = {
  idle:    'var(--text-muted)',
  running: 'var(--accent)',
  done:    'var(--success)',
  error:   'var(--error)',
};

export function ExecutorNode({ data }: NodeProps<ExecutorNodeType>) {
  const { label, status, toolCallCount } = data;
  const isRunning = status === 'running';
  const borderColor = statusBorderColor[status] ?? statusBorderColor.idle;
  const glow = statusGlow[status] ?? 'none';
  const [hovered, setHovered] = useState(false);

  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: 'linear-gradient(135deg, rgba(88,166,255,0.06) 0%, rgba(13,17,23,0.97) 100%)',
        backdropFilter: 'blur(16px)',
        border: `1px solid ${borderColor}`,
        borderRadius: '40px',
        padding: '12px 20px 10px',
        minWidth: '190px',
        maxWidth: '230px',
        position: 'relative',
        boxShadow: glow,
        animation: isRunning ? 'pulse-glow 2s ease-in-out infinite' : 'node-appear 0.4s ease-out',
        transition: 'all 300ms ease',
        transform: hovered ? 'translateY(-1px)' : 'translateY(0)',
        overflow: 'hidden',
      }}
    >
      {/* Inner radial glow layer */}
      <div style={{
        position: 'absolute',
        inset: 0,
        borderRadius: 'inherit',
        background: 'radial-gradient(ellipse at 30% 0%, rgba(88,166,255,0.10), transparent 70%)',
        pointerEvents: 'none',
      }} />

      {/* Bottom status bar */}
      <div style={{
        position: 'absolute',
        bottom: 0,
        left: 0,
        right: 0,
        height: '3px',
        background: statusBarColor[status] ?? statusBarColor.idle,
        boxShadow: statusBarShadow[status] ?? 'none',
        transition: 'background 300ms ease, box-shadow 300ms ease',
        borderRadius: '0 0 40px 40px',
      }} />

      {/* Tool call count badge */}
      {toolCallCount !== undefined && toolCallCount > 0 && (
        <div style={{
          position: 'absolute',
          top: '8px',
          right: '12px',
          fontSize: '10px',
          color: 'var(--accent)',
          fontFamily: 'var(--font-mono)',
          background: 'rgba(88,166,255,0.1)',
          border: '1px solid rgba(88,166,255,0.2)',
          borderRadius: '4px',
          padding: '1px 5px',
          lineHeight: 1.4,
        }}>
          {toolCallCount}×
        </div>
      )}

      <Handle
        type="target"
        position={Position.Top}
        style={{
          background: borderColor,
          border: '2px solid var(--bg-base)',
          width: 8,
          height: 8,
          top: -4,
        }}
      />

      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '5px' }}>
        {/* Circle badge — 28×28 */}
        <div
          style={{
            width: '28px',
            height: '28px',
            borderRadius: '50%',
            background: 'linear-gradient(135deg, #58a6ff, #1d4ed8)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            boxShadow: '0 0 8px rgba(88,166,255,0.45)',
          }}
        >
          <span style={{ color: '#fff', fontSize: '10px', fontWeight: 800, fontFamily: 'var(--font-mono)', lineHeight: 1 }}>
            E
          </span>
        </div>

        <span
          style={{
            color: 'var(--text-primary)',
            fontFamily: 'var(--font-ui)',
            fontWeight: 600,
            fontSize: '12px',
            letterSpacing: '0.02em',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {label}
        </span>
      </div>

      <div style={{
        fontFamily: 'var(--font-mono)',
        fontSize: '10px',
        color: statusLabelColor[status] ?? 'var(--text-muted)',
        letterSpacing: '0.08em',
        paddingLeft: '36px',
        display: 'flex',
        alignItems: 'center',
        gap: '5px',
      }}>
        {isRunning && (
          <span style={{
            display: 'inline-block',
            width: '6px',
            height: '6px',
            borderRadius: '50%',
            background: 'var(--accent)',
            animation: 'blink 1s ease-in-out infinite',
            flexShrink: 0,
          }} />
        )}
        {isRunning ? 'RUNNING' : status === 'done' ? '✓ DONE' : status === 'error' ? '✕ ERROR' : '● IDLE'}
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        style={{
          background: borderColor,
          border: '2px solid var(--bg-base)',
          width: 8,
          height: 8,
          bottom: -4,
        }}
      />
    </div>
  );
}
