import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';
import type { AgentNodeData, AgentStatus } from '../types/graph';

const STATUS_COLORS: Record<AgentStatus, string> = {
  idle: '#9ca3af',
  running: '#3b82f6',
  done: '#22c55e',
  error: '#ef4444',
};

type AgentGraphNodeProps = NodeProps & { data: AgentNodeData };

function AgentGraphNodeComponent({ data }: AgentGraphNodeProps) {
  const color = STATUS_COLORS[data.status];
  const isLeader = data.agentType === 'leader';
  const isRunning = data.status === 'running';

  const containerStyle: React.CSSProperties = {
    width: 180,
    minHeight: 60,
    background: '#1e1e2e',
    border: `2px solid ${color}`,
    borderRadius: isLeader ? '10px' : '30px',
    display: 'flex',
    alignItems: 'center',
    padding: '8px 12px',
    gap: '8px',
    animation: isRunning ? 'nodePulse 1.5s ease-in-out infinite' : 'none',
    cursor: 'default',
    userSelect: 'none',
  };

  const badgeStyle: React.CSSProperties = {
    width: 22,
    height: 22,
    borderRadius: '50%',
    background: color,
    color: '#fff',
    fontSize: '0.65rem',
    fontWeight: 700,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
  };

  const textStyle: React.CSSProperties = {
    flex: 1,
    overflow: 'hidden',
  };

  const labelStyle: React.CSSProperties = {
    fontSize: '0.78rem',
    fontWeight: 600,
    color: '#e0e0e0',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  };

  const subtitleStyle: React.CSSProperties = {
    fontSize: '0.65rem',
    color: '#888',
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    marginTop: 2,
  };

  const handleStyle: React.CSSProperties = {
    background: color,
    width: 8,
    height: 8,
    border: 'none',
  };

  return (
    <>
      <Handle type="target" position={Position.Top} style={handleStyle} />
      <div style={containerStyle}>
        <div style={badgeStyle}>{isLeader ? 'L' : 'E'}</div>
        <div style={textStyle}>
          <div style={labelStyle} title={data.label}>{data.label}</div>
          {data.lastMessage && (
            <div style={subtitleStyle} title={data.lastMessage}>{data.lastMessage}</div>
          )}
        </div>
      </div>
      <Handle type="source" position={Position.Bottom} style={handleStyle} />
    </>
  );
}

export const AgentGraphNode = memo(AgentGraphNodeComponent);
