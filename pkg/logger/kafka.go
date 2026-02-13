package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaWriter handles sending logs to Kafka
type KafkaWriter struct {
	writer *kafka.Writer
	mu     sync.Mutex
	config Config
	closed bool
}

// NewKafkaWriter creates a new Kafka writer
func NewKafkaWriter(config Config) (*KafkaWriter, error) {
	if !config.KafkaEnabled || len(config.KafkaBrokers) == 0 {
		return nil, nil // Not enabled, return nil
	}

	topic := config.KafkaTopic
	if topic == "" {
		topic = "app-logs" // Default topic
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.KafkaBrokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Async:        true, // Non-blocking writes
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
		MaxAttempts:  3,
	}

	kw := &KafkaWriter{
		writer: writer,
		config: config,
	}

	return kw, nil
}

// Write sends a log entry to Kafka
func (kw *KafkaWriter) Write(entry LogEntry) error {
	if kw == nil || kw.closed {
		return nil
	}

	kw.mu.Lock()
	defer kw.mu.Unlock()

	// Serialize log entry to JSON
	data, err := json.Marshal(map[string]interface{}{
		"timestamp":  entry.Time.Format(time.RFC3339Nano),
		"level":      entry.Level.String(),
		"trace_id":   entry.TraceID,
		"file":       entry.ClassName,
		"function":   entry.MethodName,
		"line":       entry.Line,
		"message":    entry.Message,
		"fields":     entry.Fields,
		"duration_ms": entry.Duration.Milliseconds(),
		"goroutine_id": entry.ThreadID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Send to Kafka asynchronously
	// Timeout context to prevent blocking too long
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	msg := kafka.Message{
		Key:   []byte(entry.TraceID), // Use TraceID as partition key
		Value: data,
		Time:  entry.Time,
	}

	return kw.writer.WriteMessages(ctx, msg)
}

// Close closes the Kafka writer
func (kw *KafkaWriter) Close() error {
	if kw == nil || kw.closed {
		return nil
	}

	kw.mu.Lock()
	defer kw.mu.Unlock()

	kw.closed = true
	if kw.writer != nil {
		return kw.writer.Close()
	}
	return nil
}
