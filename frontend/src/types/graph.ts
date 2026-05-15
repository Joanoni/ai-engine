export type AgentStatus = 'idle' | 'running' | 'done' | 'error';
export type AgentType = 'leader' | 'executor';

export interface StaticAgent {
  name: string;
  type: AgentType;
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
  type: 'agentNode';
  data: AgentNodeData;
  position: { x: number; y: number };
}

export interface AgentEdge {
  id: string;
  source: string;
  target: string;
  animated: boolean;
}
