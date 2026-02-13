package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// Config holds the application configuration
type Config struct {
	mu         sync.RWMutex
	configPath string

	// AI Configuration
	AI AIConfig `json:"ai"`

	// Chunking Configuration
	Chunking ChunkingConfig `json:"chunking"`

	// Watcher Configuration
	Watcher WatcherConfig `json:"watcher"`

	// LLM Configuration (for chat completion)
	LLM LLMConfig `json:"llm"`

	// RAG Configuration
	RAG RAGConfig `json:"rag"`

	// Graph Configuration
	Graph GraphConfig `json:"graph"`
}

// AIConfig holds AI service configuration
type AIConfig struct {
	// Provider is the default embedding provider ("openai" or "ollama")
	Provider string `json:"provider"`

	// OpenAI Configuration
	OpenAI OpenAIConfig `json:"openai"`

	// Ollama Configuration
	Ollama OllamaConfig `json:"ollama"`

	// EmbeddingModel is the default model to use for embeddings
	EmbeddingModel string `json:"embedding_model"`

	// BatchSize is the number of texts to embed in a single batch request
	BatchSize int `json:"batch_size"`
}

// OpenAIConfig holds OpenAI-specific configuration
type OpenAIConfig struct {
	// APIKey is the OpenAI API key
	APIKey string `json:"api_key"`

	// BaseURL is the optional custom base URL (for Azure or proxy)
	BaseURL string `json:"base_url"`

	// Organization is the optional OpenAI organization ID
	Organization string `json:"organization"`

	// Default models
	EmbeddingModel string `json:"embedding_model"` // e.g., "text-embedding-3-small", "text-embedding-3-large"
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	// BaseURL is the Ollama server URL (default: http://localhost:11434)
	BaseURL string `json:"base_url"`

	// EmbeddingModel is the default embedding model (e.g., "nomic-embed-text", "mxbai-embed-large")
	EmbeddingModel string `json:"embedding_model"`

	// Timeout is the request timeout in seconds
	Timeout int `json:"timeout"`
}

// ChunkingConfig holds text chunking configuration
type ChunkingConfig struct {
	// Strategy is the chunking strategy to use ("fixed", "heading", "sliding")
	Strategy string `json:"strategy"`

	// ChunkSize is the target size of each chunk in characters
	ChunkSize int `json:"chunk_size"`

	// ChunkOverlap is the overlap between adjacent chunks in characters
	ChunkOverlap int `json:"chunk_overlap"`

	// MinChunkSize is the minimum chunk size (smaller chunks will be merged)
	MinChunkSize int `json:"min_chunk_size"`

	// MaxChunkSize is the maximum chunk size (larger chunks will be split)
	MaxChunkSize int `json:"max_chunk_size"`

	// PreserveHeading indicates whether to include heading context in chunks
	PreserveHeading bool `json:"preserve_heading"`

	// HeadingSeparator is the separator used between heading and content (default: "\n\n")
	HeadingSeparator string `json:"heading_separator"`
}

// WatcherConfig holds file watcher configuration
type WatcherConfig struct {
	// Enabled enables automatic file watching and indexing
	Enabled bool `json:"enabled"`

	// DebounceMS is the debounce delay in milliseconds for file events
	DebounceMS int `json:"debounce_ms"`

	// Workers is the number of concurrent indexing workers
	Workers int `json:"workers"`

	// FullIndexOnStart enables full background indexing on startup
	FullIndexOnStart bool `json:"full_index_on_start"`
}

// LLMConfig holds LLM (chat completion) configuration
type LLMConfig struct {
	// Provider is the LLM provider ("openai")
	Provider string `json:"provider"`

	// Model is the default chat model
	Model string `json:"model"`

	// Temperature controls response randomness (0.0 - 2.0)
	Temperature float32 `json:"temperature"`

	// MaxTokens is the maximum tokens for completion
	MaxTokens int `json:"max_tokens"`

	// OpenAI Configuration for Chat
	OpenAI OpenAIConfig `json:"openai"`

	// Ollama Configuration for Chat
	Ollama OllamaConfig `json:"ollama"`
}

// RAGConfig holds RAG (Retrieval Augmented Generation) configuration
type RAGConfig struct {
	// MaxContextChunks is the maximum number of chunks to include as context
	MaxContextChunks int `json:"max_context_chunks"`

	// Temperature for RAG responses
	Temperature float32 `json:"temperature"`

	// SystemPrompt is the system prompt for RAG
	SystemPrompt string `json:"system_prompt"`
}

// GraphConfig holds knowledge graph configuration
type GraphConfig struct {
	// MinSimilarityThreshold is the minimum similarity for implicit links
	MinSimilarityThreshold float32 `json:"min_similarity_threshold"`

	// MaxNodes is the maximum number of nodes to display
	MaxNodes int `json:"max_nodes"`

	// ShowImplicitLinks controls whether to show semantic similarity links
	ShowImplicitLinks bool `json:"show_implicit_links"`
}

var (
	globalConfig *Config
	once         sync.Once
)

func New() *Config {
	cfg := &Config{}
	cfg.setDefaults()
	return cfg
}

// Get returns the global configuration instance
func Get() *Config {
	once.Do(func() {
		globalConfig = New()
	})
	return globalConfig
}

// setDefaults sets default values for configuration
func (c *Config) setDefaults() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// AI Defaults
	c.AI.Provider = "ollama" // Default to local-first approach
	c.AI.BatchSize = 32

	// OpenAI Defaults
	c.AI.OpenAI.EmbeddingModel = "text-embedding-3-small"
	c.AI.OpenAI.BaseURL = "https://api.openai.com/v1"

	// Ollama Defaults
	c.AI.Ollama.BaseURL = "http://localhost:11434"
	c.AI.Ollama.EmbeddingModel = "nomic-embed-text"
	c.AI.Ollama.Timeout = 30

	// Set default model based on provider
	c.AI.EmbeddingModel = c.AI.Ollama.EmbeddingModel

	// Chunking Defaults
	c.Chunking.Strategy = "heading"
	c.Chunking.ChunkSize = 1000
	c.Chunking.ChunkOverlap = 200
	c.Chunking.MinChunkSize = 100
	c.Chunking.MaxChunkSize = 4000
	c.Chunking.PreserveHeading = true
	c.Chunking.HeadingSeparator = "\n\n"

	// Watcher Defaults
	c.Watcher.Enabled = true
	c.Watcher.DebounceMS = 500
	c.Watcher.Workers = 3
	c.Watcher.FullIndexOnStart = true

	// LLM Defaults
	c.LLM.Provider = "openai"
	c.LLM.Model = "gpt-4o-mini"
	c.LLM.Temperature = 0.7
	c.LLM.MaxTokens = 2000

	// RAG Defaults
	c.RAG.MaxContextChunks = 5
	c.RAG.Temperature = 0.7
	// SystemPrompt set at runtime, uses ai.DefaultSystemPrompt as default

	// Graph Defaults
	c.Graph.MinSimilarityThreshold = 0.75
	c.Graph.MaxNodes = 100
	c.Graph.ShowImplicitLinks = true
}

// LoadFromFile loads configuration from a JSON file
func (c *Config) LoadFromFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configPath = path

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File doesn't exist, use defaults
			return nil
		}
		return err
	}

	// Create a temporary config to unmarshal into
	temp := Config{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Merge with defaults (keep defaults for unset fields)
	c.mergeWithDefaults(&temp)

	return nil
}

// SaveToFile saves the current configuration to a JSON file
func (c *Config) SaveToFile(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Save saves the configuration to the last loaded path
func (c *Config) Save() error {
	c.mu.RLock()
	path := c.configPath
	c.mu.RUnlock()
	if path == "" {
		return errors.New("no config path set")
	}
	return c.SaveToFile(path)
}

// mergeWithDefaults merges loaded config with defaults
func (c *Config) mergeWithDefaults(loaded *Config) {
	// AI Provider
	if loaded.AI.Provider != "" {
		c.AI.Provider = loaded.AI.Provider
	}

	// OpenAI Config
	if loaded.AI.OpenAI.APIKey != "" {
		c.AI.OpenAI.APIKey = loaded.AI.OpenAI.APIKey
	}
	if loaded.AI.OpenAI.BaseURL != "" {
		c.AI.OpenAI.BaseURL = loaded.AI.OpenAI.BaseURL
	}
	if loaded.AI.OpenAI.Organization != "" {
		c.AI.OpenAI.Organization = loaded.AI.OpenAI.Organization
	}
	if loaded.AI.OpenAI.EmbeddingModel != "" {
		c.AI.OpenAI.EmbeddingModel = loaded.AI.OpenAI.EmbeddingModel
	}

	// Ollama Config
	if loaded.AI.Ollama.BaseURL != "" {
		c.AI.Ollama.BaseURL = loaded.AI.Ollama.BaseURL
	}
	if loaded.AI.Ollama.EmbeddingModel != "" {
		c.AI.Ollama.EmbeddingModel = loaded.AI.Ollama.EmbeddingModel
	}
	if loaded.AI.Ollama.Timeout > 0 {
		c.AI.Ollama.Timeout = loaded.AI.Ollama.Timeout
	}

	// AI Config
	if loaded.AI.EmbeddingModel != "" {
		c.AI.EmbeddingModel = loaded.AI.EmbeddingModel
	}
	if loaded.AI.BatchSize > 0 {
		c.AI.BatchSize = loaded.AI.BatchSize
	}

	// Chunking Config
	if loaded.Chunking.Strategy != "" {
		c.Chunking.Strategy = loaded.Chunking.Strategy
	}
	if loaded.Chunking.ChunkSize > 0 {
		c.Chunking.ChunkSize = loaded.Chunking.ChunkSize
	}
	if loaded.Chunking.ChunkOverlap >= 0 {
		c.Chunking.ChunkOverlap = loaded.Chunking.ChunkOverlap
	}
	if loaded.Chunking.MinChunkSize > 0 {
		c.Chunking.MinChunkSize = loaded.Chunking.MinChunkSize
	}
	if loaded.Chunking.MaxChunkSize > 0 {
		c.Chunking.MaxChunkSize = loaded.Chunking.MaxChunkSize
	}
	// Always load boolean and string values
	c.Chunking.PreserveHeading = loaded.Chunking.PreserveHeading
	if loaded.Chunking.HeadingSeparator != "" {
		c.Chunking.HeadingSeparator = loaded.Chunking.HeadingSeparator
	}

	// Watcher Config
	c.Watcher.Enabled = loaded.Watcher.Enabled
	if loaded.Watcher.DebounceMS > 0 {
		c.Watcher.DebounceMS = loaded.Watcher.DebounceMS
	}
	if loaded.Watcher.Workers > 0 {
		c.Watcher.Workers = loaded.Watcher.Workers
	}
	c.Watcher.FullIndexOnStart = loaded.Watcher.FullIndexOnStart

	// LLM Config
	if loaded.LLM.Provider != "" {
		c.LLM.Provider = loaded.LLM.Provider
	}
	if loaded.LLM.Model != "" {
		c.LLM.Model = loaded.LLM.Model
	}
	if loaded.LLM.Temperature >= 0 {
		c.LLM.Temperature = loaded.LLM.Temperature
	}
	if loaded.LLM.MaxTokens > 0 {
		c.LLM.MaxTokens = loaded.LLM.MaxTokens
	}
	// LLM OpenAI
	if loaded.LLM.OpenAI.APIKey != "" {
		c.LLM.OpenAI.APIKey = loaded.LLM.OpenAI.APIKey
	}
	if loaded.LLM.OpenAI.BaseURL != "" {
		c.LLM.OpenAI.BaseURL = loaded.LLM.OpenAI.BaseURL
	}
	if loaded.LLM.OpenAI.Organization != "" {
		c.LLM.OpenAI.Organization = loaded.LLM.OpenAI.Organization
	}
	// LLM Ollama
	if loaded.LLM.Ollama.BaseURL != "" {
		c.LLM.Ollama.BaseURL = loaded.LLM.Ollama.BaseURL
	}
	if loaded.LLM.Ollama.EmbeddingModel != "" {
		c.LLM.Ollama.EmbeddingModel = loaded.LLM.Ollama.EmbeddingModel
	}
	if loaded.LLM.Ollama.Timeout > 0 {
		c.LLM.Ollama.Timeout = loaded.LLM.Ollama.Timeout
	}

	// RAG Config
	if loaded.RAG.MaxContextChunks > 0 {
		c.RAG.MaxContextChunks = loaded.RAG.MaxContextChunks
	}
	if loaded.RAG.Temperature >= 0 {
		c.RAG.Temperature = loaded.RAG.Temperature
	}
	if loaded.RAG.SystemPrompt != "" {
		c.RAG.SystemPrompt = loaded.RAG.SystemPrompt
	}

	// Graph Config
	if loaded.Graph.MinSimilarityThreshold >= 0 {
		c.Graph.MinSimilarityThreshold = loaded.Graph.MinSimilarityThreshold
	}
	if loaded.Graph.MaxNodes > 0 {
		c.Graph.MaxNodes = loaded.Graph.MaxNodes
	}
	c.Graph.ShowImplicitLinks = loaded.Graph.ShowImplicitLinks
}

// SetOpenAIConfig sets the OpenAI configuration
func (c *Config) SetOpenAIConfig(apiKey, baseURL, organization string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AI.OpenAI.APIKey = apiKey
	if baseURL != "" {
		c.AI.OpenAI.BaseURL = baseURL
	}
	if organization != "" {
		c.AI.OpenAI.Organization = organization
	}
}

// SetOllamaConfig sets the Ollama configuration
func (c *Config) SetOllamaConfig(baseURL, embeddingModel string, timeout int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if baseURL != "" {
		c.AI.Ollama.BaseURL = baseURL
	}
	if embeddingModel != "" {
		c.AI.Ollama.EmbeddingModel = embeddingModel
	}
	if timeout > 0 {
		c.AI.Ollama.Timeout = timeout
	}
}

// SetProvider sets the default embedding provider
func (c *Config) SetProvider(provider string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AI.Provider = provider
}

// GetProvider returns the current embedding provider
func (c *Config) GetProvider() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AI.Provider
}

// SetEmbeddingModel sets the default embedding model
func (c *Config) SetEmbeddingModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.AI.EmbeddingModel = model
}

// GetEmbeddingModel returns the default embedding model
func (c *Config) GetEmbeddingModel() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AI.EmbeddingModel
}

// GetOpenAIConfig returns a copy of the OpenAI configuration
func (c *Config) GetOpenAIConfig() OpenAIConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AI.OpenAI
}

// GetOllamaConfig returns a copy of the Ollama configuration
func (c *Config) GetOllamaConfig() OllamaConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AI.Ollama
}

// GetChunkingConfig returns a copy of the chunking configuration
func (c *Config) GetChunkingConfig() ChunkingConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Chunking
}

// SetChunkingConfig sets the chunking configuration
func (c *Config) SetChunkingConfig(cfg ChunkingConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Chunking = cfg
}

// ModelDimensions returns the expected output dimension for a given model
func (c *Config) ModelDimensions(model string) int {
	// Known model dimensions
	dimensions := map[string]int{
		// OpenAI models
		"text-embedding-3-small": 1536,
		"text-embedding-3-large": 3072,
		"text-embedding-ada-002": 1536,

		// Ollama models
		"nomic-embed-text":  768,
		"mxbai-embed-large": 1024,
		"all-minilm":        384,
		"llama2":            4096, // fallback
	}

	if dim, ok := dimensions[model]; ok {
		return dim
	}
	return 1536 // Default fallback
}

// IsOpenAIConfigured checks if OpenAI is properly configured
func (c *Config) IsOpenAIConfigured() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AI.OpenAI.APIKey != ""
}

// IsOllamaConfigured checks if Ollama is properly configured
func (c *Config) IsOllamaConfigured() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.AI.Ollama.BaseURL != ""
}

// GetWatcherConfig returns a copy of the watcher configuration
func (c *Config) GetWatcherConfig() WatcherConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Watcher
}

// SetWatcherConfig sets the watcher configuration
func (c *Config) SetWatcherConfig(cfg WatcherConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Watcher = cfg
}

// GetLLMConfig returns a copy of the LLM configuration
func (c *Config) GetLLMConfig() LLMConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.LLM
}

// SetLLMConfig sets the LLM configuration
func (c *Config) SetLLMConfig(cfg LLMConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.LLM = cfg
}

// SetLLMOpenAIConfig sets the OpenAI configuration for LLM
func (c *Config) SetLLMOpenAIConfig(apiKey, baseURL, organization string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.LLM.OpenAI.APIKey = apiKey
	if baseURL != "" {
		c.LLM.OpenAI.BaseURL = baseURL
	}
	if organization != "" {
		c.LLM.OpenAI.Organization = organization
	}
}

// GetRAGConfig returns a copy of the RAG configuration
func (c *Config) GetRAGConfig() RAGConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.RAG
}

// SetRAGConfig sets the RAG configuration
func (c *Config) SetRAGConfig(cfg RAGConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.RAG = cfg
}

// GetGraphConfig returns a copy of the Graph configuration
func (c *Config) GetGraphConfig() GraphConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.Graph
}

// SetGraphConfig sets the Graph configuration
func (c *Config) SetGraphConfig(cfg GraphConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Graph = cfg
}
