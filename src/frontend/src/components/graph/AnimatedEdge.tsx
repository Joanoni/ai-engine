import { getBezierPath, type EdgeProps } from '@xyflow/react';

export function AnimatedEdge({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  animated,
}: EdgeProps) {
  const [edgePath] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <>
      {/* Glow halo layer when animated */}
      {animated && (
        <path
          d={edgePath}
          stroke="var(--accent)"
          strokeWidth={6}
          fill="none"
          opacity={0.15}
          style={{ animation: 'edge-glow-pulse 2s ease-in-out infinite' }}
        />
      )}

      {/* Base path */}
      <path
        id={id}
        className="react-flow__edge-path"
        d={edgePath}
        stroke={animated ? 'rgba(88,166,255,0.6)' : 'rgba(33,38,45,0.6)'}
        strokeWidth={animated ? 2 : 1.5}
        fill="none"
        opacity={1}
      />

      {/* Animated dashes when running */}
      {animated && (
        <>
          <path
            d={edgePath}
            stroke="var(--accent)"
            strokeWidth={2}
            fill="none"
            strokeDasharray="8 12"
            style={{ animation: 'dash-flow 0.8s linear infinite' }}
            opacity={0.8}
          />

          {/* Particle 1 */}
          <circle r="3" fill="var(--accent)" style={{ filter: 'drop-shadow(0 0 4px var(--accent))' }}>
            <animateMotion dur="1.2s" repeatCount="indefinite" path={edgePath} />
          </circle>

          {/* Particle 2 — offset */}
          <circle r="2.5" fill="var(--accent)" style={{ filter: 'drop-shadow(0 0 3px var(--accent))', opacity: 0.7 }}>
            <animateMotion dur="1.8s" begin="0.6s" repeatCount="indefinite" path={edgePath} />
          </circle>
        </>
      )}
    </>
  );
}
