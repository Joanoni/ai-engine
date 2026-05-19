import type { AgentNode, AgentEdge } from '../../types/graph';
import type { EngineEvent } from '../../types/events';
import { AgentGraph } from '../graph/AgentGraph';
import { LiveTerminal } from '../terminal/LiveTerminal';
import { ResizeHandle } from './ResizeHandle';
import { useResizable } from '../../hooks/useResizable';

interface CockpitAreaProps {
  nodes: AgentNode[];
  edges: AgentEdge[];
  events: EngineEvent[];
  terminalEvents: EngineEvent[];
  onNodeClick: (agentName: string) => void;
  onClearEvents: () => void;
}

export function CockpitArea({
  nodes,
  edges,
  events: _events,
  terminalEvents,
  onNodeClick,
  onClearEvents,
}: CockpitAreaProps) {
  const { ratio, handleMouseDown } = useResizable(0.6);

  return (
    <div
      id="cockpit-area"
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        minWidth: 0,
      }}
    >
      {/* Agent Graph zone */}
      <div
        style={{
          height: `${ratio * 100}%`,
          overflow: 'hidden',
          flexShrink: 0,
        }}
      >
        <AgentGraph nodes={nodes} edges={edges} onNodeClick={onNodeClick} />
      </div>

      {/* Resize handle */}
      <ResizeHandle onMouseDown={handleMouseDown} />

      {/* Live Terminal zone */}
      <div
        style={{
          flex: 1,
          overflow: 'hidden',
          minHeight: 0,
        }}
      >
        <LiveTerminal events={terminalEvents} onClear={onClearEvents} />
      </div>
    </div>
  );
}
