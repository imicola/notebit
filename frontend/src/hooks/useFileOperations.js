/**
 * useFileOperations Hook
 * Manages file operations state and handlers
 */
import { useState, useCallback } from 'react';
import { fileService } from '../services/fileService';
import { createAsyncHandler, ERROR_MESSAGES } from '../utils/asyncHandler';

/**
 * Custom hook for managing file operations
 * @param {Object} options - Configuration options
 * @param {Function} options.onSuccess - Callback for successful operations
 * @returns {Object} File state and handlers
 */
export const useFileOperations = (options = {}) => {
  const { onSuccess } = options;

  // File state
  const [fileTree, setFileTree] = useState(null);
  const [currentFile, setCurrentFile] = useState(null);
  const [currentContent, setCurrentContent] = useState('');
  const [basePath, setBasePath] = useState('');

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Create handler factory with shared state setters
  const createHandler = useCallback((asyncFn, errorMessage, successCallback) => {
    return createAsyncHandler(asyncFn, {
      setLoading,
      setError,
      errorMessage,
      onSuccess: successCallback
    });
  }, []);

  // Refresh file tree
  const refreshFileTree = useCallback(async () => {
    const tree = await fileService.listFiles();
    setFileTree(tree);
    return tree;
  }, []);

  // Open folder
  const openFolder = useCallback(
    createHandler(
      async () => {
        const path = await fileService.openFolder();
        if (path) {
          setBasePath(path);
          await refreshFileTree();
        }
        return path;
      },
      ERROR_MESSAGES.OPEN_FOLDER
    ),
    [refreshFileTree]
  );

  // Select and load file
  const selectFile = useCallback(
    createHandler(
      async (node) => {
        if (node.isDir) return null;
        const result = await fileService.readFile(node.path);
        setCurrentFile(node);
        setCurrentContent(result.content);
        return result;
      },
      ERROR_MESSAGES.READ_FILE
    ),
    []
  );

  // Save file
  const saveFile = useCallback(
    createHandler(
      async (content) => {
        if (!currentFile) return null;
        await fileService.saveFile(currentFile.path, content);
        setCurrentContent(content);
        onSuccess?.('File saved successfully');
        return true;
      },
      ERROR_MESSAGES.SAVE_FILE
    ),
    [currentFile, onSuccess]
  );

  // Update content (for editor onChange)
  const updateContent = useCallback((content) => {
    setCurrentContent(content);
  }, []);

  // Clear current file
  const clearFile = useCallback(() => {
    setCurrentFile(null);
    setCurrentContent('');
  }, []);

  return {
    // State
    fileTree,
    currentFile,
    currentContent,
    basePath,
    loading,
    error,

    // Handlers
    openFolder,
    selectFile,
    saveFile,
    updateContent,
    refreshFileTree,
    clearFile,

    // State setters (for advanced use)
    setFileTree,
    setCurrentFile,
    setCurrentContent,
    setBasePath,
    setError
  };
};

export default useFileOperations;
