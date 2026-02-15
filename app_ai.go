package main

import (
	"fmt"
	"notebit/pkg/ai"
	"notebit/pkg/config"
	"time"
)

// ============ AI SERVICE API METHODS ============

// GetOpenAIConfig returns the OpenAI configuration
func (a *App) GetOpenAIConfig() (config.OpenAIConfig, error) {
	return a.cfg.GetOpenAIConfig(), nil
}

// GetOllamaConfig returns the Ollama configuration
func (a *App) GetOllamaConfig() (config.OllamaConfig, error) {
	return a.cfg.GetOllamaConfig(), nil
}

// GetChunkingConfig returns the chunking configuration
func (a *App) GetChunkingConfig() (config.ChunkingConfig, error) {
	return a.cfg.GetChunkingConfig(), nil
}

// SetChunkingConfig sets the chunking configuration
func (a *App) SetChunkingConfig(strategy string, chunkSize, chunkOverlap, minChunkSize, maxChunkSize int, preserveHeading bool, headingSeparator string) error {
	cfg := config.ChunkingConfig{
		Strategy:         strategy,
		ChunkSize:        chunkSize,
		ChunkOverlap:     chunkOverlap,
		MinChunkSize:     minChunkSize,
		MaxChunkSize:     maxChunkSize,
		PreserveHeading:  preserveHeading,
		HeadingSeparator: headingSeparator,
	}
	a.cfg.SetChunkingConfig(cfg)

	if err := a.ai.Reconfigure(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// GetRAGConfig returns the RAG configuration
func (a *App) GetRAGConfig() (config.RAGConfig, error) {
	return a.cfg.GetRAGConfig(), nil
}

// SetRAGConfig sets the RAG configuration
func (a *App) SetRAGConfig(maxContextChunks int, temperature float32, systemPrompt string) error {
	cfg := config.RAGConfig{
		MaxContextChunks: maxContextChunks,
		Temperature:      temperature,
		SystemPrompt:     systemPrompt,
	}
	a.cfg.SetRAGConfig(cfg)
	return a.cfg.Save()
}

// GetAIStatus returns the current status of the AI service
func (a *App) GetAIStatus() (map[string]interface{}, error) {
	status, err := a.ai.GetStatus()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"current_provider":     status.CurrentProvider,
		"available_providers":  status.AvailableProviders,
		"current_model":        status.CurrentModel,
		"chunking_strategy":    status.ChunkingStrategy,
		"available_strategies": status.AvailableStrategies,
		"model_dimension":      status.ModelDimension,
		"provider_healthy":     status.ProviderHealthy,
	}, nil
}

// SetAIProvider sets the current AI provider
func (a *App) SetAIProvider(provider string) error {
	if err := a.ai.SetProvider(provider); err != nil {
		return err
	}
	return a.cfg.Save()
}

// SetAIModel sets the default embedding model
func (a *App) SetAIModel(model string) error {
	a.cfg.SetEmbeddingModel(model)
	if err := a.ai.Reconfigure(); err != nil {
		return err
	}
	return a.cfg.Save()
}

// SetOpenAIConfig sets the OpenAI configuration
func (a *App) SetOpenAIConfig(apiKey, baseURL, organization, embeddingModel string) error {
	if err := a.ai.SetOpenAIConfig(apiKey, baseURL, organization, embeddingModel); err != nil {
		return err
	}
	return a.cfg.Save()
}

// SetOllamaConfig sets the Ollama configuration
func (a *App) SetOllamaConfig(baseURL, model string, timeout int) error {
	if err := a.ai.SetOllamaConfig(baseURL, model, timeout); err != nil {
		return err
	}
	return a.cfg.Save()
}

// TestOpenAIConnection tests an OpenAI connection with given credentials
func (a *App) TestOpenAIConnection(apiKey, baseURL, organization, model string) (map[string]interface{}, error) {
	provider, err := ai.NewOpenAIProvider(ai.OpenAIConfig{
		APIKey:       apiKey,
		BaseURL:      baseURL,
		Organization: organization,
		Timeout:      15 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if model == "" {
		model = provider.GetDefaultModel()
	}

	resp, err := provider.GenerateEmbedding(&ai.EmbeddingRequest{
		Text:  "ping",
		Model: model,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"model":     resp.Model,
		"dimension": len(resp.Embedding),
	}, nil
}

// GenerateEmbedding generates an embedding for a single text
func (a *App) GenerateEmbedding(text string) ([]float32, error) {
	resp, err := a.ai.GenerateEmbedding(text)
	if err != nil {
		return nil, err
	}
	return resp.Embedding, nil
}

// GenerateEmbeddingsBatch generates embeddings for multiple texts
func (a *App) GenerateEmbeddingsBatch(texts []string) ([][]float32, error) {
	responses, err := a.ai.GenerateEmbeddingsBatch(texts)
	if err != nil {
		return nil, err
	}

	result := make([][]float32, len(responses))
	for i, resp := range responses {
		if resp != nil {
			result[i] = resp.Embedding
		}
	}
	return result, nil
}

// ChunkText splits text using the configured chunking strategy
func (a *App) ChunkText(text string) ([]map[string]interface{}, error) {
	chunks, err := a.ai.ChunkText(text)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		result[i] = map[string]interface{}{
			"content": chunk.Content,
			"heading": chunk.Heading,
			"index":   chunk.Index,
		}
	}
	return result, nil
}

// ProcessDocument chunks text and generates embeddings for all chunks
func (a *App) ProcessDocument(text string) ([]map[string]interface{}, error) {
	chunks, err := a.ai.ProcessDocument(text)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(chunks))
	for i, chunk := range chunks {
		result[i] = map[string]interface{}{
			"content":         chunk.Content,
			"heading":         chunk.Heading,
			"index":           chunk.Index,
			"embedding":       chunk.Embedding,
			"embedding_model": chunk.ModelName,
		}
	}
	return result, nil
}

// IndexFileWithEmbedding indexes a file and generates embeddings for its chunks
func (a *App) IndexFileWithEmbedding(path string) error {
	return a.ks.IndexFileWithEmbedding(path)
}

// ReindexAllWithEmbeddings reindexes all files with embeddings
func (a *App) ReindexAllWithEmbeddings() (map[string]interface{}, error) {
	if a.ks == nil {
		return nil, fmt.Errorf("knowledge service not initialized - please open a folder first")
	}
	return a.ks.ReindexAllWithEmbeddings()
}

// ============ LLM CONFIG API METHODS ============

// GetLLMConfig returns the LLM configuration
func (a *App) GetLLMConfig() (config.LLMConfig, error) {
	return a.cfg.GetLLMConfig(), nil
}

// SetLLMConfig sets the LLM configuration
func (a *App) SetLLMConfig(provider string, model string, temperature float32, maxTokens int, apiKey, baseURL, organization string) error {
	// Get existing config to preserve other fields if needed
	currentConfig := a.cfg.GetLLMConfig()

	llmConfig := config.LLMConfig{
		Provider:    provider,
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		OpenAI:      currentConfig.OpenAI,
		Ollama:      currentConfig.Ollama,
	}

	if apiKey != "" {
		llmConfig.OpenAI.APIKey = apiKey
	}
	if baseURL != "" {
		llmConfig.OpenAI.BaseURL = baseURL
	}
	if organization != "" {
		llmConfig.OpenAI.Organization = organization
	}

	a.cfg.SetLLMConfig(llmConfig)

	// Reinitialize LLM provider
	a.initializeLLM()

	// Reinitialize RAG service if needed
	if a.rag != nil {
		a.initializeRAG()
	}

	return a.cfg.Save()
}
