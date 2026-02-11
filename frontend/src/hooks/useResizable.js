/**
 * useResizable Hook
 * Manages resizable panel behavior
 */
import { useState, useCallback, useEffect } from 'react';
import { SIDEBAR, STORAGE_KEYS } from '../constants';

/**
 * Custom hook for managing resizable sidebar
 * @param {Object} options - Configuration options
 * @param {number} options.defaultWidth - Default width
 * @param {number} options.minWidth - Minimum width
 * @param {number} options.maxWidth - Maximum width
 * @param {boolean} options.persist - Whether to persist width to localStorage
 * @returns {Object} Resize state and handlers
 */
export const useResizable = (options = {}) => {
  const {
    defaultWidth = SIDEBAR.DEFAULT_WIDTH,
    minWidth = SIDEBAR.MIN_WIDTH,
    maxWidth = SIDEBAR.MAX_WIDTH,
    persist = true
  } = options;

  const [width, setWidth] = useState(() => {
    if (persist) {
      const saved = localStorage.getItem(STORAGE_KEYS.SIDEBAR_WIDTH);
      if (saved) {
        const parsed = parseInt(saved, 10);
        if (!isNaN(parsed) && parsed >= minWidth && parsed <= maxWidth) {
          return parsed;
        }
      }
    }
    return defaultWidth;
  });

  const [isResizing, setIsResizing] = useState(false);

  // Start resizing
  const startResizing = useCallback(() => {
    setIsResizing(true);
  }, []);

  // Stop resizing
  const stopResizing = useCallback(() => {
    setIsResizing(false);
  }, []);

  // Handle resize movement
  const resize = useCallback(
    (clientX) => {
      if (!isResizing) return;
      const newWidth = Math.max(minWidth, Math.min(clientX, maxWidth));
      setWidth(newWidth);

      if (persist) {
        localStorage.setItem(STORAGE_KEYS.SIDEBAR_WIDTH, newWidth.toString());
      }
    },
    [isResizing, minWidth, maxWidth, persist]
  );

  // Reset to default
  const resetWidth = useCallback(() => {
    setWidth(defaultWidth);
    if (persist) {
      localStorage.setItem(STORAGE_KEYS.SIDEBAR_WIDTH, defaultWidth.toString());
    }
  }, [defaultWidth, persist]);

  // Set up global mouse event listeners
  useEffect(() => {
    if (!isResizing) return;

    const handleMouseMove = (e) => resize(e.clientX);
    const handleMouseUp = stopResizing;

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [isResizing, resize, stopResizing]);

  return {
    width,
    isResizing,
    startResizing,
    stopResizing,
    resize,
    resetWidth
  };
};

export default useResizable;
