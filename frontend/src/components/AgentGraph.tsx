import { ReactFlow, Background, Controls, MiniMap } from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { AgentGraphNode } from './AgentGraphNode';
import type { AgentNode, AgentEdge } from '../types/graph';

interface AgentGraphProps {
  nodes: AgentNode[];
  edges: AgentEdge[];
}

const nodeTypes = { agentNode: AgentGraphNode };

export function AgentGraph({ nodes, edges }: AgentGraphProps) {
  if (nodes.length === 0) {
    return (
      <div className="graph-panel graph-empty">
        <span className="graph-empty-msg">Send a prompt to see the agent graph</span>
      </div>
    );
  }

  return (
    <div className="graph-panel">
      <ReactFlow
        nodes={nodes as Node[]}
        edges={edges as Edge[]}
        nodeTypes={nodeTypes}
        fitView
        nodesDraggable={false}
        nodesConnectable={false}
        proOptions={{ hideAttribution: true }}
      >
        <Background color="#2e2e2e" gap={20} />
        <Controls />
        <MiniMap
          nodeColor={(node) => {
            const data = node.data as { status?: string };
            switch (data.status) {
              case 'running': return '#3b82f6';
              case 'done': return '#22c55e';
              case 'error': return '#ef4444';
              default: return '#9ca3af';
            }
          }}
          style={{ background: '#1a1a1a', border: '1px solid #2e2e2e' }}
        />
      </ReactFlow>
    </div>
  );
}
