import { useState, useEffect, useRef, useCallback } from 'react';
import { fetchGraph } from '../api';
import type { GraphNode, GraphEdge, GraphData } from '../types';
import '../styles/graph.css';

const NODE_COLOURS: Record<GraphNode['type'], string> = {
  character: '#6366f1',
  location:  '#10b981',
  faction:   '#f59e0b',
  event:     '#ef4444',
};

const NODE_RADIUS = 22;
const REPULSION   = 6000;
const ATTRACTION  = 0.04;
const CENTERING   = 0.012;
const DAMPING     = 0.82;
const TICKS       = 300;
const SVG_W       = 900;
const SVG_H       = 600;

interface SimNode extends GraphNode {
  x: number;
  y: number;
  vx: number;
  vy: number;
}

function buildSim(nodes: GraphNode[]): SimNode[] {
  const cx = SVG_W / 2;
  const cy = SVG_H / 2;
  const count = nodes.length;
  return nodes.map((n, i) => {
    const angle = (2 * Math.PI * i) / count;
    const r = Math.min(cx, cy) * 0.55;
    return { ...n, x: cx + r * Math.cos(angle), y: cy + r * Math.sin(angle), vx: 0, vy: 0 };
  });
}

function runSimulation(simNodes: SimNode[], edges: GraphEdge[]): SimNode[] {
  const nodes = simNodes.map(n => ({ ...n }));
  const cx = SVG_W / 2;
  const cy = SVG_H / 2;

  for (let tick = 0; tick < TICKS; tick++) {
    // Repulsion between every pair of nodes
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const dx = nodes[i].x - nodes[j].x;
        const dy = nodes[i].y - nodes[j].y;
        const distSq = Math.max(dx * dx + dy * dy, 1);
        const dist   = Math.sqrt(distSq);
        const force  = REPULSION / distSq;
        const fx = (dx / dist) * force;
        const fy = (dy / dist) * force;
        nodes[i].vx += fx;
        nodes[i].vy += fy;
        nodes[j].vx -= fx;
        nodes[j].vy -= fy;
      }
    }

    // Attraction along edges (spring)
    const idxMap = new Map(nodes.map((n, i) => [n.id, i]));
    for (const edge of edges) {
      const si = idxMap.get(edge.source);
      const ti = idxMap.get(edge.target);
      if (si === undefined || ti === undefined) continue;
      const dx = nodes[ti].x - nodes[si].x;
      const dy = nodes[ti].y - nodes[si].y;
      nodes[si].vx += dx * ATTRACTION;
      nodes[si].vy += dy * ATTRACTION;
      nodes[ti].vx -= dx * ATTRACTION;
      nodes[ti].vy -= dy * ATTRACTION;
    }

    // Centering force
    for (const n of nodes) {
      n.vx += (cx - n.x) * CENTERING;
      n.vy += (cy - n.y) * CENTERING;
    }

    // Integrate + dampen + clamp to SVG bounds
    for (const n of nodes) {
      n.vx *= DAMPING;
      n.vy *= DAMPING;
      n.x = Math.max(NODE_RADIUS, Math.min(SVG_W - NODE_RADIUS, n.x + n.vx));
      n.y = Math.max(NODE_RADIUS, Math.min(SVG_H - NODE_RADIUS, n.y + n.vy));
    }
  }

  return nodes;
}

export default function GraphPanel() {
  const [graphData, setGraphData]       = useState<GraphData | null>(null);
  const [simNodes,  setSimNodes]        = useState<SimNode[]>([]);
  const [selected,  setSelected]        = useState<SimNode | null>(null);
  const [loading,   setLoading]         = useState(true);
  const [error,     setError]           = useState<string | null>(null);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    fetchGraph()
      .then(data => {
        if (cancelled) return;
        setGraphData(data);
        const initial = buildSim(data.nodes);
        const settled = runSimulation(initial, data.edges);
        setSimNodes(settled);
        setLoading(false);
      })
      .catch(e => {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : 'Failed to load graph');
        setLoading(false);
      });
    return () => {
      cancelled = true;
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current);
    };
  }, []);

  const handleNodeClick = useCallback((node: SimNode) => {
    setSelected(prev => (prev?.id === node.id ? null : node));
  }, []);

  if (loading) return <div className="loading-text">Computing graph…</div>;
  if (error)   return <div className="error-banner">{error}</div>;
  if (!graphData || graphData.nodes.length === 0)
    return <div className="empty-state">No graph data available.</div>;

  const edges: GraphEdge[] = graphData.edges;

  return (
    <div className="graph-container">
      {/* SVG area */}
      <div className="graph-svg-area">
        <svg
          width="100%"
          viewBox={`0 0 ${SVG_W} ${SVG_H}`}
          className="graph-svg"
          preserveAspectRatio="xMidYMid meet"
        >
          <defs>
            <marker
              id="arrowhead"
              markerWidth="8"
              markerHeight="6"
              refX="7"
              refY="3"
              orient="auto"
            >
              <polygon points="0 0, 8 3, 0 6" fill="#8b949e" />
            </marker>
          </defs>

          {/* Edges */}
          {edges.map((edge, i) => {
            const src = simNodes.find(n => n.id === edge.source);
            const tgt = simNodes.find(n => n.id === edge.target);
            if (!src || !tgt) return null;

            // Shorten line to not overlap node circles
            const dx = tgt.x - src.x;
            const dy = tgt.y - src.y;
            const dist = Math.sqrt(dx * dx + dy * dy) || 1;
            const x1 = src.x + (dx / dist) * NODE_RADIUS;
            const y1 = src.y + (dy / dist) * NODE_RADIUS;
            const x2 = tgt.x - (dx / dist) * (NODE_RADIUS + 6); // +6 for arrowhead
            const y2 = tgt.y - (dy / dist) * (NODE_RADIUS + 6);
            const mx = (src.x + tgt.x) / 2;
            const my = (src.y + tgt.y) / 2;

            return (
              <g key={i}>
                <line
                  x1={x1} y1={y1} x2={x2} y2={y2}
                  stroke="#30363d"
                  strokeWidth={1.5}
                  markerEnd="url(#arrowhead)"
                />
                <text
                  x={mx} y={my - 4}
                  textAnchor="middle"
                  fontSize={10}
                  fill="#8b949e"
                  style={{ pointerEvents: 'none', userSelect: 'none' }}
                >
                  {edge.label}
                </text>
              </g>
            );
          })}

          {/* Nodes */}
          {simNodes.map(node => (
            <g
              key={node.id}
              className="graph-node"
              onClick={() => handleNodeClick(node)}
              style={{ cursor: 'pointer' }}
            >
              <circle
                cx={node.x}
                cy={node.y}
                r={NODE_RADIUS}
                fill={NODE_COLOURS[node.type]}
                stroke={selected?.id === node.id ? '#ffffff' : 'transparent'}
                strokeWidth={2.5}
                opacity={0.92}
              />
              <text
                x={node.x}
                y={node.y + NODE_RADIUS + 14}
                textAnchor="middle"
                fontSize={11}
                fill="#e6edf3"
                style={{ pointerEvents: 'none', userSelect: 'none' }}
              >
                {node.label.length > 14 ? node.label.slice(0, 13) + '…' : node.label}
              </text>
            </g>
          ))}
        </svg>

        {/* Legend */}
        <div className="graph-legend">
          {(Object.entries(NODE_COLOURS) as [GraphNode['type'], string][]).map(([type, colour]) => (
            <span key={type} className="graph-legend-item">
              <span className="graph-legend-dot" style={{ background: colour }} />
              {type.charAt(0).toUpperCase() + type.slice(1)}
            </span>
          ))}
        </div>
      </div>

      {/* Inspector panel */}
      <div className={`graph-inspector${selected ? ' graph-inspector--visible' : ''}`}>
        {selected ? (
          <>
            <div className="graph-inspector-header">
              <span
                className="graph-inspector-dot"
                style={{ background: NODE_COLOURS[selected.type] }}
              />
              <span className="graph-inspector-title">{selected.label}</span>
              <button
                className="graph-inspector-close"
                onClick={() => setSelected(null)}
                aria-label="Close"
              >
                ✕
              </button>
            </div>
            <div className="graph-inspector-body">
              <div className="graph-inspector-row">
                <span className="graph-inspector-key">Type</span>
                <span className="graph-inspector-value">{selected.type}</span>
              </div>
              <div className="graph-inspector-row">
                <span className="graph-inspector-key">ID</span>
                <span className="graph-inspector-value graph-inspector-id">{selected.id}</span>
              </div>
              <div className="graph-inspector-row">
                <span className="graph-inspector-key">Connections</span>
                <span className="graph-inspector-value">
                  {graphData.edges.filter(
                    e => e.source === selected.id || e.target === selected.id
                  ).length}
                </span>
              </div>
            </div>
            {/* Connected edges list */}
            {graphData.edges.filter(
              e => e.source === selected.id || e.target === selected.id
            ).length > 0 && (
              <div className="graph-inspector-edges">
                <div className="graph-inspector-edges-title">Relationships</div>
                {graphData.edges
                  .filter(e => e.source === selected.id || e.target === selected.id)
                  .map((e, i) => {
                    const otherId   = e.source === selected.id ? e.target : e.source;
                    const otherNode = simNodes.find(n => n.id === otherId);
                    const direction = e.source === selected.id ? '→' : '←';
                    return (
                      <div key={i} className="graph-inspector-edge-item">
                        <span className="graph-inspector-edge-dir">{direction}</span>
                        <span
                          className="graph-inspector-edge-dot"
                          style={{ background: otherNode ? NODE_COLOURS[otherNode.type] : '#8b949e' }}
                        />
                        <span className="graph-inspector-edge-label">{e.label}</span>
                        <span className="graph-inspector-edge-name">{otherNode?.label ?? otherId}</span>
                      </div>
                    );
                  })}
              </div>
            )}
          </>
        ) : (
          <div className="graph-inspector-empty">
            <span>🔍</span>
            <p>Click a node to inspect it</p>
          </div>
        )}
      </div>
    </div>
  );
}
