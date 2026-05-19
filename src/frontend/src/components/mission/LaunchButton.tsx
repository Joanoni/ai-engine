import { useState, useCallback, useRef } from 'react';

interface LaunchButtonProps {
  onClick: () => void;
  disabled: boolean;
  isRunning: boolean;
}

export function LaunchButton({ onClick, disabled, isRunning }: LaunchButtonProps) {
  const [countdown, setCountdown] = useState<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleClick = useCallback(() => {
    if (disabled || isRunning || countdown !== null) return;

    setCountdown('3...');
    timerRef.current = setTimeout(() => {
      setCountdown('2...');
      timerRef.current = setTimeout(() => {
        setCountdown('1...');
        timerRef.current = setTimeout(() => {
          setCountdown('🚀 Launching');
          timerRef.current = setTimeout(() => {
            setCountdown(null);
            onClick();
          }, 200);
        }, 200);
      }, 200);
    }, 200);
  }, [disabled, isRunning, countdown, onClick]);

  const isDisabled = disabled && !isRunning;
  const label = countdown
    ? countdown
    : isRunning
    ? '⏳ Mission Running...'
    : '🚀 Launch Mission';

  return (
    <button
      onClick={handleClick}
      disabled={isDisabled || isRunning || countdown !== null}
      style={{
        width: '100%',
        height: '48px',
        borderRadius: '8px',
        background: isDisabled
          ? 'var(--bg-surface-2)'
          : 'linear-gradient(135deg, #1d4ed8, #58a6ff)',
        color: isDisabled ? 'var(--text-muted)' : '#fff',
        fontFamily: 'var(--font-ui)',
        fontWeight: 600,
        fontSize: '15px',
        cursor: isDisabled || isRunning || countdown !== null ? 'not-allowed' : 'pointer',
        border: isRunning ? '2px solid var(--accent)' : '2px solid transparent',
        transition: 'all 200ms ease',
        animation: isRunning ? 'pulse-border 2s ease-in-out infinite' : 'none',
        letterSpacing: '0.02em',
      }}
    >
      {label}
    </button>
  );
}
