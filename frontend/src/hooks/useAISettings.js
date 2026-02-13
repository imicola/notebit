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

  // Config States
  const [provider, setProvider] = useState('ollama');
  const [openaiConfig, setOpenaiConfig] = useState({
    api_key: '', base_url: '', organization: '', embedding_model: ''
  });
  const [ollamaConfig, setOllamaConfig] = useState({
    base_url: '', embedding_model: '', timeout: 30
  });
  const [chunkingConfig, setChunkingConfig] = useState({
    strategy: 'heading', chunk_size: 1000, chunk_overlap: 200,
    min_chunk_size: 100, max_chunk_size: 4000,
    preserve_heading: true, heading_separator: '\n\n'
  });
  const [llmConfig, setLLMConfig] = useState({
    provider: 'openai', model: 'gpt-4o-mini', temperature: 0.7, max_tokens: 2000
  });
  const [llmOpenAIConfig, setLLMOpenAIConfig] = useState({
    api_key: '', base_url: '', organization: ''
  });
  const [ragConfig, setRAGConfig] = useState({
    max_context_chunks: 5, temperature: 0.7
  });
  const [graphConfig, setGraphConfig] = useState({
    min_similarity_threshold: 0.75, max_nodes: 100, show_implicit_links: true
  });

  // Load all settings on mount
  const loadSettings = useCallback(async () => {
    setLoading(true);
    try {
      const [aiStatus, openai, ollama, chunking, llm, rag, graph] = await Promise.all([
        aiService.getStatus(),
        aiService.getOpenAIConfig(),
        aiService.getOllamaConfig(),
        aiService.getChunkingConfig(),
        aiService.getLLMConfig(),
        aiService.getRAGConfig(),
        aiService.getGraphConfig()
      ]);

      setStatus(aiStatus);
      setProvider(aiStatus.current_provider || 'ollama');
      setOpenaiConfig(openai);
      setOllamaConfig(ollama);
      setChunkingConfig(chunking);

      setLLMConfig({
        provider: llm.provider, model: llm.model,
        temperature: llm.temperature, max_tokens: llm.max_tokens
      });

      if (llm.openai) {
        setLLMOpenAIConfig({
          api_key: llm.openai.api_key || '',
          base_url: llm.openai.base_url || '',
          organization: llm.openai.organization || ''
        });
      }

      setRAGConfig(rag);
      setGraphConfig(graph);
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
        await aiService.setOpenAIConfig(openaiConfig.api_key, openaiConfig.base_url, openaiConfig.organization);
      } else {
        await aiService.setOllamaConfig(ollamaConfig.base_url, ollamaConfig.embedding_model, parseInt(ollamaConfig.timeout));
      }

      await aiService.setProvider(provider);

      await aiService.setChunkingConfig(
        chunkingConfig.strategy,
        parseInt(chunkingConfig.chunk_size),
        parseInt(chunkingConfig.chunk_overlap),
        parseInt(chunkingConfig.min_chunk_size),
        parseInt(chunkingConfig.max_chunk_size),
        chunkingConfig.preserve_heading,
        chunkingConfig.heading_separator
      );

      await aiService.setLLMConfig(
        llmConfig.provider, llmConfig.model, llmConfig.temperature, llmConfig.max_tokens,
        llmOpenAIConfig.api_key, llmOpenAIConfig.base_url, llmOpenAIConfig.organization
      );

      await aiService.setRAGConfig(ragConfig.max_context_chunks, ragConfig.temperature);
      await aiService.setGraphConfig(graphConfig.min_similarity_threshold, graphConfig.max_nodes, graphConfig.show_implicit_links);

      const newStatus = await aiService.getStatus();
      setStatus(newStatus);
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  }, [provider, openaiConfig, ollamaConfig, chunkingConfig, llmConfig, llmOpenAIConfig, ragConfig, graphConfig]);

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

  return {
    // Loading/saving state
    loading, saving, status,
    // Provider
    provider, setProvider,
    // OpenAI
    openaiConfig, setOpenaiConfig,
    testingOpenAI, openaiTestResult, handleTestOpenAI,
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
