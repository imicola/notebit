package ai

import "strings"

// DefaultEmbeddingDimension is the fallback dimension when a model is not found.
const DefaultEmbeddingDimension = 1536

// knownModelDimensions maps embedding model names to their output vector dimensions.
// This is the single source of truth for all dimension lookups across the application.
var knownModelDimensions = map[string]int{
	// OpenAI models
	"text-embedding-3-small": 1536,
	"text-embedding-3-large": 3072,
	"text-embedding-ada-002": 1536,

	// Alibaba Cloud (DashScope/Bailian)
	"text-embedding-v1":  1536,
	"text-embedding-v2":  1536,
	"text-embedding-v3":  1024,
	"text-embedding-v4":  1536,

	// Azure OpenAI (same as OpenAI models)
	"text-embedding-ada-002-v2": 1536,

	// Ollama models
	"nomic-embed-text":  768,
	"mxbai-embed-large": 1024,
	"all-minilm":        384,
	"llama2":            4096,
	"mistral":           4096,
	"mixtral":           4096,
	"codellama":         4096,
	"phi":               2048,
	"gemma":             2048,
	"gemma2":            2048,
}

// LookupModelDimension returns the embedding dimension for a known model.
// It tries exact match first, then strips the ":tag" suffix for Ollama-style names.
// Returns (dimension, true) if found, or (0, false) if unknown.
func LookupModelDimension(model string) (int, bool) {
	if dim, ok := knownModelDimensions[model]; ok {
		return dim, true
	}
	// Strip ":tag" suffix (e.g., "nomic-embed-text:latest" → "nomic-embed-text")
	baseName := strings.Split(model, ":")[0]
	if dim, ok := knownModelDimensions[baseName]; ok {
		return dim, true
	}
	return 0, false
}
