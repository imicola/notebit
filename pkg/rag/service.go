package rag

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"notebit/pkg/ai"
	"notebit/pkg/config"
	"notebit/pkg/database"
)

// Service handles RAG (Retrieval-Augmented Generation) operations
type Service struct {
	mu  sync.RWMutex
	db  *database.Manager
	ai  *ai.Service
	llm ai.LLMProvider
	cfg *config.Config
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	ID        string     `json:"id"`
	Role      string     `json:"role"` // "user", "assistant", "system"
	Content   string     `json:"content"`
	Sources   []ChunkRef `json:"sources,omitempty"`
	Timestamp int64      `json:"timestamp"`
}

// ChunkRef represents a reference to a source chunk
type ChunkRef struct {
	Path       string  `json:"path"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Heading    string  `json:"heading"`
	Similarity float32 `json:"similarity"`
	ChunkID    uint    `json:"chunk_id"`
}

// ChatResponse represents a response from the RAG service
type ChatResponse struct {
	MessageID  string     `json:"message_id"`
	Content    string     `json:"content"`
	Sources    []ChunkRef `json:"sources"`
	TokensUsed *int       `json:"tokens_used,omitempty"`
}

// NewService creates a new RAG service
func NewService(db *database.Manager, aiSvc *ai.Service, llm ai.LLMProvider, cfg *config.Config) *Service {
	return &Service{
		db:  db,
		ai:  aiSvc,
		llm: llm,
		cfg: cfg,
	}
}

// Query performs a RAG query
func (s *Service) Query(ctx context.Context, query string) (*ChatResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if !s.db.IsInitialized() {
		return nil, fmt.Errorf("database is not initialized")
	}

	if s.llm == nil {
		return nil, fmt.Errorf("LLM provider is not configured")
	}

	// Step 1: Generate query embedding
	queryEmbedding, err := s.ai.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Step 2: Search for similar chunks
	repo := s.db.Repository()
	ragConfig := s.cfg.GetRAGConfig()

	limit := ragConfig.MaxContextChunks
	if limit <= 0 {
		limit = 5 // Default
	}

	similarChunks, err := repo.SearchSimilar(queryEmbedding.Embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar chunks: %w", err)
	}
	if len(similarChunks) == 0 {
		return nil, fmt.Errorf("knowledge base has no indexed context yet, please save or reindex notes first")
	}

	// Step 3: Build context from retrieved chunks
	ragContext := s.buildContext(similarChunks)

	// Step 4: Generate completion with context
	messages := s.buildMessages(query, ragContext, ragConfig)

	completion, err := s.llm.GenerateCompletion(&ai.CompletionRequest{
		Messages:    messages,
		Model:       s.cfg.GetLLMConfig().Model,
		Temperature: ragConfig.Temperature,
		MaxTokens:   s.cfg.GetLLMConfig().MaxTokens,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate completion: %w", err)
	}

	// Step 5: Build response with sources
	sources := s.buildSources(similarChunks)

	var tokensUsed *int
	if completion.TokensUsed != nil {
		tokensUsed = &completion.TokensUsed.TotalTokens
	}

	return &ChatResponse{
		MessageID:  generateMessageID(),
		Content:    completion.Content,
		Sources:    sources,
		TokensUsed: tokensUsed,
	}, nil
}

// buildContext creates context string from chunks
func (s *Service) buildContext(chunks []database.SimilarChunk) string {
	var sb strings.Builder
	sb.WriteString("Context from notes:\n\n")

	for i, chunk := range chunks {
		sourceNum := i + 1
		sb.WriteString(fmt.Sprintf("[Source %d] ", sourceNum))

		// Add file title and heading (nil-safe)
		if chunk.File != nil && chunk.File.Title != "" {
			sb.WriteString(chunk.File.Title)
		}
		if chunk.Heading != "" {
			sb.WriteString(fmt.Sprintf(" > %s", chunk.Heading))
		}
		sb.WriteString("\n")

		// Add content (truncated if too long)
		content := truncateContent(chunk.Content, 500)
		sb.WriteString(fmt.Sprintf("%s\n\n", content))
	}

	return sb.String()
}

// buildMessages constructs the message list for LLM
func (s *Service) buildMessages(query, context string, ragConfig config.RAGConfig) []ai.ChatMessage {
	systemPrompt := ragConfig.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = ai.DefaultSystemPrompt
	}

	return []ai.ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Context:\n%s\n\nQuestion: %s", context, query),
		},
	}
}

// buildSources converts chunks to ChunkRefs
func (s *Service) buildSources(chunks []database.SimilarChunk) []ChunkRef {
	sources := make([]ChunkRef, 0, len(chunks))
	for _, chunk := range chunks {
		ref := ChunkRef{
			Content:    chunk.Content,
			Heading:    chunk.Heading,
			Similarity: chunk.Similarity,
			ChunkID:    chunk.ChunkID,
		}
		if chunk.File != nil {
			ref.Path = chunk.File.Path
			ref.Title = chunk.File.Title
		}
		sources = append(sources, ref)
	}
	return sources
}

// generateMessageID generates a unique message ID using crypto/rand
func generateMessageID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}
	return "msg_" + hex.EncodeToString(b)
}

func truncateContent(content string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(content)
	if len(runes) <= max {
		return content
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}

// GetStatus returns the current status of the RAG service
type ServiceStatus struct {
	Available     bool   `json:"available"`
	LLMProvider   string `json:"llm_provider"`
	LLMModel      string `json:"llm_model"`
	DatabaseReady bool   `json:"database_ready"`
}

// GetStatus returns the status of the RAG service
func (s *Service) GetStatus() *ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &ServiceStatus{
		DatabaseReady: s.db.IsInitialized(),
	}

	if s.llm != nil {
		status.Available = true
		status.LLMProvider = s.llm.Name()
		status.LLMModel = s.cfg.GetLLMConfig().Model
	}

	return status
}

// ValidateConfig checks if the RAG service is properly configured
func (s *Service) ValidateConfig() error {
	if !s.db.IsInitialized() {
		return fmt.Errorf("database is not initialized")
	}

	if s.llm == nil {
		return fmt.Errorf("LLM provider is not configured")
	}

	if err := s.llm.ValidateConfig(); err != nil {
		return fmt.Errorf("LLM provider configuration is invalid: %w", err)
	}

	return nil
}
