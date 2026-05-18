export type AgentStatus = 'idle' | 'running' | 'done' | 'error';
export type AgentType = 'leaderNode' | 'executorNode';

export interface StaticAgent {
  name: string;
  type: 'leader' | 'executor';
  team: string[];
}

export interface AgentNodeData extends Record<string, unknown> {
  label: string;
  agentType: AgentType;
  status: AgentStatus;
  lastMessage?: string;
}

export interface AgentNode {
  id: string;
  type: AgentType;
  data: AgentNodeData;
  position: { x: number; y: number };
}

export interface AgentEdge {
  id: string;
  source: string;
  target: string;
  animated: boolean;
  type?: string;
}
