import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  useReactFlow,
  type NodeTypes,
  type EdgeTypes,
  type Node,
  type Edge,
  type ReactFlowInstance,
} from '@xyflow/react';
import '@xyflow/react/dist/base.css';
import { useEffect, useRef } from 'react';
import type { AgentNode, AgentEdge } from '../../types/graph';
import { LeaderNode } from './LeaderNode';
import { ExecutorNode } from './ExecutorNode';
import { AnimatedEdge } from './AnimatedEdge';

const nodeTypes: NodeTypes = {
  leaderNode: LeaderNode,
  executorNode: ExecutorNode,
};

const edgeTypes: EdgeTypes = {
  animatedEdge: AnimatedEdge,
};

interface AgentGraphProps {
  nodes: AgentNode[];
  edges: AgentEdge[];
  onNodeClick: (agentName: string) => void;
}

export function AgentGraph({ nodes, edges, onNodeClick }: AgentGraphProps) {
  const isEmpty = nodes.length === 0;
  const rfNodes = nodes as unknown as Node[];
  const rfEdges = edges as unknown as Edge[];
  const { fitView } = useReactFlow();
  const instanceRef = useRef<ReactFlowInstance | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Re-fit whenever nodes change
  useEffect(() => {
    if (nodes.length === 0) return;
    const timer = setTimeout(() => {
      fitView({ padding: 0.4, duration: 400 });
    }, 150);
    return () => clearTimeout(timer);
  }, [nodes, fitView]);

  // Re-fit whenever the container is resized (e.g. panel drag)
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const observer = new ResizeObserver(() => {
      setTimeout(() => {
        fitView({ padding: 0.4, duration: 300 });
      }, 50);
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [fitView]);

  function handleInit(instance: ReactFlowInstance) {
    instanceRef.current = instance;
    setTimeout(() => {
      instance.fitView({ padding: 0.4, duration: 400 });
    }, 100);
  }

  return (
    <div ref={containerRef} style={{ width: '100%', height: '100%', position: 'relative' }}>
      <ReactFlow
        nodes={rfNodes}
        edges={rfEdges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={true}
        onInit={handleInit}
        onNodeClick={(_event, node) => onNodeClick(node.id)}
        style={{ background: 'var(--bg-base)' }}
        proOptions={{ hideAttribution: true }}
      >
        <Background
          variant={BackgroundVariant.Lines}
          gap={40}
          size={0.5}
          color="rgba(33,38,45,0.4)"
        />
        <Controls
          style={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border)',
            borderRadius: '8px',
          }}
        />
      </ReactFlow>

      {/* Radial depth overlay */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          background: 'radial-gradient(ellipse at 50% 40%, rgba(88,166,255,0.04) 0%, transparent 70%)',
          pointerEvents: 'none',
        }}
      />

      {isEmpty && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            pointerEvents: 'none',
          }}
        >
          <div style={{ textAlign: 'center', color: 'var(--text-muted)', fontFamily: 'var(--font-ui)' }}>
            <svg width="48" height="48" viewBox="0 0 48 48" fill="none" style={{ marginBottom: '16px', opacity: 0.4 }}>
              <circle cx="24" cy="8" r="5" stroke="currentColor" strokeWidth="1.5"/>
              <circle cx="8" cy="32" r="5" stroke="currentColor" strokeWidth="1.5"/>
              <circle cx="40" cy="32" r="5" stroke="currentColor" strokeWidth="1.5"/>
              <line x1="24" y1="13" x2="8" y2="27" stroke="currentColor" strokeWidth="1.5" strokeDasharray="3 2"/>
              <line x1="24" y1="13" x2="40" y2="27" stroke="currentColor" strokeWidth="1.5" strokeDasharray="3 2"/>
            </svg>
            <div style={{ fontSize: '14px', letterSpacing: '0.04em' }}>Launch a mission to see the agent graph</div>
          </div>
        </div>
      )}
    </div>
  );
}
