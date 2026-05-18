import { useState, useCallback, useRef, useEffect } from 'react';

interface UseResizableReturn {
  ratio: number;
  handleMouseDown: (e: React.MouseEvent) => void;
}

export function useResizable(defaultRatio: number): UseResizableReturn {
  const [ratio, setRatio] = useState(defaultRatio);
  const isDragging = useRef(false);
  const containerRef = useRef<HTMLElement | null>(null);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging.current) return;

    // Find the cockpit container to calculate relative position
    const cockpit = document.getElementById('cockpit-area');
    if (!cockpit) return;

    const rect = cockpit.getBoundingClientRect();
    const relativeY = e.clientY - rect.top;
    const newRatio = Math.min(Math.max(relativeY / rect.height, 0.2), 0.8);
    setRatio(newRatio);
  }, []);

  const handleMouseUp = useCallback(() => {
    isDragging.current = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  }, []);

  useEffect(() => {
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [handleMouseMove, handleMouseUp]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    isDragging.current = true;
    document.body.style.cursor = 'ns-resize';
    document.body.style.userSelect = 'none';
    containerRef.current = (e.currentTarget as HTMLElement).parentElement;
  }, []);

  return { ratio, handleMouseDown };
}
