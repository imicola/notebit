import { useState, useEffect } from 'react';
import {
  GetOpenAIConfig,
  GetOllamaConfig,
  GetChunkingConfig,
  GetAIStatus,
  SetAIProvider,
  SetOpenAIConfig,
  SetOllamaConfig,
  SetChunkingConfig,
  TestOpenAIConnection,
  GetLLMConfig,
  SetLLMConfig,
  GetRAGConfig,
  SetRAGConfig,
  GetGraphConfig,
  SetGraphConfig
} from '../../wailsjs/go/main/App';
import { RefreshCw, CheckCircle, AlertCircle, Save, Server, Cpu, Layers, MessageSquare, Network, Sparkles } from 'lucide-react';

export default function AISettings() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState(null);
  const [testingOpenAI, setTestingOpenAI] = useState(false);
  const [openaiTestResult, setOpenaiTestResult] = useState(null);
  const [activeTab, setActiveTab] = useState('embedding'); // 'embedding' | 'llm' | 'rag' | 'graph'

  // Config States
  const [provider, setProvider] = useState('ollama');
  const [openaiConfig, setOpenaiConfig] = useState({
    api_key: '',
    base_url: '',
    organization: '',
    embedding_model: ''
  });
  const [ollamaConfig, setOllamaConfig] = useState({
    base_url: '',
    embedding_model: '',
    timeout: 30
  });
  const [chunkingConfig, setChunkingConfig] = useState({
    strategy: 'heading',
    chunk_size: 1000,
    chunk_overlap: 200,
    min_chunk_size: 100,
    max_chunk_size: 4000,
    preserve_heading: true,
    heading_separator: '\n\n'
  });

  // LLM Config (for chat)
  const [llmConfig, setLLMConfig] = useState({
    provider: 'openai',
    model: 'gpt-4o-mini',
    temperature: 0.7,
    max_tokens: 2000
  });

  const [llmOpenAIConfig, setLLMOpenAIConfig] = useState({
    api_key: '',
    base_url: '',
    organization: ''
  });

  // RAG Config
  const [ragConfig, setRAGConfigState] = useState({
    max_context_chunks: 5,
    temperature: 0.7
  });

  // Graph Config
  const [graphConfig, setGraphConfigState] = useState({
    min_similarity_threshold: 0.75,
    max_nodes: 100,
    show_implicit_links: true
  });

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    setLoading(true);
    try {
      // Load all configs in parallel
      const [aiStatus, openai, ollama, chunking, llm, rag, graph] = await Promise.all([
        GetAIStatus(),
        GetOpenAIConfig(),
        GetOllamaConfig(),
        GetChunkingConfig(),
        GetLLMConfig(),
        GetRAGConfig(),
        GetGraphConfig()
      ]);

      setStatus(aiStatus);
      setProvider(aiStatus.current_provider || 'ollama');
      setOpenaiConfig(openai);
      setOllamaConfig(ollama);
      setChunkingConfig(chunking);
      
      setLLMConfig({
        provider: llm.provider,
        model: llm.model,
        temperature: llm.temperature,
        max_tokens: llm.max_tokens
      });
      
      if (llm.openai) {
        setLLMOpenAIConfig({
          api_key: llm.openai.api_key || '',
          base_url: llm.openai.base_url || '',
          organization: llm.openai.organization || ''
        });
      }

      setRAGConfigState(rag);
      setGraphConfigState(graph);
    } catch (error) {
      console.error('Failed to load settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      // Save based on current provider
      if (provider === 'openai') {
        await SetOpenAIConfig(openaiConfig.api_key, openaiConfig.base_url, openaiConfig.organization);
      } else {
        await SetOllamaConfig(ollamaConfig.base_url, ollamaConfig.embedding_model, parseInt(ollamaConfig.timeout));
      }

      // Save provider
      await SetAIProvider(provider);

      // Save chunking config
      await SetChunkingConfig(
        chunkingConfig.strategy,
        parseInt(chunkingConfig.chunk_size),
        parseInt(chunkingConfig.chunk_overlap),
        parseInt(chunkingConfig.min_chunk_size),
        parseInt(chunkingConfig.max_chunk_size),
        chunkingConfig.preserve_heading,
        chunkingConfig.heading_separator
      );

      // Save LLM config
      await SetLLMConfig(
        llmConfig.provider,
        llmConfig.model,
        llmConfig.temperature,
        llmConfig.max_tokens,
        llmOpenAIConfig.api_key,
        llmOpenAIConfig.base_url,
        llmOpenAIConfig.organization
      );

      // Save RAG config
      await SetRAGConfig(
        ragConfig.max_context_chunks,
        ragConfig.temperature
      );

      // Save Graph config
      await SetGraphConfig(
        graphConfig.min_similarity_threshold,
        graphConfig.max_nodes,
        graphConfig.show_implicit_links
      );

      // Reload status to verify
      const newStatus = await GetAIStatus();
      setStatus(newStatus);

    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleTestOpenAI = async () => {
    setTestingOpenAI(true);
    setOpenaiTestResult(null);
    try {
      const result = await TestOpenAIConnection(
        openaiConfig.api_key,
        openaiConfig.base_url,
        openaiConfig.organization,
        openaiConfig.embedding_model
      );
      setOpenaiTestResult({ ok: true, message: `Connected (${result.model || 'ok'}, ${result.dimension || 0}d)` });
    } catch (error) {
      setOpenaiTestResult({ ok: false, message: error?.message || 'Connection failed' });
    } finally {
      setTestingOpenAI(false);
    }
  };

  if (loading) {
    return <div className="flex justify-center p-8"><RefreshCw className="animate-spin text-muted" /></div>;
  }

  return (
    <div className="space-y-6 pb-8">
      {/* Tab Navigation */}
      <div className="flex gap-2 border-b border-modifier-border">
        <button
          onClick={() => setActiveTab('embedding')}
          className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${
            activeTab === 'embedding'
              ? 'border-obsidian-purple text-normal'
              : 'border-transparent text-muted hover:text-normal'
          }`}
        >
          <Server size={16} className="mr-2" />
          Embedding
        </button>
        <button
          onClick={() => setActiveTab('llm')}
          className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${
            activeTab === 'llm'
              ? 'border-obsidian-purple text-normal'
              : 'border-transparent text-muted hover:text-normal'
          }`}
        >
          <MessageSquare size={16} className="mr-2" />
          LLM Chat
        </button>
        <button
          onClick={() => setActiveTab('rag')}
          className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${
            activeTab === 'rag'
              ? 'border-obsidian-purple text-normal'
              : 'border-transparent text-muted hover:text-normal'
          }`}
        >
          <Sparkles size={16} className="mr-2" />
          RAG
        </button>
        <button
          onClick={() => setActiveTab('graph')}
          className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${
            activeTab === 'graph'
              ? 'border-obsidian-purple text-normal'
              : 'border-transparent text-muted hover:text-normal'
          }`}
        >
          <Network size={16} className="mr-2" />
          Graph
        </button>
      </div>

      {/* Tab Content */}
      {activeTab === 'embedding' && (
        <div className="space-y-6">
          {/* Provider Selection */}
          <section>
            <div className="flex items-center gap-2 mb-4">
              <Server className="text-obsidian-purple" size={20} />
              <h3 className="text-lg font-medium text-normal">AI Provider</h3>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <button
                onClick={() => setProvider('ollama')}
                className={`p-4 rounded-lg border text-left transition-all ${
                  provider === 'ollama'
                    ? 'border-obsidian-purple bg-obsidian-purple/10 ring-1 ring-obsidian-purple'
                    : 'border-modifier-border hover:border-obsidian-purple/50 bg-primary-alt'
                }`}
              >
                <div className="font-medium text-normal mb-1">Ollama (Local)</div>
                <div className="text-xs text-muted">Run models locally. Private and free.</div>
              </button>

              <button
                onClick={() => setProvider('openai')}
                className={`p-4 rounded-lg border text-left transition-all ${
                  provider === 'openai'
                    ? 'border-obsidian-purple bg-obsidian-purple/10 ring-1 ring-obsidian-purple'
                    : 'border-modifier-border hover:border-obsidian-purple/50 bg-primary-alt'
                }`}
              >
                <div className="font-medium text-normal mb-1">OpenAI</div>
                <div className="text-xs text-muted">Cloud-based models. Requires API key.</div>
              </button>
            </div>
          </section>

          {/* Provider Settings */}
          <section className="space-y-4">
            <div className="flex items-center gap-2 mb-4">
              <Cpu className="text-obsidian-purple" size={20} />
              <h3 className="text-lg font-medium text-normal">
                {provider === 'ollama' ? 'Ollama Configuration' : 'OpenAI Configuration'}
              </h3>
            </div>

            {provider === 'ollama' ? (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">Base URL</label>
                  <input
                    type="text"
                    value={ollamaConfig.base_url}
                    onChange={(e) => setOllamaConfig({...ollamaConfig, base_url: e.target.value})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                    placeholder="http://localhost:11434"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-normal mb-1">Embedding Model</label>
                    <input
                      type="text"
                      value={ollamaConfig.embedding_model}
                      onChange={(e) => setOllamaConfig({...ollamaConfig, embedding_model: e.target.value})}
                      className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                      placeholder="nomic-embed-text"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-normal mb-1">Timeout (seconds)</label>
                    <input
                      type="number"
                      value={ollamaConfig.timeout}
                      onChange={(e) => setOllamaConfig({...ollamaConfig, timeout: e.target.value})}
                      className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                    />
                  </div>
                </div>
              </div>
            ) : (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">API Key</label>
                  <input
                    type="password"
                    value={openaiConfig.api_key}
                    onChange={(e) => setOpenaiConfig({...openaiConfig, api_key: e.target.value})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                    placeholder="sk-..."
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">Embedding Model</label>
                  <input
                    type="text"
                    value={openaiConfig.embedding_model}
                    onChange={(e) => setOpenaiConfig({...openaiConfig, embedding_model: e.target.value})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                    placeholder="text-embedding-3-small"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-normal mb-1">Base URL (Optional)</label>
                    <input
                      type="text"
                      value={openaiConfig.base_url}
                      onChange={(e) => setOpenaiConfig({...openaiConfig, base_url: e.target.value})}
                      className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                      placeholder="https://api.openai.com/v1"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-normal mb-1">Organization ID (Optional)</label>
                    <input
                      type="text"
                      value={openaiConfig.organization}
                      onChange={(e) => setOpenaiConfig({...openaiConfig, organization: e.target.value})}
                      className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                    />
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <button
                    onClick={handleTestOpenAI}
                    disabled={testingOpenAI || !openaiConfig.api_key}
                    className="flex items-center gap-2 px-3 py-2 rounded-md border border-modifier-border bg-primary-alt text-sm text-normal hover:border-obsidian-purple/60 transition-colors disabled:opacity-50"
                  >
                    {testingOpenAI ? <RefreshCw className="animate-spin" size={14} /> : <CheckCircle size={14} />}
                    {testingOpenAI ? 'Testing...' : 'Test OpenAI API'}
                  </button>
                  {openaiTestResult ? (
                    openaiTestResult.ok ? (
                      <div className="flex items-center gap-1.5 text-green-500 text-sm">
                        <CheckCircle size={16} />
                        <span>{openaiTestResult.message}</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1.5 text-orange-500 text-sm">
                        <AlertCircle size={16} />
                        <span>{openaiTestResult.message}</span>
                      </div>
                    )
                  ) : null}
                </div>
              </div>
            )}
          </section>

          {/* Chunking Settings */}
          <section>
            <div className="flex items-center gap-2 mb-4">
              <Layers className="text-obsidian-purple" size={20} />
              <h3 className="text-lg font-medium text-normal">Chunking Strategy</h3>
            </div>

            <div className="space-y-4 bg-primary-alt/30 p-4 rounded-lg border border-modifier-border">
              <div>
                <label className="block text-sm font-medium text-normal mb-1">Strategy</label>
                <select
                  value={chunkingConfig.strategy}
                  onChange={(e) => setChunkingConfig({...chunkingConfig, strategy: e.target.value})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                >
                  <option value="heading">Heading-based (Recommended)</option>
                  <option value="fixed">Fixed Size</option>
                  <option value="sliding">Sliding Window</option>
                </select>
                <p className="text-xs text-muted mt-1">
                  {chunkingConfig.strategy === 'heading'
                    ? 'Splits content by markdown headers (#, ##, etc.) to preserve semantic context.'
                    : 'Splits content into fixed-size blocks.'}
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">Chunk Size</label>
                  <input
                    type="number"
                    value={chunkingConfig.chunk_size}
                    onChange={(e) => setChunkingConfig({...chunkingConfig, chunk_size: e.target.value})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">Overlap</label>
                  <input
                    type="number"
                    value={chunkingConfig.chunk_overlap}
                    onChange={(e) => setChunkingConfig({...chunkingConfig, chunk_overlap: e.target.value})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                  />
                </div>
              </div>
            </div>
          </section>
        </div>
      )}

      {activeTab === 'llm' && (
        <div className="space-y-6">
          <section>
            <div className="flex items-center gap-2 mb-4">
              <MessageSquare className="text-obsidian-purple" size={20} />
              <h3 className="text-lg font-medium text-normal">LLM Configuration (Chat)</h3>
            </div>

            <div className="space-y-4 bg-primary-alt/30 p-4 rounded-lg border border-modifier-border">
              <div>
                <label className="block text-sm font-medium text-normal mb-1">Provider</label>
                <select
                  value={llmConfig.provider}
                  onChange={(e) => setLLMConfig({...llmConfig, provider: e.target.value})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                >
                  <option value="openai">OpenAI</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-normal mb-1">Model</label>
                <input
                  type="text"
                  value={llmConfig.model}
                  onChange={(e) => setLLMConfig({...llmConfig, model: e.target.value})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                  placeholder="gpt-4o-mini"
                />
                <p className="text-xs text-muted mt-1">
                  Available models: gpt-4o, gpt-4o-mini, gpt-3.5-turbo
                </p>
              </div>

              {llmConfig.provider === 'openai' && (
                <div className="space-y-4 border-t border-modifier-border pt-4 mt-4">
                  <h4 className="text-sm font-medium text-normal">OpenAI Settings (Chat Specific)</h4>
                  <div>
                    <label className="block text-sm font-medium text-normal mb-1">API Key</label>
                    <input
                      type="password"
                      value={llmOpenAIConfig.api_key}
                      onChange={(e) => setLLMOpenAIConfig({...llmOpenAIConfig, api_key: e.target.value})}
                      className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                      placeholder="Leave empty to use global AI settings"
                    />
                    <p className="text-xs text-muted mt-1">Overrides the global AI provider key if set.</p>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-normal mb-1">Base URL</label>
                      <input
                        type="text"
                        value={llmOpenAIConfig.base_url}
                        onChange={(e) => setLLMOpenAIConfig({...llmOpenAIConfig, base_url: e.target.value})}
                        className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                        placeholder="https://api.openai.com/v1"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-normal mb-1">Organization ID</label>
                      <input
                        type="text"
                        value={llmOpenAIConfig.organization}
                        onChange={(e) => setLLMOpenAIConfig({...llmOpenAIConfig, organization: e.target.value})}
                        className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                      />
                    </div>
                  </div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">Temperature</label>
                  <input
                    type="number"
                    step="0.1"
                    min="0"
                    max="2"
                    value={llmConfig.temperature}
                    onChange={(e) => setLLMConfig({...llmConfig, temperature: parseFloat(e.target.value)})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                  />
                  <p className="text-xs text-muted mt-1">Lower = more focused, Higher = more creative</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-normal mb-1">Max Tokens</label>
                  <input
                    type="number"
                    value={llmConfig.max_tokens}
                    onChange={(e) => setLLMConfig({...llmConfig, max_tokens: parseInt(e.target.value)})}
                    className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                  />
                  <p className="text-xs text-muted mt-1">Maximum response length</p>
                </div>
              </div>
            </div>
          </section>
        </div>
      )}

      {activeTab === 'rag' && (
        <div className="space-y-6">
          <section>
            <div className="flex items-center gap-2 mb-4">
              <Sparkles className="text-obsidian-purple" size={20} />
              <h3 className="text-lg font-medium text-normal">RAG Configuration</h3>
            </div>

            <div className="space-y-4 bg-primary-alt/30 p-4 rounded-lg border border-modifier-border">
              <div>
                <label className="block text-sm font-medium text-normal mb-1">Max Context Chunks</label>
                <input
                  type="number"
                  min="1"
                  max="20"
                  value={ragConfig.max_context_chunks}
                  onChange={(e) => setRAGConfigState({...ragConfig, max_context_chunks: parseInt(e.target.value)})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                />
                <p className="text-xs text-muted mt-1">
                  Maximum number of document chunks to include as context for each query. More chunks = more context but slower responses.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-normal mb-1">Temperature</label>
                <input
                  type="number"
                  step="0.1"
                  min="0"
                  max="2"
                  value={ragConfig.temperature}
                  onChange={(e) => setRAGConfigState({...ragConfig, temperature: parseFloat(e.target.value)})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                />
                <p className="text-xs text-muted mt-1">Controls response randomness</p>
              </div>
            </div>
          </section>
        </div>
      )}

      {activeTab === 'graph' && (
        <div className="space-y-6">
          <section>
            <div className="flex items-center gap-2 mb-4">
              <Network className="text-obsidian-purple" size={20} />
              <h3 className="text-lg font-medium text-normal">Knowledge Graph Configuration</h3>
            </div>

            <div className="space-y-4 bg-primary-alt/30 p-4 rounded-lg border border-modifier-border">
              <div>
                <label className="block text-sm font-medium text-normal mb-1">Similarity Threshold</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  max="1"
                  value={graphConfig.min_similarity_threshold}
                  onChange={(e) => setGraphConfigState({...graphConfig, min_similarity_threshold: parseFloat(e.target.value)})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                />
                <p className="text-xs text-muted mt-1">
                  Minimum similarity score (0-1) for creating implicit links between semantically similar files.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-normal mb-1">Max Nodes</label>
                <input
                  type="number"
                  min="10"
                  max="500"
                  value={graphConfig.max_nodes}
                  onChange={(e) => setGraphConfigState({...graphConfig, max_nodes: parseInt(e.target.value)})}
                  className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
                />
                <p className="text-xs text-muted mt-1">
                  Maximum number of nodes to display in the graph (for performance).
                </p>
              </div>

              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="show-implicit-links"
                  checked={graphConfig.show_implicit_links}
                  onChange={(e) => setGraphConfigState({...graphConfig, show_implicit_links: e.target.checked})}
                  className="rounded border-modifier-border bg-primary-alt"
                />
                <label htmlFor="show-implicit-links" className="text-sm text-normal cursor-pointer">
                  Show implicit (semantic similarity) links
                </label>
                <p className="text-xs text-muted">
                  When enabled, shows dashed lines for semantically similar files.
                </p>
              </div>
            </div>
          </section>
        </div>
      )}

      {/* Footer / Status */}
      <div className="flex items-center justify-between pt-4 border-t border-modifier-border">
        <div className="flex items-center gap-2">
          {status?.provider_healthy ? (
            <div className="flex items-center gap-1.5 text-green-500 text-sm">
              <CheckCircle size={16} />
              <span>Service Ready</span>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-orange-500 text-sm">
              <AlertCircle size={16} />
              <span>Service Not Ready</span>
            </div>
          )}
        </div>

        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 px-4 py-2 bg-obsidian-purple hover:bg-obsidian-purple-hover text-white rounded-md font-medium transition-colors disabled:opacity-50"
        >
          {saving ? <RefreshCw className="animate-spin" size={16} /> : <Save size={16} />}
          {saving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  );
}
