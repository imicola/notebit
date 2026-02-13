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

/**
 * Detect node type based on properties
 * @param {object} node - Node object
 * @returns {string} Node type: 'concept', 'note', or 'tag'
 */
function detectNodeType(node) {
  // Tags typically start with #
  if (node.id?.toString().startsWith('#') || node.label?.startsWith('#')) {
    return 'tag';
  }
  
  // Concept nodes are highly connected (high degree)
  // Using size as proxy for degree (larger nodes = more connections)
  if (node.size > 5 || node.val > 5) {
    return 'concept';
  }
  
  // Default to note type
  return 'note';
}

/**
 * Get color scheme for node type
 * @param {string} type - Node type
 * @returns {object} Color configuration
 */
function getNodeColorScheme(type) {
  const colors = {
    concept: {
      background: '#00D4FF', // Aurora blue
      border: '#00D4FF',
      highlight: {
        background: '#00D4FF',
        border: '#FFFFFF',
      },
    },
    note: {
      background: '#00FF88', // Emerald green
      border: '#00FF88',
      highlight: {
        background: '#00FF88',
        border: '#FFFFFF',
      },
    },
    tag: {
      background: '#FFBF00', // Amber yellow
      border: '#FFBF00',
      highlight: {
        background: '#FFBF00',
        border: '#FFFFFF',
      },
    },
  };
  
  return colors[type] || colors.note;
}

/**
 * Enhance node with type and color information
 * @param {object} node - Raw node data
 * @returns {object} Enhanced node
 */
function enhanceNode(node) {
  const type = detectNodeType(node);
  const colorScheme = getNodeColorScheme(type);
  
  return {
    ...node,
    type,
    colorScheme,
  };
}

export const graphService = {
  /**
   * Get graph data (nodes and links)
   * @returns {Promise<{nodes: Array, links: Array}>} Graph data
   */
  async getGraphData() {
    const data = await wrapCall('getGraphData', GetGraphData);
    
    // Enhance nodes with type and color information
    const enhancedNodes = Array.isArray(data?.nodes) 
      ? data.nodes.map(enhanceNode) 
      : [];
    
    return {
      nodes: enhancedNodes,
      links: Array.isArray(data?.links) ? data.links : [],
    };
  },
};

export default graphService;
