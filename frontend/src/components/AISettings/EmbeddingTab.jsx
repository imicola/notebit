import { Server, Cpu, Layers, RefreshCw, CheckCircle, AlertCircle } from 'lucide-react';

/**
 * Embedding configuration tab for AISettings
 */
export default function EmbeddingTab({
  provider, setProvider,
  ollamaConfig, setOllamaConfig,
  openaiConfig, setOpenaiConfig,
  chunkingConfig, setChunkingConfig,
  testingOpenAI, openaiTestResult, handleTestOpenAI
}) {
  return (
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
  );
}
