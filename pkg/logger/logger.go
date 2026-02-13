package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	config       atomic.Value // Stores Config
	logChan      chan LogEntry
	writer       *FileWriter
	kafkaWriter  *KafkaWriter
	writerMu     sync.Mutex // Protects writer replacement
	wg           sync.WaitGroup
	isClosed     atomic.Bool
	consoleOut   io.Writer
	metrics      *Metrics
	batchBuffer  []LogEntry
	batchMu      sync.Mutex
	flushTicker  *time.Ticker
	doneChan     chan struct{} // Signal channel for graceful shutdown
}

var defaultLogger *Logger
var once sync.Once

// New creates a new Logger instance
func New(cfg Config) (*Logger, error) {
	// Set defaults
	if cfg.AsyncBufferSize <= 0 {
		cfg.AsyncBufferSize = 1000
	}
	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = 100 * 1024 * 1024 // Default 100MB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 15 // Default 15 days
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 100 // 100ms
	}
	if cfg.KafkaTopic == "" {
		cfg.KafkaTopic = "app-logs"
	}

	fw, err := NewFileWriter(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize Kafka writer if enabled
	kw, err := NewKafkaWriter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kafka writer: %w", err)
	}

	l := &Logger{
		logChan:     make(chan LogEntry, cfg.AsyncBufferSize),
		writer:      fw,
		kafkaWriter: kw,
		consoleOut:  os.Stdout,
		metrics:     NewMetrics(),
		batchBuffer: make([]LogEntry, 0, cfg.BatchSize),
		flushTicker: time.NewTicker(time.Duration(cfg.FlushInterval) * time.Millisecond),
		doneChan:    make(chan struct{}),
	}
	l.config.Store(cfg)

	// Start log processor
	l.wg.Add(1)
	go l.processLogs()

	// Start flush timer
	l.wg.Add(1)
	go l.periodicFlush()

	return l, nil
}

// GetDefault returns the global default logger
func GetDefault() *Logger {
	return defaultLogger
}

// SetDefault sets the global default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Initialize initializes the global default logger
func Initialize(cfg Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(cfg)
	})
	return err
}

// periodicFlush flushes the batch buffer periodically
func (l *Logger) periodicFlush() {
	defer l.wg.Done()
	for {
		select {
		case <-l.flushTicker.C:
			l.flushBatch()
		case <-l.doneChan:
			return
		}
	}
}

func (l *Logger) processLogs() {
	defer l.wg.Done()
	for entry := range l.logChan {
		l.metrics.IncrementLevel(entry.Level)
		l.addToBatch(entry)
	}
	// Flush remaining logs on shutdown
	l.flushBatch()
}

func (l *Logger) addToBatch(entry LogEntry) {
	l.batchMu.Lock()
	defer l.batchMu.Unlock()

	l.batchBuffer = append(l.batchBuffer, entry)
	cfg := l.config.Load().(Config)
	
	if len(l.batchBuffer) >= cfg.BatchSize {
		l.flushBatchLocked()
	}
}

func (l *Logger) flushBatch() {
	l.batchMu.Lock()
	defer l.batchMu.Unlock()
	l.flushBatchLocked()
}

func (l *Logger) flushBatchLocked() {
	if len(l.batchBuffer) == 0 {
		return
	}

	startTime := time.Now()
	batchSize := len(l.batchBuffer)

	for _, entry := range l.batchBuffer {
		l.writeEntry(entry)
	}

	l.batchBuffer = l.batchBuffer[:0] // Clear buffer
	
	duration := time.Since(startTime)
	l.metrics.RecordFlushLatency(duration)
	l.metrics.RecordBatch(batchSize)
}

func (l *Logger) writeEntry(entry LogEntry) {
	cfg := l.config.Load().(Config)

	// Format message
	msg := l.formatEntry(entry, false)

	// Write to file
	l.writerMu.Lock()
	if l.writer != nil {
		l.writer.Write([]byte(msg))
	}
	l.writerMu.Unlock()

	// Write to console if enabled
	if cfg.ConsoleOutput {
		if cfg.ConsoleColor {
			colorMsg := l.formatEntryWithColor(entry)
			fmt.Fprint(l.consoleOut, colorMsg)
		} else {
			fmt.Fprint(l.consoleOut, msg)
		}
	}

	// Write to Kafka if enabled
	if l.kafkaWriter != nil {
		// Non-blocking write to Kafka
		go l.kafkaWriter.Write(entry)
	}
}

func (l *Logger) formatEntry(entry LogEntry, withColor bool) string {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	
	msg := fmt.Sprintf("%s [%s] [%d]",
		timestamp,
		entry.Level.String(),
		entry.ThreadID,
	)

	if entry.TraceID != "" {
		msg += fmt.Sprintf(" [%s]", entry.TraceID)
	}

	msg += fmt.Sprintf(" [%s.%s:%d]",
		entry.ClassName,
		entry.MethodName,
		entry.Line,
	)

	if entry.Duration > 0 {
		msg += fmt.Sprintf(" [%dms]", entry.Duration.Milliseconds())
	}

	msg += fmt.Sprintf(" - %s", entry.Message)

	// Add fields if present
	if len(entry.Fields) > 0 {
		msg += fmt.Sprintf(" %v", entry.Fields)
	}

	msg += "\n"
	return msg
}

func (l *Logger) formatEntryWithColor(entry LogEntry) string {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	
	msg := fmt.Sprintf("%s %s[%s]%s [%d]",
		timestamp,
		entry.Level.Color(),
		entry.Level.String(),
		ColorReset,
		entry.ThreadID,
	)

	if entry.TraceID != "" {
		msg += fmt.Sprintf(" [\033[1m%s\033[0m]", entry.TraceID) // Bold trace ID
	}

	msg += fmt.Sprintf(" [%s.%s:%d]",
		entry.ClassName,
		entry.MethodName,
		entry.Line,
	)

	if entry.Duration > 0 {
		msg += fmt.Sprintf(" [\033[1;33m%dms\033[0m]", entry.Duration.Milliseconds()) // Yellow duration
	}

	msg += fmt.Sprintf(" - %s", entry.Message)

	if len(entry.Fields) > 0 {
		msg += fmt.Sprintf(" \033[90m%v\033[0m", entry.Fields) // Gray fields
	}

	msg += "\n"
	return msg
}

func (l *Logger) log(level Level, msg string) {
	l.logWithContext(nil, level, msg, nil, 0)
}

func (l *Logger) logWithContext(ctx context.Context, level Level, msg string, fields map[string]interface{}, duration time.Duration) {
	if l.isClosed.Load() {
		return
	}

	cfg := l.config.Load().(Config)
	if level < cfg.Level {
		return
	}

	// Capture caller info (skip 3 levels: logWithContext -> log/DebugCtx -> Debug/Info/...)
	fileName, funcName, line := getCallerInfo(4)

	// Extract trace ID and fields from context
	traceID := GetTraceID(ctx)
	ctxFields := GetFields(ctx)

	// Merge context fields with provided fields
	mergedFields := make(map[string]interface{})
	for k, v := range ctxFields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	entry := LogEntry{
		Time:       time.Now(),
		Level:      level,
		ThreadID:   getGID(),
		TraceID:    traceID,
		ClassName:  fileName,
		MethodName: funcName,
		Line:       line,
		Message:    msg,
		Fields:     mergedFields,
		Duration:   duration,
	}

	// Smart dropping strategy: prefer dropping DEBUG logs when buffer is full
	select {
	case l.logChan <- entry:
		// Successfully queued
		l.metrics.UpdateQueueLength(len(l.logChan))
	default:
		// Buffer full - implement smart dropping
		if level == DEBUG {
			// Drop DEBUG logs to prevent blocking
			l.metrics.IncrementDropped()
		} else {
			// For higher priority logs, try harder to deliver
			// Drop oldest DEBUG log if possible
			select {
			case l.logChan <- entry:
				l.metrics.UpdateQueueLength(len(l.logChan))
			default:
				// Still full, log to stderr as last resort
				l.metrics.IncrementDropped()
				fmt.Fprintf(os.Stderr, "[LOGGER] Buffer full, dropping %s: %s\n", level.String(), msg)
			}
		}
	}
}

// Public API

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.log(FATAL, msg)
	// Ensure fatal logs are flushed
	l.Close()
	os.Exit(1)
}

// Context-aware logging methods

func (l *Logger) DebugCtx(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, DEBUG, fmt.Sprintf(format, args...), nil, 0)
}

func (l *Logger) InfoCtx(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, INFO, fmt.Sprintf(format, args...), nil, 0)
}

func (l *Logger) WarnCtx(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, WARN, fmt.Sprintf(format, args...), nil, 0)
}

func (l *Logger) ErrorCtx(ctx context.Context, format string, args ...interface{}) {
	l.logWithContext(ctx, ERROR, fmt.Sprintf(format, args...), nil, 0)
}

func (l *Logger) FatalCtx(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.logWithContext(ctx, FATAL, msg, nil, 0)
	l.Close()
	os.Exit(1)
}

// Logging with additional fields

func (l *Logger) DebugWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	l.logWithContext(ctx, DEBUG, fmt.Sprintf(format, args...), fields, 0)
}

func (l *Logger) InfoWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	l.logWithContext(ctx, INFO, fmt.Sprintf(format, args...), fields, 0)
}

func (l *Logger) WarnWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	l.logWithContext(ctx, WARN, fmt.Sprintf(format, args...), fields, 0)
}

func (l *Logger) ErrorWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	l.logWithContext(ctx, ERROR, fmt.Sprintf(format, args...), fields, 0)
}

// Logging with duration (performance tracking)

func (l *Logger) DebugWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	l.logWithContext(ctx, DEBUG, fmt.Sprintf(format, args...), nil, duration)
}

func (l *Logger) InfoWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	l.logWithContext(ctx, INFO, fmt.Sprintf(format, args...), nil, duration)
}

func (l *Logger) WarnWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	l.logWithContext(ctx, WARN, fmt.Sprintf(format, args...), nil, duration)
}

func (l *Logger) ErrorWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	l.logWithContext(ctx, ERROR, fmt.Sprintf(format, args...), nil, duration)
}

// Dynamic Configuration

func (l *Logger) SetLevel(level Level) {
	cfg := l.config.Load().(Config)
	cfg.Level = level
	l.config.Store(cfg)
}

func (l *Logger) GetLevel() Level {
	cfg := l.config.Load().(Config)
	return cfg.Level
}

func (l *Logger) SetLogDir(dir string) error {
	l.writerMu.Lock()
	defer l.writerMu.Unlock()

	cfg := l.config.Load().(Config)
	cfg.LogDir = dir
	
	newWriter, err := NewFileWriter(cfg)
	if err != nil {
		return err
	}

	oldWriter := l.writer
	l.writer = newWriter
	l.config.Store(cfg)

	if oldWriter != nil {
		oldWriter.Close()
	}
	return nil
}

// GetMetrics returns current logger metrics
func (l *Logger) GetMetrics() MetricsSnapshot {
	return l.metrics.GetSnapshot()
}

// Close gracefully shuts down the logger
func (l *Logger) Close() {
	if l.isClosed.CompareAndSwap(false, true) {
		// Stop the flush ticker
		if l.flushTicker != nil {
			l.flushTicker.Stop()
		}
		
		// Signal periodic flush to stop
		close(l.doneChan)
		
		// Close log channel to stop processing
		close(l.logChan)
		
		// Wait for all goroutines to finish
		l.wg.Wait()
		
		// Close writers
		l.writerMu.Lock()
		if l.writer != nil {
			l.writer.Close()
		}
		if l.kafkaWriter != nil {
			l.kafkaWriter.Close()
		}
		l.writerMu.Unlock()
	}
}

// Global helpers

func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	} else {
		// Fallback if no logger initialized
		fmt.Printf("FATAL: "+format+"\n", args...)
		os.Exit(1)
	}
}

// Context-aware global helpers

func DebugCtx(ctx context.Context, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.DebugCtx(ctx, format, args...)
	}
}

func InfoCtx(ctx context.Context, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.InfoCtx(ctx, format, args...)
	}
}

func WarnCtx(ctx context.Context, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.WarnCtx(ctx, format, args...)
	}
}

func ErrorCtx(ctx context.Context, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.ErrorCtx(ctx, format, args...)
	}
}

func FatalCtx(ctx context.Context, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.FatalCtx(ctx, format, args...)
	} else {
		fmt.Printf("FATAL: "+format+"\n", args...)
		os.Exit(1)
	}
}

// Global helpers with fields

func DebugWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.DebugWithFields(ctx, fields, format, args...)
	}
}

func InfoWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.InfoWithFields(ctx, fields, format, args...)
	}
}

func WarnWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.WarnWithFields(ctx, fields, format, args...)
	}
}

func ErrorWithFields(ctx context.Context, fields map[string]interface{}, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.ErrorWithFields(ctx, fields, format, args...)
	}
}

// Global helpers with duration

func DebugWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.DebugWithDuration(ctx, duration, format, args...)
	}
}

func InfoWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.InfoWithDuration(ctx, duration, format, args...)
	}
}

func ErrorWithDuration(ctx context.Context, duration time.Duration, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.ErrorWithDuration(ctx, duration, format, args...)
	}
}

// SetLevel sets the log level for the default logger
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
	}
}

// GetMetrics returns metrics from the default logger
func GetMetrics() MetricsSnapshot {
	if defaultLogger != nil {
		return defaultLogger.GetMetrics()
	}
	return MetricsSnapshot{}
}
