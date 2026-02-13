import { MessageSquare } from 'lucide-react';

/**
 * LLM Chat configuration tab for AISettings
 */
export default function LLMTab({
  llmConfig, setLLMConfig,
  llmOpenAIConfig, setLLMOpenAIConfig
}) {
  return (
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
  );
}
