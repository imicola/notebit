/**
 * RAG Service - Abstraction layer for RAG chat operations
 * Wraps Wails API calls with consistent error handling
 */
import { RAGQuery } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

/**
 * Custom error class for RAG operations
 */
export class RAGServiceError extends Error {
  constructor(operation, originalError) {
    super(`RAG operation failed: ${operation}`);
    this.name = 'RAGServiceError';
    this.operation = operation;
    this.originalError = originalError;
  }
}

const wrapCall = async (operation, apiCall) => {
  try {
    return await apiCall();
  } catch (error) {
    throw new RAGServiceError(operation, error);
  }
};

export const ragService = {
  /**
   * Send a RAG query
   * @param {string} query - The query text
   * @returns {Promise<Object>} RAG response with content and sources
   */
  async query(query) {
    return wrapCall('ragQuery', () => RAGQuery(query));
  },

  /**
   * Subscribe to streaming RAG chunks
   * @param {function} callback - Called with each chunk data { messageId, content }
   * @returns {function} Cleanup function to unsubscribe
   */
  onChunk(callback) {
    return EventsOn('rag_chunk', callback);
  },
};

export default ragService;
