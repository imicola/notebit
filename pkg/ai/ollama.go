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

// OllamaProvider implements EmbeddingProvider for Ollama's local API
type OllamaProvider struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// OllamaConfig holds the configuration for Ollama provider
type OllamaConfig struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

// NewOllamaProvider creates a new Ollama embedding provider
func NewOllamaProvider(cfg OllamaConfig) (*OllamaProvider, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	model := cfg.Model
	if model == "" {
		model = "nomic-embed-text"
	}

	return &OllamaProvider{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

// ollamaEmbeddingRequest is the request body for Ollama's embeddings API
type ollamaEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// ollamaEmbeddingResponse is the response from Ollama's embeddings API
type ollamaEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	Model     string    `json:"model"`
}

// ollamaErrorResponse represents an error response from Ollama
type ollamaErrorResponse struct {
	Error string `json:"error"`
}

// GenerateEmbedding creates an embedding for a single text
func (p *OllamaProvider) GenerateEmbedding(req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	// Build request body
	body := ollamaEmbeddingRequest{
		Model: model,
		Input: req.Text,
	}

	// Marshal request
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := p.baseURL + "api/embeddings"
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")

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
		var errResp ollamaErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("Ollama error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("request failed with status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var resp ollamaEmbeddingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(resp.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return &EmbeddingResponse{
		Embedding: resp.Embedding,
		Model:     resp.Model,
		Usage:     nil, // Ollama doesn't provide token usage
	}, nil
}

// GenerateEmbeddingsBatch creates embeddings for multiple texts
// Note: Ollama API doesn't natively support batch embeddings,
// so we implement it by making multiple parallel requests
func (p *OllamaProvider) GenerateEmbeddingsBatch(texts []string) ([]*EmbeddingResponse, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("texts cannot be empty")
	}

	// Ollama doesn't have a true batch API, so we make parallel requests
	// Use a semaphore to limit concurrency
	const maxConcurrency = 5
	sem := make(chan struct{}, maxConcurrency)
	results := make([]*EmbeddingResponse, len(texts))
	errs := make(chan error, len(texts))

	for i, text := range texts {
		sem <- struct{}{} // Acquire semaphore
		go func(idx int, txt string) {
			defer func() { <-sem }() // Release semaphore

			resp, err := p.GenerateEmbedding(&EmbeddingRequest{Text: txt})
			if err != nil {
				errs <- fmt.Errorf("error at index %d: %w", idx, err)
				return
			}
			results[idx] = resp
		}(i, text)
	}

	// Wait for all goroutines to complete
	for i := 0; i < maxConcurrency; i++ {
		sem <- struct{}{}
	}

	// Check for errors
	close(errs)
	var errList []error
	for err := range errs {
		errList = append(errList, err)
	}

	if len(errList) > 0 {
		return results, fmt.Errorf("encountered %d errors: %v", len(errList), errList[0])
	}

	return results, nil
}

// GetModelDimension returns the output dimension for a given model
func (p *OllamaProvider) GetModelDimension(model string) (int, error) {
	return p.getKnownModelDimension(model), nil
}

// getKnownModelDimension returns known dimensions for common models
func (p *OllamaProvider) getKnownModelDimension(model string) int {
	dimensions := map[string]int{
		"nomic-embed-text":  768,
		"mxbai-embed-large": 1024,
		"all-minilm":        384,
		"llama2":            4096, // fallback for LLaMA models
		"mistral":           4096,
		"mixtral":           4096,
		"codellama":         4096,
		"phi":               2048,
		"gemma":             2048,
		"gemma2":            2048,
	}

	if dim, ok := dimensions[model]; ok {
		return dim
	}

	// Try to extract base model name (remove :tag suffix)
	baseName := strings.Split(model, ":")[0]
	if dim, ok := dimensions[baseName]; ok {
		return dim
	}

	// Default fallback
	return 768
}

// GetDefaultModel returns the default model name
func (p *OllamaProvider) GetDefaultModel() string {
	return p.model
}

// ValidateConfig checks if the provider configuration is valid
func (p *OllamaProvider) ValidateConfig() error {
	if p.baseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	// Try to reach the Ollama server
	url := p.baseURL + "api/tags"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach Ollama server at %s: %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	return nil
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// SetModel sets the model to use for embeddings
func (p *OllamaProvider) SetModel(model string) {
	p.model = model
}

// SetBaseURL sets the base URL for the Ollama API
func (p *OllamaProvider) SetBaseURL(url string) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	p.baseURL = url
}
