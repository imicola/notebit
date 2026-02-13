/**
 * Similarity Service - Abstraction layer for semantic search operations
 * Wraps Wails API calls with consistent error handling
 */
import { FindSimilar, GetSimilarityStatus } from '../../wailsjs/go/main/App';

/**
 * Custom error class for similarity operations
 */
export class SimilarityServiceError extends Error {
  constructor(operation, originalError) {
    super(`Similarity operation failed: ${operation}`);
    this.name = 'SimilarityServiceError';
    this.operation = operation;
    this.originalError = originalError;
  }
}

const wrapCall = async (operation, apiCall) => {
  try {
    return await apiCall();
  } catch (error) {
    throw new SimilarityServiceError(operation, error);
  }
};

export const similarityService = {
  /**
   * Find similar notes by content
   * @param {string} content - Content to search for
   * @param {number} limit - Maximum results to return
   * @returns {Promise<Array>} Array of similar note results
   */
  async findSimilar(content, limit) {
    return wrapCall('findSimilar', () => FindSimilar(content, limit));
  },

  /**
   * Check if similarity search is available
   * @returns {Promise<{available: boolean, db_initialized: boolean}>} Status
   */
  async getStatus() {
    return wrapCall('getStatus', GetSimilarityStatus);
  },
};

export default similarityService;
