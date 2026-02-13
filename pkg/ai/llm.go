package ai

import (
	"context"
	"time"
)

// LLMProvider defines the interface for language model providers
// This interface extends the AI service capabilities from embeddings to text generation
type LLMProvider interface {
	// GenerateCompletion generates a text completion
	GenerateCompletion(req *CompletionRequest) (*CompletionResponse, error)

	// GenerateCompletionStream generates a streaming completion
	// Returns a channel that receives chunks as they are generated
	GenerateCompletionStream(req *CompletionRequest) (<-chan *CompletionChunk, error)

	// GetAvailableModels returns a list of available models
	GetAvailableModels() ([]string, error)

	// GetDefaultModel returns the default model name
	GetDefaultModel() string

	// ValidateConfig checks if the provider configuration is valid
	ValidateConfig() error

	// Name returns the provider name
	Name() string
}

// CompletionRequest represents a request for text generation
type CompletionRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string       `json:"model"`
	Temperature float32      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens"`
	Stream      bool         `json:"stream"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// CompletionResponse represents a response from LLM
type CompletionResponse struct {
	Content      string      `json:"content"`
	Model        string      `json:"model"`
	TokensUsed   *TokenUsage `json:"tokens_used,omitempty"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

// CompletionChunk represents a streaming chunk from LLM
type CompletionChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   error  `json:"error,omitempty"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamTimeout is the default timeout for streaming requests
const StreamTimeout = 120 * time.Second

// DefaultMaxTokens is the default maximum tokens for completion
const DefaultMaxTokens = 2000

// DefaultTemperature is the default temperature for completion
const DefaultTemperature = float32(0.7)

// DefaultSystemPrompt is the default system prompt for RAG
const DefaultSystemPrompt = `You are a helpful assistant for a personal knowledge management system called Notebit.

You will be given context from the user's notes and asked questions. Use only the provided context to answer questions. If the context doesn't contain enough information to answer the question, say so clearly.

Always cite your sources using the [Source N] notation where N is the source number provided in the context.

Keep your answers concise and directly address the user's question.`

// DefaultChatModels defines the available chat models for each provider
var DefaultChatModels = map[string][]string{
	"openai": {"gpt-4o-mini", "gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"},
	"ollama": {"llama3.2", "llama3.1", "llama3", "mistral", "qwen2.5"},
}

// GetDefaultChatModel returns the default model for a given provider
func GetDefaultChatModel(provider string) string {
	if models, ok := DefaultChatModels[provider]; ok && len(models) > 0 {
		return models[0]
	}
	return "gpt-4o-mini" // Default fallback
}

// IsChatModelAvailable checks if a model is in the available list
func IsChatModelAvailable(provider, model string) bool {
	if models, ok := DefaultChatModels[provider]; ok {
		for _, m := range models {
			if m == model {
				return true
			}
		}
	}
	return false
}

// ContextWithTimeout adds timeout to a context for streaming requests
func ContextWithStreamTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, StreamTimeout)
}
