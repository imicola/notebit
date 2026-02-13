import { Sparkles } from 'lucide-react';

/**
 * RAG configuration tab for AISettings
 */
export default function RAGTab({ ragConfig, setRAGConfig }) {
  return (
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
              onChange={(e) => setRAGConfig({...ragConfig, max_context_chunks: parseInt(e.target.value)})}
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
              onChange={(e) => setRAGConfig({...ragConfig, temperature: parseFloat(e.target.value)})}
              className="w-full rounded-md border border-modifier-border bg-primary-alt px-3 py-2 text-sm text-normal focus:border-obsidian-purple focus:outline-none"
            />
            <p className="text-xs text-muted mt-1">Controls response randomness</p>
          </div>
        </div>
      </section>
    </div>
  );
}
