import { useEffect } from 'react';

interface ShortcutMap {
  [key: string]: () => void;
}

export function useKeyboardShortcuts(shortcuts: ShortcutMap): void {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // Build a key string like "ctrl+b", "escape", etc.
      const parts: string[] = [];
      if (e.ctrlKey || e.metaKey) parts.push('ctrl');
      if (e.shiftKey) parts.push('shift');
      if (e.altKey) parts.push('alt');
      parts.push(e.key.toLowerCase());

      const keyString = parts.join('+');

      if (shortcuts[keyString]) {
        e.preventDefault();
        shortcuts[keyString]();
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [shortcuts]);
}
