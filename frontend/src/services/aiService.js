/**
 * AI Service - Abstraction layer for AI configuration and status operations
 * Wraps Wails API calls with consistent error handling
 */
import {
  GetAIStatus,
  GetOpenAIConfig,
  SetOpenAIConfig,
  GetOllamaConfig,
  SetOllamaConfig,
  SetAIProvider,
  GetChunkingConfig,
  SetChunkingConfig,
  TestOpenAIConnection,
  GetLLMConfig,
  SetLLMConfig,
  GetRAGConfig,
  SetRAGConfig,
  GetGraphConfig,
  SetGraphConfig,
  GetSimilarityStatus
} from '../../wailsjs/go/main/App';

/**
 * Custom error class for AI operations
 */
export class AIServiceError extends Error {
  constructor(operation, originalError) {
    super(`AI operation failed: ${operation}`);
    this.name = 'AIServiceError';
    this.operation = operation;
    this.originalError = originalError;
  }
}

const wrapCall = async (operation, apiCall) => {
  try {
    return await apiCall();
  } catch (error) {
    throw new AIServiceError(operation, error);
  }
};

export const aiService = {
  // --- Status ---
  async getStatus() {
    return wrapCall('getAIStatus', GetAIStatus);
  },

  async getSimilarityStatus() {
    return wrapCall('getSimilarityStatus', GetSimilarityStatus);
  },

  // --- OpenAI ---
  async getOpenAIConfig() {
    return wrapCall('getOpenAIConfig', GetOpenAIConfig);
  },

  async setOpenAIConfig(config) {
    return wrapCall('setOpenAIConfig', () => SetOpenAIConfig(config));
  },

  async testOpenAIConnection() {
    return wrapCall('testOpenAIConnection', TestOpenAIConnection);
  },

  // --- Ollama ---
  async getOllamaConfig() {
    return wrapCall('getOllamaConfig', GetOllamaConfig);
  },

  async setOllamaConfig(config) {
    return wrapCall('setOllamaConfig', () => SetOllamaConfig(config));
  },

  // --- Provider ---
  async setProvider(provider) {
    return wrapCall('setAIProvider', () => SetAIProvider(provider));
  },

  // --- Chunking ---
  async getChunkingConfig() {
    return wrapCall('getChunkingConfig', GetChunkingConfig);
  },

  async setChunkingConfig(config) {
    return wrapCall('setChunkingConfig', () => SetChunkingConfig(config));
  },

  // --- LLM ---
  async getLLMConfig() {
    return wrapCall('getLLMConfig', GetLLMConfig);
  },

  async setLLMConfig(config) {
    return wrapCall('setLLMConfig', () => SetLLMConfig(config));
  },

  // --- RAG ---
  async getRAGConfig() {
    return wrapCall('getRAGConfig', GetRAGConfig);
  },

  async setRAGConfig(config) {
    return wrapCall('setRAGConfig', () => SetRAGConfig(config));
  },

  // --- Graph Config ---
  async getGraphConfig() {
    return wrapCall('getGraphConfig', GetGraphConfig);
  },

  async setGraphConfig(config) {
    return wrapCall('setGraphConfig', () => SetGraphConfig(config));
  },
};

export default aiService;
