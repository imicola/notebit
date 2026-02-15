package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"notebit/pkg/config"
	"notebit/pkg/logger"
)

// Service manages AI operations including embedding and chunking
type Service struct {
	mu              sync.RWMutex
	cfg             *config.Config
	providers       map[string]EmbeddingProvider
	chunkers        map[string]ChunkingStrategy
	currentProvider string
}

// NewService creates a new AI service
func NewService(cfg *config.Config) *Service {
	if cfg == nil {
		cfg = config.Get()
	}

	s := &Service{
		cfg:       cfg,
		providers: make(map[string]EmbeddingProvider),
		chunkers:  make(map[string]ChunkingStrategy),
	}

	// Initialize current provider
	s.currentProvider = cfg.GetProvider()

	return s
}

// Initialize sets up the AI service with providers based on configuration
func (s *Service) Initialize() error {
	timer := logger.StartTimer()
	logger.Info("Initializing AI service")

	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentProvider = s.cfg.GetProvider()

	// Initialize OpenAI provider if configured
	if s.cfg.IsOpenAIConfigured() {
		openaiCfg := s.cfg.GetOpenAIConfig()
		provider, err := NewOpenAIProvider(OpenAIConfig{
			APIKey:         openaiCfg.APIKey,
			BaseURL:        openaiCfg.BaseURL,
			Organization:   openaiCfg.Organization,
			Timeout:        30 * time.Second,
			EmbeddingModel: openaiCfg.EmbeddingModel,
		})
		if err == nil {
			s.providers["openai"] = provider
			logger.Debug("OpenAI provider initialized")
		} else {
			logger.WarnWithFields(context.TODO(), map[string]interface{}{"error": err.Error()}, "Failed to initialize OpenAI provider")
		}
	}

	// Initialize Ollama provider
	ollamaCfg := s.cfg.GetOllamaConfig()
	provider, err := NewOllamaProvider(OllamaConfig{
		BaseURL: ollamaCfg.BaseURL,
		Model:   ollamaCfg.EmbeddingModel,
		Timeout: time.Duration(ollamaCfg.Timeout) * time.Second,
	})
	if err == nil {
		s.providers["ollama"] = provider
		logger.DebugWithFields(context.TODO(), map[string]interface{}{
			"base_url": ollamaCfg.BaseURL,
			"model":    ollamaCfg.EmbeddingModel,
		}, "Ollama provider initialized")
	} else {
		logger.WarnWithFields(context.TODO(), map[string]interface{}{"error": err.Error()}, "Failed to initialize Ollama provider")
	}

	// Initialize chunkers
	chunkCfg := s.cfg.GetChunkingConfig()

	s.chunkers["fixed"] = NewFixedSizeChunker(
		chunkCfg.ChunkSize,
		chunkCfg.ChunkOverlap,
		chunkCfg.MinChunkSize,
	)

	s.chunkers["heading"] = NewHeadingChunker(
		chunkCfg.MaxChunkSize,
		chunkCfg.MinChunkSize,
		chunkCfg.PreserveHeading,
		chunkCfg.HeadingSeparator,
	)

	s.chunkers["sliding"] = NewSlidingWindowChunker(
		chunkCfg.ChunkSize,
		chunkCfg.ChunkSize/2, // Default step: half the window size
		chunkCfg.MinChunkSize,
	)

	s.chunkers["sentence"] = NewSentenceChunker(
		chunkCfg.MaxChunkSize,
		chunkCfg.MinChunkSize,
		1, // Default: 1 sentence overlap
	)

	// Validate that we have at least one provider
	if len(s.providers) == 0 {
		logger.Error("No embedding provider available")
		return fmt.Errorf("no embedding provider available - please configure OpenAI or ensure Ollama is running")
	}

	// Validate that the current provider is available
	if _, ok := s.providers[s.currentProvider]; !ok {
		for name := range s.providers {
			s.currentProvider = name
			break
		}
	}

	logger.InfoWithDuration(context.TODO(), timer(), "AI service initialized with %d providers", len(s.providers))
	return nil
}

// GetProvider returns the current embedding provider
func (s *Service) GetProvider() (EmbeddingProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, ok := s.providers[s.currentProvider]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not available", s.currentProvider)
	}

	return provider, nil
}

// SetProvider changes the current embedding provider
func (s *Service) SetProvider(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.providers[name]; !ok {
		return fmt.Errorf("provider '%s' not configured", name)
	}

	s.currentProvider = name
	s.cfg.SetProvider(name)
	return nil
}

// GetAvailableProviders returns a list of available provider names
func (s *Service) GetAvailableProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getAvailableProvidersLocked()
}

// getAvailableProvidersLocked returns available providers without acquiring lock (caller must hold lock)
func (s *Service) getAvailableProvidersLocked() []string {
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// GenerateEmbedding creates an embedding for a single text using the current provider
func (s *Service) GenerateEmbedding(text string) (*EmbeddingResponse, error) {
	provider, err := s.GetProvider()
	if err != nil {
		return nil, err
	}

	var resp *EmbeddingResponse
	err = retryWithBackoff(func() error {
		var opErr error
		resp, opErr = provider.GenerateEmbedding(&EmbeddingRequest{
			Text:  text,
			Model: s.cfg.GetEmbeddingModel(),
		})
		return opErr
	})

	return resp, err
}

// GenerateEmbeddingsBatch creates embeddings for multiple texts
func (s *Service) GenerateEmbeddingsBatch(texts []string) ([]*EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	provider, err := s.GetProvider()
	if err != nil {
		return nil, err
	}

	batchSize := s.cfg.AI.BatchSize
	if batchSize <= 0 {
		batchSize = 32
	}

	// Process in batches if needed
	if len(texts) <= batchSize {
		var resp []*EmbeddingResponse
		err = retryWithBackoff(func() error {
			var opErr error
			resp, opErr = provider.GenerateEmbeddingsBatch(texts)
			return opErr
		})
		return resp, err
	}

	// Split into multiple batches
	var allResults []*EmbeddingResponse
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		var results []*EmbeddingResponse

		err := retryWithBackoff(func() error {
			var opErr error
			results, opErr = provider.GenerateEmbeddingsBatch(batch)
			return opErr
		})

		if err != nil {
			return allResults, fmt.Errorf("batch %d-%d failed: %w", i, end, err)
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// ChunkText splits text using the configured chunking strategy
func (s *Service) ChunkText(text string) ([]TextChunk, error) {
	s.mu.RLock()
	chunkCfg := s.cfg.GetChunkingConfig()
	chunker, ok := s.chunkers[chunkCfg.Strategy]
	if !ok {
		// Fall back to heading strategy
		chunker = s.chunkers["heading"]
	}
	s.mu.RUnlock()

	return chunker.Chunk(text)
}

// ChunkTextWithStrategy splits text using a specific strategy
func (s *Service) ChunkTextWithStrategy(text, strategy string) ([]TextChunk, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chunker, ok := s.chunkers[strategy]
	if !ok {
		return nil, fmt.Errorf("unknown chunking strategy: %s", strategy)
	}

	return chunker.Chunk(text)
}

// GetAvailableStrategies returns a list of available chunking strategies
func (s *Service) GetAvailableStrategies() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getAvailableStrategiesLocked()
}

// getAvailableStrategiesLocked returns available strategies without acquiring lock (caller must hold lock)
func (s *Service) getAvailableStrategiesLocked() []string {
	strategies := make([]string, 0, len(s.chunkers))
	for name := range s.chunkers {
		strategies = append(strategies, name)
	}
	return strategies
}

// ProcessDocument chunks text and generates embeddings for all chunks
func (s *Service) ProcessDocument(text string) ([]TextChunk, error) {
	// First, chunk the text
	chunks, err := s.ChunkText(text)
	if err != nil {
		return nil, fmt.Errorf("chunking failed: %w", err)
	}

	if len(chunks) == 0 {
		return chunks, nil
	}

	// Generate embeddings for all chunks
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := s.GenerateEmbeddingsBatch(texts)
	if err != nil {
		return chunks, fmt.Errorf("embedding generation failed: %w", err)
	}

	// Attach embeddings to chunks
	for i, emb := range embeddings {
		if i < len(chunks) && emb != nil {
			chunks[i].Embedding = emb.Embedding
			chunks[i].ModelName = emb.Model
		}
	}

	return chunks, nil
}

// ValidateProvider checks if a provider is properly configured
func (s *Service) ValidateProvider(name string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, ok := s.providers[name]
	if !ok {
		return fmt.Errorf("provider '%s' not configured", name)
	}

	return provider.ValidateConfig()
}

// GetModelDimension returns the output dimension for the current model
func (s *Service) GetModelDimension() (int, error) {
	provider, err := s.GetProvider()
	if err != nil {
		return 0, err
	}

	model := s.cfg.GetEmbeddingModel()
	if model == "" {
		model = provider.GetDefaultModel()
	}

	return provider.GetModelDimension(model)
}

// Reconfigure reloads the configuration and reinitializes providers
func (s *Service) Reconfigure() error {
	return s.Initialize()
}

// SetOpenAIConfig updates the OpenAI configuration
func (s *Service) SetOpenAIConfig(apiKey, baseURL, organization, embeddingModel string) error {
	s.cfg.SetOpenAIConfig(apiKey, baseURL, organization, embeddingModel)

	// Reinitialize to apply changes
	return s.Initialize()
}

// SetOllamaConfig updates the Ollama configuration
func (s *Service) SetOllamaConfig(baseURL, model string, timeout int) error {
	s.cfg.SetOllamaConfig(baseURL, model, timeout)

	// Reinitialize to apply changes
	return s.Initialize()
}

// Status returns the current status of the AI service
type ServiceStatus struct {
	CurrentProvider     string   `json:"current_provider"`
	AvailableProviders  []string `json:"available_providers"`
	CurrentModel        string   `json:"current_model"`
	ChunkingStrategy    string   `json:"chunking_strategy"`
	AvailableStrategies []string `json:"available_strategies"`
	ModelDimension      int      `json:"model_dimension"`
	ProviderHealthy     bool     `json:"provider_healthy"`
}

// GetStatus returns the current status of the AI service
func (s *Service) GetStatus() (*ServiceStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &ServiceStatus{
		CurrentProvider:     s.currentProvider,
		AvailableProviders:  s.getAvailableProvidersLocked(),
		CurrentModel:        s.cfg.GetEmbeddingModel(),
		ChunkingStrategy:    s.cfg.GetChunkingConfig().Strategy,
		AvailableStrategies: s.getAvailableStrategiesLocked(),
	}

	// Get model dimension
	if provider, ok := s.providers[s.currentProvider]; ok {
		status.ProviderHealthy = provider.ValidateConfig() == nil
		model := status.CurrentModel
		if model == "" {
			model = provider.GetDefaultModel()
		}
		if dim, err := provider.GetModelDimension(model); err == nil {
			status.ModelDimension = dim
		}
	}

	return status, nil
}

// retryWithBackoff executes an operation with exponential backoff retries
func retryWithBackoff(operation func() error) error {
	maxRetries := 3
	backoff := 500 * time.Millisecond

	var err error
	for i := 0; i < maxRetries; i++ {
		if err = operation(); err == nil {
			return nil
		}

		// Don't sleep after the last attempt
		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return err
}
