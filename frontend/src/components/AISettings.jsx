import { useState } from 'react';
import { RefreshCw, CheckCircle, AlertCircle, Save, Server, MessageSquare, Sparkles, Network } from 'lucide-react';
import useAISettings from '../hooks/useAISettings';
import EmbeddingTab from './AISettings/EmbeddingTab';
import LLMTab from './AISettings/LLMTab';
import RAGTab from './AISettings/RAGTab';
import GraphTab from './AISettings/GraphTab';

export default function AISettings() {
  const [activeTab, setActiveTab] = useState('embedding');

  const {
    loading, saving, status,
    similarityStatus, vectorEngine, setVectorEngine, availableVectorEngines, reindexing, reindexResult,
    provider, setProvider,
    openaiConfig, setOpenaiConfig,
    testingOpenAI, openaiTestResult, handleTestOpenAI,
    handleReindexEmbeddings,
    ollamaConfig, setOllamaConfig,
    chunkingConfig, setChunkingConfig,
    llmConfig, setLLMConfig,
    llmOpenAIConfig, setLLMOpenAIConfig,
    embeddingProfiles,
    llmProfiles,
    saveEmbeddingProfile,
    applyEmbeddingProfile,
    deleteEmbeddingProfile,
    saveLLMProfile,
    applyLLMProfile,
    deleteLLMProfile,
    ragConfig, setRAGConfig,
    graphConfig, setGraphConfig,
    handleSave,
  } = useAISettings();

  if (loading) {
    return <div className="flex justify-center p-8"><RefreshCw className="animate-spin text-muted" /></div>;
  }

  const tabs = [
    { id: 'embedding', label: 'Embedding', icon: Server },
    { id: 'llm', label: 'LLM Chat', icon: MessageSquare },
    { id: 'rag', label: 'RAG', icon: Sparkles },
    { id: 'graph', label: 'Graph', icon: Network },
  ];

  return (
    <div className="space-y-6 pb-8">
      {/* Tab Navigation */}
      <div className="flex gap-2 border-b border-modifier-border">
        {tabs.map(tab => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${
                activeTab === tab.id
                  ? 'border-obsidian-purple text-normal'
                  : 'border-transparent text-muted hover:text-normal'
              }`}
            >
              <Icon size={16} className="mr-2" />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      {activeTab === 'embedding' && (
        <EmbeddingTab
          provider={provider} setProvider={setProvider}
          ollamaConfig={ollamaConfig} setOllamaConfig={setOllamaConfig}
          openaiConfig={openaiConfig} setOpenaiConfig={setOpenaiConfig}
          chunkingConfig={chunkingConfig} setChunkingConfig={setChunkingConfig}
          testingOpenAI={testingOpenAI} openaiTestResult={openaiTestResult}
          handleTestOpenAI={handleTestOpenAI}
          embeddingProfiles={embeddingProfiles}
          saveEmbeddingProfile={saveEmbeddingProfile}
          applyEmbeddingProfile={applyEmbeddingProfile}
          deleteEmbeddingProfile={deleteEmbeddingProfile}
          similarityStatus={similarityStatus}
          vectorEngine={vectorEngine}
          setVectorEngine={setVectorEngine}
          availableVectorEngines={availableVectorEngines}
          reindexing={reindexing}
          reindexResult={reindexResult}
          handleReindexEmbeddings={handleReindexEmbeddings}
        />
      )}

      {activeTab === 'llm' && (
        <LLMTab
          llmConfig={llmConfig} setLLMConfig={setLLMConfig}
          llmOpenAIConfig={llmOpenAIConfig} setLLMOpenAIConfig={setLLMOpenAIConfig}
          llmProfiles={llmProfiles}
          saveLLMProfile={saveLLMProfile}
          applyLLMProfile={applyLLMProfile}
          deleteLLMProfile={deleteLLMProfile}
        />
      )}

      {activeTab === 'rag' && (
        <RAGTab ragConfig={ragConfig} setRAGConfig={setRAGConfig} />
      )}

      {activeTab === 'graph' && (
        <GraphTab graphConfig={graphConfig} setGraphConfig={setGraphConfig} />
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
