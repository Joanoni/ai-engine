import { useMemo, useState, useEffect } from 'react';
import dagre from '@dagrejs/dagre';
import type { EngineEvent } from '../types/events';
import type { AgentNode, AgentEdge, AgentStatus, AgentType, StaticAgent } from '../types/graph';

const NODE_WIDTH = 260;
const NODE_HEIGHT = 100;

function applyDagreLayout(nodes: AgentNode[], edges: AgentEdge[]): AgentNode[] {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: 'TB', nodesep: 80, ranksep: 120 });

  nodes.forEach((node) => {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  });

  edges.forEach((edge) => {
    g.setEdge(edge.source, edge.target);
  });

  dagre.layout(g);

  return nodes.map((node) => {
    const pos = g.node(node.id);
    return {
      ...node,
      position: {
        x: pos ? pos.x - NODE_WIDTH / 2 : 0,
        y: pos ? pos.y - NODE_HEIGHT / 2 : 0,
      },
    };
  });
}

function inferAgentType(agentName: string): AgentType {
  const lower = agentName.toLowerCase();
  if (
    lower.includes('leader') ||
    lower.includes('orchestrat') ||
    lower.includes('manager') ||
    lower.includes('swarmito')
  ) {
    return 'leaderNode';
  }
  return 'executorNode';
}

function toNodeType(apiType: string): AgentType {
  if (apiType === 'leader') return 'leaderNode';
  if (apiType === 'executor') return 'executorNode';
  return inferAgentType(apiType);
}

export function useAgentGraph(events: EngineEvent[]): { nodes: AgentNode[]; edges: AgentEdge[] } {
  const [staticAgents, setStaticAgents] = useState<StaticAgent[]>([]);

  useEffect(() => {
    fetch('/agents')
      .then((res) => res.json())
      .then((data: { agents: StaticAgent[] }) => {
        if (Array.isArray(data.agents)) {
          setStaticAgents(data.agents);
        }
      })
      .catch(() => {
        // silently ignore — graph will be populated by events
      });
  }, []);

  return useMemo(() => {
    const nodeMap = new Map<string, AgentNode>();
    const edgeMap = new Map<string, AgentEdge>();

    function upsertNode(id: string, patch: Partial<AgentNode['data']>) {
      const existing = nodeMap.get(id);
      const agentType = patch.agentType ?? inferAgentType(id);
      if (existing) {
        nodeMap.set(id, {
          ...existing,
          type: patch.agentType ?? existing.type,
          data: { ...existing.data, ...patch },
        });
      } else {
        nodeMap.set(id, {
          id,
          type: agentType,
          position: { x: 0, y: 0 },
          data: {
            label: id,
            agentType,
            status: 'idle',
            ...patch,
          },
        });
      }
    }

    function upsertEdge(source: string, target: string, animated: boolean) {
      const edgeId = `${source}->${target}`;
      edgeMap.set(edgeId, {
        id: edgeId,
        source,
        target,
        animated,
        type: 'animatedEdge',
      });
    }

    function setNodeStatus(id: string, status: AgentStatus) {
      const node = nodeMap.get(id);
      if (node) {
        nodeMap.set(id, { ...node, data: { ...node.data, status } });
      }
    }

    function setIncomingEdgesAnimated(targetId: string, animated: boolean) {
      edgeMap.forEach((edge, key) => {
        if (edge.target === targetId) {
          edgeMap.set(key, { ...edge, animated });
        }
      });
    }

    function populateFromStatic() {
      for (const agent of staticAgents) {
        upsertNode(agent.name, { agentType: toNodeType(agent.type), status: 'idle' });
        for (const member of agent.team) {
          upsertEdge(agent.name, member, false);
        }
      }
    }

    // Seed graph from static agent tree
    populateFromStatic();

    for (const event of events) {
      switch (event.type) {
        case 'session.started': {
          nodeMap.clear();
          edgeMap.clear();
          // Re-populate from static data so graph never goes empty
          populateFromStatic();
          break;
        }

        case 'agent.started': {
          const agentName = event.agent_name;
          if (!agentName) break;
          upsertNode(agentName, { status: 'running' });

          const triggeredBy = event.payload?.triggered_by as string | undefined;
          if (triggeredBy) {
            if (!nodeMap.has(triggeredBy)) {
              upsertNode(triggeredBy, { status: 'idle' });
            }
            upsertEdge(triggeredBy, agentName, true);
          }
          break;
        }

        case 'tool.called': {
          const tool = event.payload?.tool as string | undefined;
          if (tool !== 'create_chat') break;

          const callerAgent = event.agent_name;
          const inputRaw = event.payload?.input;
          if (!inputRaw) break;

          let targetAgentName: string | undefined;
          try {
            const parsed =
              typeof inputRaw === 'string' ? JSON.parse(inputRaw) : inputRaw;
            targetAgentName = parsed?.agent_name as string | undefined;
          } catch {
            // not valid JSON, skip
          }

          if (!targetAgentName) break;

          if (!nodeMap.has(targetAgentName)) {
            upsertNode(targetAgentName, { status: 'idle' });
          }

          if (callerAgent) {
            upsertEdge(callerAgent, targetAgentName, true);
          }
          break;
        }

        case 'agent.finished': {
          const agentName = event.agent_name;
          if (!agentName) break;
          setNodeStatus(agentName, 'done');
          setIncomingEdgesAnimated(agentName, false);
          break;
        }

        case 'error': {
          const agentName = event.payload?.agent as string | undefined;
          if (agentName) {
            setNodeStatus(agentName, 'error');
          }
          break;
        }

        case 'session.finished': {
          nodeMap.forEach((node, id) => {
            if (node.data.status === 'running') {
              setNodeStatus(id, 'done');
            }
          });
          edgeMap.forEach((edge, key) => {
            edgeMap.set(key, { ...edge, animated: false });
          });
          break;
        }
      }
    }

    const rawNodes = Array.from(nodeMap.values());
    const edges = Array.from(edgeMap.values());
    const nodes = applyDagreLayout(rawNodes, edges);

    return { nodes, edges };
  }, [events, staticAgents]);
}
