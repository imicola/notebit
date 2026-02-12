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
  TestOpenAIConnection
} from '../../wailsjs/go/main/App';
import { RefreshCw, CheckCircle, AlertCircle, Save, Server, Cpu, Layers } from 'lucide-react';

export default function AISettings() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState(null);
  const [testingOpenAI, setTestingOpenAI] = useState(false);
  const [openaiTestResult, setOpenaiTestResult] = useState(null);
  
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

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    setLoading(true);
    try {
      // Load all configs in parallel
      const [aiStatus, openai, ollama, chunking] = await Promise.all([
        GetAIStatus(),
        GetOpenAIConfig(),
        GetOllamaConfig(),
        GetChunkingConfig()
      ]);

      setStatus(aiStatus);
      setProvider(aiStatus.current_provider || 'ollama');
      setOpenaiConfig(openai);
      setOllamaConfig(ollama);
      setChunkingConfig(chunking);
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
    <div className="space-y-8 pb-8">
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
