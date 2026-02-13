package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type ctxKey string

const (
	traceIDKey ctxKey = "traceID"
	fieldsKey  ctxKey = "fields"
)

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// NewTraceID generates a new unique trace ID
func NewTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// WithFields adds context fields to the context
func WithFields(ctx context.Context, fields map[string]interface{}) context.Context {
	return context.WithValue(ctx, fieldsKey, fields)
}

// GetFields extracts context fields from context
func GetFields(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return nil
	}
	if fields, ok := ctx.Value(fieldsKey).(map[string]interface{}); ok {
		return fields
	}
	return nil
}

// StartTimer returns a function that when called, returns the duration since start
func StartTimer() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}
