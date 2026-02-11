/**
 * useToast Hook
 * Manages toast notification state
 */
import { useState, useCallback } from 'react';
import { TOAST_DURATION } from '../constants';

/**
 * Custom hook for managing toast notifications
 * @returns {Object} Toast state and handlers
 */
export const useToast = () => {
  const [toast, setToast] = useState({
    show: false,
    message: '',
    type: 'success'
  });

  // Show a toast message
  const showToast = useCallback((message, type = 'success') => {
    setToast({
      show: true,
      message,
      type
    });
  }, []);

  // Hide the current toast
  const hideToast = useCallback(() => {
    setToast(prev => ({
      ...prev,
      show: false
    }));
  }, []);

  // Show success toast (convenience method)
  const showSuccess = useCallback((message) => {
    showToast(message, 'success');
  }, [showToast]);

  // Show error toast (convenience method)
  const showError = useCallback((message) => {
    showToast(message, 'error');
  }, [showToast]);

  return {
    toast,
    showToast,
    hideToast,
    showSuccess,
    showError
  };
};

export default useToast;
