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
  SetAIModel,
  GetChunkingConfig,
  SetChunkingConfig,
  TestOpenAIConnection,
  GetLLMConfig,
  SetLLMConfig,
  GetRAGConfig,
  SetRAGConfig,
  GetGraphConfig,
  SetGraphConfig,
  GetSimilarityStatus,
  ReindexAllWithEmbeddings
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

const callAppMethod = async (method, ...args) => {
  const fn = window?.go?.main?.App?.[method];
  if (typeof fn !== 'function') {
    throw new Error(`App method not available: ${method}`);
  }
  return fn(...args);
};

export const aiService = {
  // --- Status ---
  async getStatus() {
    return wrapCall('getAIStatus', GetAIStatus);
  },

  async getSimilarityStatus() {
    return wrapCall('getSimilarityStatus', GetSimilarityStatus);
  },

  async reindexAllWithEmbeddings() {
    return wrapCall('reindexAllWithEmbeddings', ReindexAllWithEmbeddings);
  },

  async getVectorSearchEngine() {
    return wrapCall('getVectorSearchEngine', () => callAppMethod('GetVectorSearchEngine'));
  },

  async setVectorSearchEngine(engine) {
    return wrapCall('setVectorSearchEngine', () => callAppMethod('SetVectorSearchEngine', engine));
  },

  // --- OpenAI ---
  async getOpenAIConfig() {
    return wrapCall('getOpenAIConfig', GetOpenAIConfig);
  },

  async setOpenAIConfig(apiKey, baseURL, organization, embeddingModel) {
    return wrapCall('setOpenAIConfig', () => SetOpenAIConfig(apiKey, baseURL, organization, embeddingModel));
  },

  async testOpenAIConnection(apiKey, baseURL, organization, model) {
    return wrapCall('testOpenAIConnection', () => TestOpenAIConnection(apiKey, baseURL, organization, model));
  },

  // --- Ollama ---
  async getOllamaConfig() {
    return wrapCall('getOllamaConfig', GetOllamaConfig);
  },

  async setOllamaConfig(baseURL, model, timeout) {
    return wrapCall('setOllamaConfig', () => SetOllamaConfig(baseURL, model, timeout));
  },

  // --- Provider ---
  async setProvider(provider) {
    return wrapCall('setAIProvider', () => SetAIProvider(provider));
  },

  async setAIModel(model) {
    return wrapCall('setAIModel', () => SetAIModel(model));
  },

  // --- Chunking ---
  async getChunkingConfig() {
    return wrapCall('getChunkingConfig', GetChunkingConfig);
  },

  async setChunkingConfig(strategy, chunkSize, chunkOverlap, minChunkSize, maxChunkSize, preserveHeading, headingSeparator) {
    return wrapCall('setChunkingConfig', () => SetChunkingConfig(
      strategy,
      chunkSize,
      chunkOverlap,
      minChunkSize,
      maxChunkSize,
      preserveHeading,
      headingSeparator
    ));
  },

  // --- LLM ---
  async getLLMConfig() {
    return wrapCall('getLLMConfig', GetLLMConfig);
  },

  async setLLMConfig(provider, model, temperature, maxTokens, apiKey, baseURL, organization) {
    return wrapCall('setLLMConfig', () => SetLLMConfig(
      provider,
      model,
      temperature,
      maxTokens,
      apiKey,
      baseURL,
      organization
    ));
  },

  // --- RAG ---
  async getRAGConfig() {
    return wrapCall('getRAGConfig', GetRAGConfig);
  },

  async setRAGConfig(maxContextChunks, temperature) {
    return wrapCall('setRAGConfig', () => SetRAGConfig(maxContextChunks, temperature, ''));
  },

  // --- Graph Config ---
  async getGraphConfig() {
    return wrapCall('getGraphConfig', GetGraphConfig);
  },

  async setGraphConfig(minSimilarityThreshold, maxNodes, showImplicitLinks) {
    return wrapCall('setGraphConfig', () => SetGraphConfig(minSimilarityThreshold, maxNodes, showImplicitLinks));
  },
};

export default aiService;
