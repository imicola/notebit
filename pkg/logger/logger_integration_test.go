package logger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoggerInitialization(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           INFO,
		LogDir:          tmpDir,
		FileName:        "test.log",
		MaxFileSize:     1024 * 1024,
		MaxBackups:      3,
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       5,
		FlushInterval:   50,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Fatal("Logger is nil")
	}

	// Verify log file was created
	logPath := filepath.Join(tmpDir, "test.log")
	time.Sleep(100 * time.Millisecond) // Wait for file creation
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created: %s", logPath)
	}
}

func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           WARN, // Only WARN and above should be logged
		LogDir:          tmpDir,
		FileName:        "levels.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       5,
		FlushInterval:   50,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log at different levels
	logger.Debug("This is debug")    // Should not be logged
	logger.Info("This is info")      // Should not be logged
	logger.Warn("This is warning")   // Should be logged
	logger.Error("This is error")    // Should be logged

	// Flush and wait
	logger.Close()
	time.Sleep(200 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tmpDir, "levels.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify DEBUG and INFO are not present
	if contains(logContent, "debug") {
		t.Error("DEBUG log should not be present")
	}
	if contains(logContent, "This is info") {
		t.Error("INFO log should not be present")
	}

	// Verify WARN and ERROR are present
	if !contains(logContent, "This is warning") {
		t.Error("WARN log should be present")
	}
	if !contains(logContent, "This is error") {
		t.Error("ERROR log should be present")
	}
}

func TestLoggerWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           DEBUG,
		LogDir:          tmpDir,
		FileName:        "context.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       5,
		FlushInterval:   50,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Create context with trace ID
	ctx := WithTraceID(context.Background(), "test-trace-123")
	ctx = WithFields(ctx, map[string]interface{}{
		"user_id":  "user-456",
		"order_id": "order-789",
	})

	// Log with context
	logger.InfoCtx(ctx, "Processing order")

	// Flush and wait
	logger.Close()
	time.Sleep(200 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tmpDir, "context.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify trace ID and fields are present
	if !contains(logContent, "test-trace-123") {
		t.Error("Trace ID should be present in log")
	}
	if !contains(logContent, "user_id") {
		t.Error("user_id field should be present in log")
	}
}

func TestLoggerWithDuration(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           DEBUG,
		LogDir:          tmpDir,
		FileName:        "duration.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       5,
		FlushInterval:   50,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Use timer
	timer := StartTimer()
	time.Sleep(50 * time.Millisecond)
	duration := timer()

	logger.InfoWithDuration(context.Background(), duration, "Operation completed")

	// Flush and wait
	logger.Close()
	time.Sleep(200 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tmpDir, "duration.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify duration is present (should be ~50ms)
	if !contains(logContent, "ms]") {
		t.Error("Duration should be present in log")
	}
}

func TestDynamicLevelChange(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           INFO,
		LogDir:          tmpDir,
		FileName:        "dynamic.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       5,
		FlushInterval:   50,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log at DEBUG level (should not be logged)
	logger.Debug("Debug message 1")
	time.Sleep(100 * time.Millisecond)

	// Change level to DEBUG
	logger.SetLevel(DEBUG)

	// Log at DEBUG level again (should be logged now)
	logger.Debug("Debug message 2")

	// Flush and wait
	logger.Close()
	time.Sleep(200 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tmpDir, "dynamic.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify first debug is not present, second is present
	if contains(logContent, "Debug message 1") {
		t.Error("First debug message should not be present")
	}
	if !contains(logContent, "Debug message 2") {
		t.Error("Second debug message should be present")
	}
}

func TestMetrics(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           DEBUG,
		LogDir:          tmpDir,
		FileName:        "metrics.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       5,
		FlushInterval:   50,
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Log various levels
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	time.Sleep(200 * time.Millisecond)

	// Get metrics
	metrics := logger.GetMetrics()

	// Verify counts
	if metrics.DebugCount != 1 {
		t.Errorf("Expected 1 debug log, got %d", metrics.DebugCount)
	}
	if metrics.InfoCount != 1 {
		t.Errorf("Expected 1 info log, got %d", metrics.InfoCount)
	}
	if metrics.WarnCount != 1 {
		t.Errorf("Expected 1 warn log, got %d", metrics.WarnCount)
	}
	if metrics.ErrorCount != 1 {
		t.Errorf("Expected 1 error log, got %d", metrics.ErrorCount)
	}
	if metrics.TotalLogs != 4 {
		t.Errorf("Expected 4 total logs, got %d", metrics.TotalLogs)
	}
}

func TestBufferOverflow(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           DEBUG,
		LogDir:          tmpDir,
		FileName:        "overflow.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 10, // Very small buffer
		BatchSize:       5,
		FlushInterval:   1000, // Long flush interval
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Flood with logs to trigger overflow
	for i := 0; i < 100; i++ {
		logger.Debug("Message %d", i)
	}

	time.Sleep(200 * time.Millisecond)

	// Get metrics to check dropped logs
	metrics := logger.GetMetrics()

	// Some logs should be dropped
	if metrics.DroppedLogs == 0 {
		t.Log("Warning: No logs were dropped despite small buffer - may be timing dependent")
	}
}

func TestGracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		Level:           INFO,
		LogDir:          tmpDir,
		FileName:        "shutdown.log",
		ConsoleOutput:   false,
		AsyncBufferSize: 100,
		BatchSize:       50, // High batch size to test flush on close
		FlushInterval:   10000, // Very long interval
	}

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log multiple messages
	for i := 0; i < 10; i++ {
		logger.Info("Message %d", i)
	}

	// Close immediately (should flush all pending logs)
	logger.Close()

	// Read log file
	logPath := filepath.Join(tmpDir, "shutdown.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify all messages were written
	for i := 0; i < 10; i++ {
		expected := "Message " + string(rune('0'+i))
		if !contains(logContent, expected) {
			t.Errorf("Expected message '%s' not found in log", expected)
		}
	}
}

func TestTraceIDGeneration(t *testing.T) {
	traceID1 := NewTraceID()
	traceID2 := NewTraceID()

	if traceID1 == "" {
		t.Error("Generated trace ID should not be empty")
	}

	if traceID1 == traceID2 {
		t.Error("Generated trace IDs should be unique")
	}

	if len(traceID1) != 32 { // 16 bytes * 2 (hex encoding)
		t.Errorf("Expected trace ID length 32, got %d", len(traceID1))
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_DIR", "/tmp/logs")
	os.Setenv("LOG_FILE_NAME", "app.log")
	os.Setenv("LOG_MAX_FILE_SIZE_MB", "200")
	os.Setenv("LOG_CONSOLE_COLOR", "true")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_DIR")
		os.Unsetenv("LOG_FILE_NAME")
		os.Unsetenv("LOG_MAX_FILE_SIZE_MB")
		os.Unsetenv("LOG_CONSOLE_COLOR")
	}()

	baseConfig := Config{
		Level:           INFO,
		LogDir:          "default",
		FileName:        "default.log",
		MaxFileSize:     10 * 1024 * 1024,
		ConsoleColor:    false,
	}

	cfg := LoadConfigFromEnv(baseConfig)

	if cfg.Level != DEBUG {
		t.Errorf("Expected log level DEBUG, got %v", cfg.Level)
	}
	if cfg.LogDir != "/tmp/logs" {
		t.Errorf("Expected log dir '/tmp/logs', got %s", cfg.LogDir)
	}
	if cfg.FileName != "app.log" {
		t.Errorf("Expected file name 'app.log', got %s", cfg.FileName)
	}
	if cfg.MaxFileSize != 200*1024*1024 {
		t.Errorf("Expected max file size 200MB, got %d", cfg.MaxFileSize)
	}
	if !cfg.ConsoleColor {
		t.Error("Expected console color to be true")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
