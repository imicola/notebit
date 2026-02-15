package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider implements EmbeddingProvider for OpenAI's API
type OpenAIProvider struct {
	apiKey       string
	baseURL      string
	organization string
	model        string
	httpClient   *http.Client
}

// OpenAIConfig holds the configuration for OpenAI provider
type OpenAIConfig struct {
	APIKey         string
	BaseURL        string
	Organization   string
	Timeout        time.Duration
	EmbeddingModel string
}

// NewOpenAIProvider creates a new OpenAI embedding provider
func NewOpenAIProvider(cfg OpenAIConfig) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	model := cfg.EmbeddingModel
	if model == "" {
		model = "text-embedding-3-small"
	}

	return &OpenAIProvider{
		apiKey:       cfg.APIKey,
		baseURL:      baseURL,
		organization: cfg.Organization,
		model:        model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// openAIEmbeddingRequest is the request body for OpenAI's embeddings API
type openAIEmbeddingRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"` // string or []string
	// Optional parameters
	Dimensions     *int   `json:"dimensions,omitempty"`
	EncodingFormat string `json:"encoding_format,omitempty"`
}

// openAIEmbeddingResponse is the response from OpenAI's embeddings API
type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// openAIErrorResponse represents an error response from OpenAI
type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// GenerateEmbedding creates an embedding for a single text
func (p *OpenAIProvider) GenerateEmbedding(req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	model := req.Model
	if model == "" {
		model = p.GetDefaultModel()
	}

	// Build request body
	body := openAIEmbeddingRequest{
		Model: model,
		Input: req.Text,
	}
	if req.Params != nil {
		if req.Params.Dimensions != nil {
			body.Dimensions = req.Params.Dimensions
		}
		if req.Params.EncodingFormat != "" {
			body.EncodingFormat = req.Params.EncodingFormat
		}
	}

	// Marshal request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := p.baseURL + "embeddings"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	if p.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", p.organization)
	}

	// Execute request
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error status
	if httpResp.StatusCode != http.StatusOK {
		var errResp openAIErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI error: %s (type: %s, code: %s)",
				errResp.Error.Message, errResp.Error.Type, errResp.Error.Code)
		}
		return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var resp openAIEmbeddingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// Return first (and only) embedding
	return &EmbeddingResponse{
		Embedding: resp.Data[0].Embedding,
		Model:     resp.Model,
		Usage: &Usage{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}, nil
}

// GenerateEmbeddingsBatch creates embeddings for multiple texts
func (p *OpenAIProvider) GenerateEmbeddingsBatch(texts []string) ([]*EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// OpenAI supports up to 2048 texts in a single batch
	const maxBatchSize = 2048
	if len(texts) > maxBatchSize {
		return nil, fmt.Errorf("batch size exceeds maximum of %d", maxBatchSize)
	}

	model := p.GetDefaultModel()

	// Build request body
	body := openAIEmbeddingRequest{
		Model: model,
		Input: texts,
	}

	// Marshal request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := p.baseURL + "embeddings"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	if p.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", p.organization)
	}

	// Execute request
	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error status
	if httpResp.StatusCode != http.StatusOK {
		var errResp openAIErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI error: %s (type: %s, code: %s)",
				errResp.Error.Message, errResp.Error.Type, errResp.Error.Code)
		}
		return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var resp openAIEmbeddingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	// Convert to response format
	results := make([]*EmbeddingResponse, len(texts))
	for i, item := range resp.Data {
		results[i] = &EmbeddingResponse{
			Embedding: item.Embedding,
			Model:     resp.Model,
		}
	}

	return results, nil
}

// GetModelDimension returns the output dimension for a given model
func (p *OpenAIProvider) GetModelDimension(model string) (int, error) {
	if dim, ok := LookupModelDimension(model); ok {
		return dim, nil
	}
	// Fallback to default dimension for unknown models (supports custom/new models)
	return DefaultEmbeddingDimension, nil
}

// GetDefaultModel returns the default model name
func (p *OpenAIProvider) GetDefaultModel() string {
	if p.model != "" {
		return p.model
	}
	return "text-embedding-3-small"
}

// ValidateConfig checks if the provider configuration is valid
func (p *OpenAIProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("API key is required")
	}
	if p.baseURL == "" {
		return fmt.Errorf("base URL is required")
	}
	return nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}
