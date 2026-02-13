package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoggerLevels(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		Level:           INFO,
		LogDir:          tmpDir,
		FileName:        "test_levels.log",
		MaxFileSize:     1024,
		MaxBackups:      1,
		ConsoleOutput:   false,
		AsyncBufferSize: 10,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	l.Debug("Debug message") // Should be ignored
	l.Info("Info message")   // Should be logged
	l.Warn("Warn message")   // Should be logged

	// Allow some time for async processing
	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(filepath.Join(tmpDir, "test_levels.log"))
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	logs := string(content)

	if strings.Contains(logs, "Debug message") {
		t.Error("Debug message should not be present")
	}
	if !strings.Contains(logs, "Info message") {
		t.Error("Info message should be present")
	}
	if !strings.Contains(logs, "Warn message") {
		t.Error("Warn message should be present")
	}
}

func TestLoggerRotation(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		Level:           DEBUG,
		LogDir:          tmpDir,
		FileName:        "test_rotation.log",
		MaxFileSize:     50, // Very small size to trigger rotation
		MaxBackups:      3,
		ConsoleOutput:   false,
		AsyncBufferSize: 10,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	// Write enough to trigger rotation
	for i := 0; i < 5; i++ {
		l.Info("Log message %d which is long enough to fill the file", i)
		time.Sleep(10 * time.Millisecond) // Ensure timestamps differ slightly if needed
	}

	// Allow async processing
	time.Sleep(200 * time.Millisecond)

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to list dir: %v", err)
	}

	logCount := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "test_rotation") {
			logCount++
		}
	}

	if logCount < 2 {
		t.Errorf("Expected at least 2 log files (rotated), got %d", logCount)
	}
}

func TestLoggerConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		Level:           DEBUG,
		LogDir:          tmpDir,
		FileName:        "test_concurrent.log",
		MaxFileSize:     1024 * 1024,
		MaxBackups:      1,
		ConsoleOutput:   false,
		AsyncBufferSize: 1000,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	var wg sync.WaitGroup
	numRoutines := 10
	numLogs := 100

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogs; j++ {
				l.Info("Routine %d Log %d", id, j)
			}
		}(i)
	}

	wg.Wait()
	// Allow async processing to finish
	time.Sleep(500 * time.Millisecond)

	// Check line count
	f, err := os.Open(filepath.Join(tmpDir, "test_concurrent.log"))
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}

	expected := numRoutines * numLogs
	if count != expected {
		t.Errorf("Expected %d logs, got %d", expected, count)
	}
}

func TestDynamicLevel(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		Level:           INFO,
		LogDir:          tmpDir,
		FileName:        "test_dynamic.log",
		MaxFileSize:     1024 * 1024,
		ConsoleOutput:   false,
		AsyncBufferSize: 10,
	}

	l, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	l.Debug("Initial Debug") // Ignored
	l.SetLevel(DEBUG)
	l.Debug("New Debug") // Logged

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(filepath.Join(tmpDir, "test_dynamic.log"))
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	logs := string(content)

	if strings.Contains(logs, "Initial Debug") {
		t.Error("Initial Debug should be ignored")
	}
	if !strings.Contains(logs, "New Debug") {
		t.Error("New Debug should be logged")
	}
}
