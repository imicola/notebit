package ai

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Text   string  // The text to embed
	Model  string  // Model identifier (e.g., "text-embedding-3-small", "nomic-embed-text")
	Params *Params // Optional parameters
}

// Params contains optional embedding parameters
type Params struct {
	Dimensions     *int   // For OpenAI: output dimensions (default: model dependent)
	EncodingFormat string // For OpenAI: "float" or "base64" (default: "float")
}

// EmbeddingResponse represents the response from an embedding API
type EmbeddingResponse struct {
	Embedding []float32 // The vector embedding
	Model     string    // The model used
	Usage     *Usage    // Token usage information (if available)
}

// Usage represents token usage information
type Usage struct {
	PromptTokens int
	TotalTokens  int
}

// EmbeddingResult represents the result of a batch embedding operation
type EmbeddingResult struct {
	Index     int       // Index in the original batch
	Embedding []float32 // The vector embedding
	Error     error     // Any error that occurred for this item
}

// EmbeddingProvider defines the interface for embedding service providers
type EmbeddingProvider interface {
	// GenerateEmbedding creates an embedding for a single text
	GenerateEmbedding(req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GenerateEmbeddingsBatch creates embeddings for multiple texts in a single request
	GenerateEmbeddingsBatch(texts []string) ([]*EmbeddingResponse, error)

	// GetModelDimension returns the output dimension for a given model
	GetModelDimension(model string) (int, error)

	// GetDefaultModel returns the default model name for this provider
	GetDefaultModel() string

	// ValidateConfig checks if the provider configuration is valid
	ValidateConfig() error

	// Name returns the provider name
	Name() string
}

// TextChunk represents a segment of text with metadata
type TextChunk struct {
	Content   string                 // The chunk content
	Heading   string                 // Associated heading (if any)
	Index     int                    // Position in the original text
	Metadata  map[string]interface{} // Additional metadata
	Embedding []float32              // Vector embedding (populated after processing)
	ModelName string                 // Model used to generate embedding
}

// ChunkingStrategy defines the interface for text chunking strategies
type ChunkingStrategy interface {
	// Chunk splits text into smaller segments for processing
	Chunk(text string) ([]TextChunk, error)

	// Name returns the strategy name
	Name() string
}
