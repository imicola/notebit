package ai

import "fmt"

// AIError represents AI service related errors
type AIError struct {
	Op  string
	Err error
}

func (e *AIError) Error() string {
	return fmt.Sprintf("ai %s: %v", e.Op, e.Err)
}

func (e *AIError) Unwrap() error {
	return e.Err
}

// Common error constructors
func NewConfigError(err error) error {
	return &AIError{Op: "config", Err: err}
}

func NewEmbeddingError(err error) error {
	return &AIError{Op: "embedding", Err: err}
}

func NewChunkingError(err error) error {
	return &AIError{Op: "chunking", Err: err}
}

func NewProviderError(provider string, err error) error {
	return &AIError{Op: provider, Err: err}
}
