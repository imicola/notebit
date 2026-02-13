/**
 * File Service - Abstraction layer for file operations
 * Wraps Wails API calls with consistent error handling
 */
import {
  OpenFolder,
  ListFiles,
  ReadFile,
  SaveFile,
  CreateFile,
  DeleteFile,
  RenameFile,
  GetBasePath,
  SetFolder
} from '../../wailsjs/go/main/App';

/**
 * Custom error class for file operations
 */
export class FileServiceError extends Error {
  constructor(operation, originalError) {
    super(`File operation failed: ${operation}`);
    this.name = 'FileServiceError';
    this.operation = operation;
    this.originalError = originalError;
  }
}

/**
 * Wrap API call with error handling
 */
const wrapCall = async (operation, apiCall) => {
  try {
    return await apiCall();
  } catch (error) {
    throw new FileServiceError(operation, error);
  }
};

/**
 * File service object with all file operations
 */
export const fileService = {
  /**
   * Open folder dialog and set as base path
   * @returns {Promise<string|null>} Selected folder path or null if cancelled
   */
  async openFolder() {
    return wrapCall('openFolder', OpenFolder);
  },

  /**
   * Set the working folder path
   * @param {string} path - Folder path to set
   */
  async setFolder(path) {
    return wrapCall('setFolder', () => SetFolder(path));
  },

  /**
   * Get file tree structure
   * @returns {Promise<FileNode>} File tree root node
   */
  async listFiles() {
    return wrapCall('listFiles', ListFiles);
  },

  /**
   * Read file content
   * @param {string} path - Relative file path
   * @returns {Promise<NoteContent>} File content object
   */
  async readFile(path) {
    return wrapCall('readFile', () => ReadFile(path));
  },

  /**
   * Save content to file
   * @param {string} path - Relative file path
   * @param {string} content - File content
   */
  async saveFile(path, content) {
    return wrapCall('saveFile', () => SaveFile(path, content));
  },

  /**
   * Create new file
   * @param {string} path - Relative file path
   * @param {string} content - Initial content
   */
  async createFile(path, content) {
    return wrapCall('createFile', () => CreateFile(path, content));
  },

  /**
   * Delete file or directory
   * @param {string} path - Relative path to delete
   */
  async deleteFile(path) {
    return wrapCall('deleteFile', () => DeleteFile(path));
  },

  /**
   * Rename file or directory
   * @param {string} oldPath - Current path
   * @param {string} newPath - New path
   */
  async renameFile(oldPath, newPath) {
    return wrapCall('renameFile', () => RenameFile(oldPath, newPath));
  },

  /**
   * Get current base path
   * @returns {Promise<string>} Current base path
   */
  async getBasePath() {
    return wrapCall('getBasePath', GetBasePath);
  }
};

export default fileService;
