import { useRef, useCallback } from 'react';

interface PromptEditorProps {
  onSend: (text: string) => void;
  disabled: boolean;
}

export function PromptEditor({ onSend, disabled }: PromptEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleInput = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = 'auto';
    el.style.height = `${Math.min(el.scrollHeight, 240)}px`;
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        const value = textareaRef.current?.value.trim();
        if (value && !disabled) {
          onSend(value);
          if (textareaRef.current) {
            textareaRef.current.value = '';
            textareaRef.current.style.height = 'auto';
          }
        }
      }
    },
    [onSend, disabled]
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
      <label
        style={{
          fontFamily: 'var(--font-ui)',
          fontSize: '11px',
          fontWeight: 600,
          color: 'var(--text-muted)',
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
        }}
      >
        Mission Brief
      </label>
      <textarea
        ref={textareaRef}
        disabled={disabled}
        placeholder="Describe your mission..."
        onInput={handleInput}
        onKeyDown={handleKeyDown}
        style={{
          background: 'var(--bg-surface-2)',
          border: '1px solid var(--border)',
          borderRadius: '8px',
          color: 'var(--text-primary)',
          fontFamily: 'var(--font-ui)',
          fontSize: '14px',
          lineHeight: '1.6',
          minHeight: '120px',
          maxHeight: '240px',
          padding: '12px',
          resize: 'none',
          width: '100%',
          transition: 'border-color 200ms ease',
          opacity: disabled ? 0.5 : 1,
          cursor: disabled ? 'not-allowed' : 'text',
        }}
        onFocus={(e) => {
          e.currentTarget.style.borderColor = 'var(--accent)';
        }}
        onBlur={(e) => {
          e.currentTarget.style.borderColor = 'var(--border)';
        }}
      />
      <div
        style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '11px',
          color: 'var(--text-muted)',
          textAlign: 'right',
        }}
      >
        Ctrl+Enter to launch
      </div>
    </div>
  );
}
