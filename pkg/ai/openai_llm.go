package ai

import (
	"bufio"
	"bytes"
	// "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"notebit/pkg/config"
)

// OpenAILLMProvider implements LLMProvider for OpenAI chat completions
type OpenAILLMProvider struct {
	apiKey       string
	baseURL      string
	organization string
	httpClient   *http.Client
	model        string
}

// NewOpenAILLMProvider creates a new OpenAI LLM provider
func NewOpenAILLMProvider(cfg config.OpenAIConfig) (*OpenAILLMProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAILLMProvider{
		apiKey:       cfg.APIKey,
		baseURL:      baseURL,
		organization: cfg.Organization,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "gpt-4o-mini",
	}, nil
}

// Name returns the provider name
func (p *OpenAILLMProvider) Name() string {
	return "openai"
}

// GetDefaultModel returns the default model
func (p *OpenAILLMProvider) GetDefaultModel() string {
	return "gpt-4o-mini"
}

// GetAvailableModels returns available models
func (p *OpenAILLMProvider) GetAvailableModels() ([]string, error) {
	return DefaultChatModels["openai"], nil
}

// ValidateConfig checks if the configuration is valid
func (p *OpenAILLMProvider) ValidateConfig() error {
	if p.apiKey == "" {
		return fmt.Errorf("OpenAI API key is required")
	}
	return nil
}

// GenerateCompletion generates a text completion
func (p *OpenAILLMProvider) GenerateCompletion(req *CompletionRequest) (*CompletionResponse, error) {
	// Set default model if not specified
	if req.Model == "" {
		req.Model = p.GetDefaultModel()
	}

	// Prepare the request body
	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   false,
	}

	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}

	if req.MaxTokens > 0 {
		requestBody["max_tokens"] = req.MaxTokens
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(p.baseURL, "/"))
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	if p.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", p.organization)
	}

	// Execute request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openAIResp struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract content
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := openAIResp.Choices[0].Message.Content

	return &CompletionResponse{
		Content:      content,
		Model:        openAIResp.Model,
		TokensUsed:   &TokenUsage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
		FinishReason: openAIResp.Choices[0].FinishReason,
	}, nil
}

// GenerateCompletionStream generates a streaming completion
func (p *OpenAILLMProvider) GenerateCompletionStream(req *CompletionRequest) (<-chan *CompletionChunk, error) {
	// Set default model if not specified
	if req.Model == "" {
		req.Model = p.GetDefaultModel()
	}

	// Create output channel
	chunkChan := make(chan *CompletionChunk, 16)

	// Prepare the request body
	requestBody := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}

	if req.Temperature > 0 {
		requestBody["temperature"] = req.Temperature
	}

	if req.MaxTokens > 0 {
		requestBody["max_tokens"] = req.MaxTokens
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		close(chunkChan)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(p.baseURL, "/"))
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		close(chunkChan)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	if p.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", p.organization)
	}

	// Execute request in goroutine
	go func() {
		defer close(chunkChan)

		resp, err := p.httpClient.Do(httpReq)
		if err != nil {
			chunkChan <- &CompletionChunk{Error: fmt.Errorf("request failed: %w", err)}
			return
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			chunkChan <- &CompletionChunk{Error: fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))}
			return
		}

		// Read SSE stream
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// SSE format: "data: {...}"
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Stream end marker
			if data == "[DONE]" {
				chunkChan <- &CompletionChunk{Done: true}
				return
			}

			// Parse chunk
			var streamChunk struct {
				ID      string `json:"id"`
				Object string `json:"object"`
				Created int64  `json:"created"`
				Model   string `json:"model"`
				Choices []struct {
					Index int `json:"index"`
					Delta struct {
						Role    string `json:"role,omitempty"`
						Content string `json:"content,omitempty"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason,omitempty"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &streamChunk); err != nil {
				continue // Skip invalid chunks
			}

			// Extract content
			if len(streamChunk.Choices) > 0 {
				delta := streamChunk.Choices[0].Delta
				if delta.Content != "" {
					chunkChan <- &CompletionChunk{
						Content: delta.Content,
					}
				}
			}

			// Check for finish reason
			if len(streamChunk.Choices) > 0 && streamChunk.Choices[0].FinishReason != "" {
				chunkChan <- &CompletionChunk{Done: true}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			chunkChan <- &CompletionChunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return chunkChan, nil
}
