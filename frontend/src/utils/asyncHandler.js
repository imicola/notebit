/**
 * Async Handler Utility
 * Provides consistent error handling pattern for async operations
 */

/**
 * Error messages for different operations
 */
export const ERROR_MESSAGES = {
  openFolder: 'Failed to open folder',
  loadFiles: 'Failed to load files',
  readFile: 'Failed to read file',
  saveFile: 'Failed to save file',
  default: 'An error occurred'
};

/**
 * Create an async handler with consistent error handling
 *
 * @param {Function} asyncFn - The async function to wrap
 * @param {Object} options - Configuration options
 * @param {Function} options.setLoading - Function to set loading state
 * @param {Function} options.setError - Function to set error state
 * @param {string} options.errorMessage - Custom error message
 * @param {Function} options.onSuccess - Callback on success
 * @param {Function} options.onError - Callback on error
 * @returns {Function} Wrapped handler function
 */
export const createAsyncHandler = (asyncFn, options = {}) => {
  const {
    setLoading = () => {},
    setError = () => {},
    errorMessage = ERROR_MESSAGES.default,
    onSuccess = () => {},
    onError = () => {}
  } = options;

  return async (...args) => {
    setLoading(true);
    setError(null);

    try {
      const result = await asyncFn(...args);
      await onSuccess(result);
      return result;
    } catch (err) {
      const message = err?.message || errorMessage;
      setError(message);
      console.error(`${errorMessage}:`, err);
      onError(err);
      throw err;
    } finally {
      setLoading(false);
    }
  };
};

