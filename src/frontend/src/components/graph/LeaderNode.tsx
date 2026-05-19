import { useState } from 'react';
import { Handle, Position, type NodeProps, type Node } from '@xyflow/react';
import type { AgentNodeData } from '../../types/graph';

type LeaderNodeType = Node<AgentNodeData & { toolCallCount?: number }, 'leaderNode'>;

const statusGlow: Record<string, string> = {
  idle:    'none',
  running: '0 0 24px rgba(188,140,255,0.6), 0 0 48px rgba(188,140,255,0.2)',
  done:    '0 0 16px rgba(63,185,80,0.4)',
  error:   '0 0 16px rgba(248,81,73,0.4)',
};

const statusBorderColor: Record<string, string> = {
  idle:    'rgba(188,140,255,0.3)',
  running: 'rgba(188,140,255,0.9)',
  done:    'rgba(63,185,80,0.7)',
  error:   'rgba(248,81,73,0.7)',
};

const statusLabelColor: Record<string, string> = {
  idle:    'var(--text-muted)',
  running: 'var(--purple)',
  done:    'var(--success)',
  error:   'var(--error)',
};

export function LeaderNode({ data }: NodeProps<LeaderNodeType>) {
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
        background: 'linear-gradient(135deg, rgba(188,140,255,0.08) 0%, rgba(13,17,23,0.97) 100%)',
        backdropFilter: 'blur(16px)',
        border: `1px solid ${borderColor}`,
        borderRadius: '14px',
        padding: '18px 22px 16px',
        minWidth: '220px',
        maxWidth: '260px',
        position: 'relative',
        boxShadow: glow,
        animation: isRunning ? 'pulse-glow-purple 2s ease-in-out infinite' : 'node-appear 0.4s ease-out',
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
        background: 'radial-gradient(ellipse at 30% 0%, rgba(188,140,255,0.12), transparent 70%)',
        pointerEvents: 'none',
      }} />

      {/* Top accent line — full width gradient */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: '3px',
        background: `linear-gradient(90deg, transparent 0%, ${borderColor} 40%, ${borderColor} 60%, transparent 100%)`,
        borderRadius: '14px 14px 0 0',
        opacity: 0.9,
      }} />

      {/* Tool call count badge */}
      {toolCallCount !== undefined && toolCallCount > 0 && (
        <div style={{
          position: 'absolute',
          top: '10px',
          right: '10px',
          fontSize: '10px',
          color: 'var(--purple)',
          fontFamily: 'var(--font-mono)',
          background: 'rgba(188,140,255,0.1)',
          border: '1px solid rgba(188,140,255,0.25)',
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
          width: 10,
          height: 10,
          top: -5,
        }}
      />

      <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '8px' }}>
        {/* Hexagon badge — 32×32 */}
        <div
          style={{
            width: '32px',
            height: '32px',
            background: 'linear-gradient(135deg, #bc8cff, #7c3aed)',
            clipPath: 'polygon(50% 0%, 100% 25%, 100% 75%, 50% 100%, 0% 75%, 0% 25%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            boxShadow: '0 0 10px rgba(188,140,255,0.55)',
          }}
        >
          <span style={{ color: '#fff', fontSize: '11px', fontWeight: 800, fontFamily: 'var(--font-mono)', lineHeight: 1 }}>
            L
          </span>
        </div>

        <span
          style={{
            color: 'var(--text-primary)',
            fontFamily: 'var(--font-ui)',
            fontWeight: 700,
            fontSize: '13px',
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
        paddingLeft: '42px',
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
            background: 'var(--purple)',
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
          width: 10,
          height: 10,
          bottom: -5,
        }}
      />
    </div>
  );
}
