/**
 * useAISettings Hook
 * Manages all AI settings state and persistence via aiService
 */
import { useState, useEffect, useCallback } from 'react';
import { aiService } from '../services/aiService';

export const useAISettings = () => {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState(null);
  const [testingOpenAI, setTestingOpenAI] = useState(false);
  const [openaiTestResult, setOpenaiTestResult] = useState(null);
  const [similarityStatus, setSimilarityStatus] = useState(null);
  const [vectorEngine, setVectorEngine] = useState('brute-force');
  const [availableVectorEngines, setAvailableVectorEngines] = useState(['brute-force', 'sqlite-vec']);
  const [reindexing, setReindexing] = useState(false);
  const [reindexResult, setReindexResult] = useState(null);

  // Config States
  const [provider, setProvider] = useState('ollama');
  const [openaiConfig, setOpenaiConfig] = useState({
    api_key: '', base_url: '', organization: '', embedding_model: ''
  });
  const [ollamaConfig, setOllamaConfig] = useState({
    base_url: '', embedding_model: '', timeout: 30
  });
  const defaultChunking = {
    strategy: 'heading',
    chunk_size: 1000,
    chunk_overlap: 200,
    min_chunk_size: 100,
    max_chunk_size: 4000,
    preserve_heading: true,
    heading_separator: '\n\n'
  };
  const [chunkingConfig, setChunkingConfig] = useState(defaultChunking);
  const defaultLLM = {
    provider: 'openai',
    model: 'gpt-4o-mini',
    temperature: 0.7,
    max_tokens: 2000
  };
  const [llmConfig, setLLMConfig] = useState(defaultLLM);
  const [llmOpenAIConfig, setLLMOpenAIConfig] = useState({
    api_key: '', base_url: '', organization: ''
  });
  const defaultRAG = { max_context_chunks: 5, temperature: 0.7 };
  const defaultGraph = {
    min_similarity_threshold: 0.75,
    max_nodes: 100,
    show_implicit_links: true
  };
  const [ragConfig, setRAGConfig] = useState(defaultRAG);
  const [graphConfig, setGraphConfig] = useState(defaultGraph);

  // Load all settings on mount
  const toInt = (value, fallback) => {
    const parsed = Number.parseInt(value, 10);
    return Number.isFinite(parsed) ? parsed : fallback;
  };
  const toFloat = (value, fallback) => {
    const parsed = Number.parseFloat(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  };

  const loadSettings = useCallback(async () => {
    setLoading(true);
    try {
      const [aiStatus, openai, ollama, chunking, llm, rag, graph, similarity, vectorEngineStatus] = await Promise.all([
        aiService.getStatus(),
        aiService.getOpenAIConfig(),
        aiService.getOllamaConfig(),
        aiService.getChunkingConfig(),
        aiService.getLLMConfig(),
        aiService.getRAGConfig(),
        aiService.getGraphConfig(),
        aiService.getSimilarityStatus(),
        aiService.getVectorSearchEngine()
      ]);

      setStatus(aiStatus);
      setSimilarityStatus(similarity);
      setVectorEngine(vectorEngineStatus?.current || similarity?.vector_engine || 'brute-force');
      setAvailableVectorEngines(vectorEngineStatus?.available || ['brute-force', 'sqlite-vec']);
      setProvider(aiStatus.current_provider || 'ollama');
      setOpenaiConfig(openai);
      setOllamaConfig(ollama);
      setChunkingConfig({
        strategy: chunking?.strategy || defaultChunking.strategy,
        chunk_size: chunking?.chunk_size > 0 ? chunking.chunk_size : defaultChunking.chunk_size,
        chunk_overlap: chunking?.chunk_overlap >= 0 ? chunking.chunk_overlap : defaultChunking.chunk_overlap,
        min_chunk_size: chunking?.min_chunk_size > 0 ? chunking.min_chunk_size : defaultChunking.min_chunk_size,
        max_chunk_size: chunking?.max_chunk_size > 0 ? chunking.max_chunk_size : defaultChunking.max_chunk_size,
        preserve_heading: typeof chunking?.preserve_heading === 'boolean' ? chunking.preserve_heading : defaultChunking.preserve_heading,
        heading_separator: chunking?.heading_separator || defaultChunking.heading_separator
      });

      setLLMConfig({
        provider: llm?.provider || defaultLLM.provider,
        model: llm?.model || defaultLLM.model,
        temperature: Number.isFinite(llm?.temperature) ? llm.temperature : defaultLLM.temperature,
        max_tokens: llm?.max_tokens > 0 ? llm.max_tokens : defaultLLM.max_tokens
      });

      if (llm.openai) {
        setLLMOpenAIConfig({
          api_key: llm.openai.api_key || '',
          base_url: llm.openai.base_url || '',
          organization: llm.openai.organization || ''
        });
      }

      setRAGConfig({
        max_context_chunks: rag?.max_context_chunks > 0 ? rag.max_context_chunks : defaultRAG.max_context_chunks,
        temperature: Number.isFinite(rag?.temperature) ? rag.temperature : defaultRAG.temperature
      });
      setGraphConfig({
        min_similarity_threshold: Number.isFinite(graph?.min_similarity_threshold) ? graph.min_similarity_threshold : defaultGraph.min_similarity_threshold,
        max_nodes: graph?.max_nodes > 0 ? graph.max_nodes : defaultGraph.max_nodes,
        show_implicit_links: typeof graph?.show_implicit_links === 'boolean' ? graph.show_implicit_links : defaultGraph.show_implicit_links
      });
    } catch (error) {
      console.error('Failed to load settings:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadSettings(); }, [loadSettings]);

  // Save all settings
  const handleSave = useCallback(async () => {
    setSaving(true);
    try {
      if (provider === 'openai') {
        await aiService.setOpenAIConfig(openaiConfig.api_key, openaiConfig.base_url, openaiConfig.organization, openaiConfig.embedding_model);
      } else {
        await aiService.setOllamaConfig(
          ollamaConfig.base_url,
          ollamaConfig.embedding_model,
          toInt(ollamaConfig.timeout, 30)
        );
      }

      await aiService.setProvider(provider);
      const selectedEmbeddingModel = provider === 'openai' ? openaiConfig.embedding_model : ollamaConfig.embedding_model;
      if (selectedEmbeddingModel) {
        await aiService.setAIModel(selectedEmbeddingModel);
      }

      await aiService.setChunkingConfig(
        chunkingConfig.strategy,
        toInt(chunkingConfig.chunk_size, defaultChunking.chunk_size),
        toInt(chunkingConfig.chunk_overlap, defaultChunking.chunk_overlap),
        toInt(chunkingConfig.min_chunk_size, defaultChunking.min_chunk_size),
        toInt(chunkingConfig.max_chunk_size, defaultChunking.max_chunk_size),
        chunkingConfig.preserve_heading,
        chunkingConfig.heading_separator
      );

      await aiService.setLLMConfig(
        llmConfig.provider,
        llmConfig.model,
        toFloat(llmConfig.temperature, defaultLLM.temperature),
        toInt(llmConfig.max_tokens, defaultLLM.max_tokens),
        llmOpenAIConfig.api_key, llmOpenAIConfig.base_url, llmOpenAIConfig.organization
      );

      await aiService.setRAGConfig(
        toInt(ragConfig.max_context_chunks, defaultRAG.max_context_chunks),
        toFloat(ragConfig.temperature, defaultRAG.temperature)
      );
      await aiService.setGraphConfig(
        toFloat(graphConfig.min_similarity_threshold, defaultGraph.min_similarity_threshold),
        toInt(graphConfig.max_nodes, defaultGraph.max_nodes),
        graphConfig.show_implicit_links
      );

      await aiService.setVectorSearchEngine(vectorEngine);

      const newStatus = await aiService.getStatus();
      setStatus(newStatus);
      const similarity = await aiService.getSimilarityStatus();
      setSimilarityStatus(similarity);
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  }, [provider, openaiConfig, ollamaConfig, chunkingConfig, llmConfig, llmOpenAIConfig, ragConfig, graphConfig, vectorEngine]);

  // Test OpenAI connection
  const handleTestOpenAI = useCallback(async () => {
    setTestingOpenAI(true);
    setOpenaiTestResult(null);
    try {
      const result = await aiService.testOpenAIConnection(
        openaiConfig.api_key, openaiConfig.base_url,
        openaiConfig.organization, openaiConfig.embedding_model
      );
      setOpenaiTestResult({ ok: true, message: `Connected (${result.model || 'ok'}, ${result.dimension || 0}d)` });
    } catch (error) {
      setOpenaiTestResult({ ok: false, message: error?.message || 'Connection failed' });
    } finally {
      setTestingOpenAI(false);
    }
  }, [openaiConfig]);

  const handleReindexEmbeddings = useCallback(async () => {
    setReindexing(true);
    setReindexResult(null);
    try {
      const result = await aiService.reindexAllWithEmbeddings();
      setReindexResult({ ok: true, result });
      const similarity = await aiService.getSimilarityStatus();
      setSimilarityStatus(similarity);
      if (similarity?.vector_engine) {
        setVectorEngine(similarity.vector_engine);
      }
      const latestStatus = await aiService.getStatus();
      setStatus(latestStatus);
    } catch (error) {
      setReindexResult({
        ok: false,
        message: error?.message || 'Reindex failed'
      });
    } finally {
      setReindexing(false);
    }
  }, []);

  return {
    // Loading/saving state
    loading, saving, status,
    similarityStatus,
    vectorEngine,
    setVectorEngine,
    availableVectorEngines,
    reindexing,
    reindexResult,
    // Provider
    provider, setProvider,
    // OpenAI
    openaiConfig, setOpenaiConfig,
    testingOpenAI, openaiTestResult, handleTestOpenAI,
    handleReindexEmbeddings,
    // Ollama
    ollamaConfig, setOllamaConfig,
    // Chunking
    chunkingConfig, setChunkingConfig,
    // LLM
    llmConfig, setLLMConfig,
    llmOpenAIConfig, setLLMOpenAIConfig,
    // RAG
    ragConfig, setRAGConfig,
    // Graph
    graphConfig, setGraphConfig,
    // Actions
    handleSave,
  };
};

export default useAISettings;
