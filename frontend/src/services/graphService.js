/**
 * Graph Service - Abstraction layer for knowledge graph operations
 * Wraps Wails API calls with consistent error handling
 */
import { GetGraphData } from '../../wailsjs/go/main/App';

/**
 * Custom error class for graph operations
 */
export class GraphServiceError extends Error {
  constructor(operation, originalError) {
    super(`Graph operation failed: ${operation}`);
    this.name = 'GraphServiceError';
    this.operation = operation;
    this.originalError = originalError;
  }
}

const wrapCall = async (operation, apiCall) => {
  try {
    return await apiCall();
  } catch (error) {
    throw new GraphServiceError(operation, error);
  }
};

export const graphService = {
  /**
   * Get graph data (nodes and links)
   * @returns {Promise<{nodes: Array, links: Array}>} Graph data
   */
  async getGraphData() {
    const data = await wrapCall('getGraphData', GetGraphData);
    return {
      nodes: Array.isArray(data?.nodes) ? data.nodes : [],
      links: Array.isArray(data?.links) ? data.links : [],
    };
  },
};

export default graphService;
