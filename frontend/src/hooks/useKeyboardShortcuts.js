/**
 * useKeyboardShortcuts Hook
 * Manages global keyboard shortcuts
 */
import { useEffect } from 'react';

/**
 * Custom hook for managing keyboard shortcuts
 * @param {Object} shortcuts - Map of key combinations to handlers
 * @param {Object} options - Configuration options
 * @param {boolean} options.enabled - Whether shortcuts are enabled
 */
export const useKeyboardShortcuts = (shortcuts = {}, options = {}) => {
  const { enabled = true } = options;

  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (e) => {
      // Build key combination string
      const parts = [];
      if (e.metaKey || e.ctrlKey) parts.push('Mod');
      if (e.altKey) parts.push('Alt');
      if (e.shiftKey) parts.push('Shift');
      parts.push(e.key);

      const combo = parts.join('+');

      // Check for direct match
      if (shortcuts[combo]) {
        e.preventDefault();
        shortcuts[combo](e);
        return;
      }

      // Check for Mod+key pattern (Cmd on Mac, Ctrl on Windows)
      const modCombo = `${e.metaKey || e.ctrlKey ? 'Mod+' : ''}${e.key}`;
      if (shortcuts[modCombo]) {
        e.preventDefault();
        shortcuts[modCombo](e);
        return;
      }

      // Check for single key matches (like F11)
      if (shortcuts[e.key]) {
        e.preventDefault();
        shortcuts[e.key](e);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [shortcuts, enabled]);
};

export default useKeyboardShortcuts;
