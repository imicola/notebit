package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	config     atomic.Value // Stores Config
	logChan    chan LogEntry
	writer     *FileWriter
	writerMu   sync.Mutex // Protects writer replacement
	wg         sync.WaitGroup
	isClosed   atomic.Bool
	consoleOut io.Writer
}

var defaultLogger *Logger
var once sync.Once

// New creates a new Logger instance
func New(cfg Config) (*Logger, error) {
	if cfg.AsyncBufferSize <= 0 {
		cfg.AsyncBufferSize = 1000 // Default buffer size
	}
	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = 10 * 1024 * 1024 // Default 10MB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 7 // Default 7 backups
	}

	fw, err := NewFileWriter(cfg)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		logChan:    make(chan LogEntry, cfg.AsyncBufferSize),
		writer:     fw,
		consoleOut: os.Stdout,
	}
	l.config.Store(cfg)

	l.wg.Add(1)
	go l.processLogs()

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

func (l *Logger) processLogs() {
	defer l.wg.Done()
	for entry := range l.logChan {
		l.writeEntry(entry)
	}
}

func (l *Logger) writeEntry(entry LogEntry) {
	// Format: YYYY-MM-DD HH:MM:SS.000 [LEVEL] [GID] [Class.Method:Line] - Message
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	
	msg := fmt.Sprintf("%s [%s] [%d] [%s.%s:%d] - %s\n",
		timestamp,
		entry.Level.String(),
		entry.ThreadID,
		entry.ClassName,
		entry.MethodName,
		entry.Line,
		entry.Message,
	)

	cfg := l.config.Load().(Config)

	// Write to file
	l.writerMu.Lock()
	if l.writer != nil {
		l.writer.Write([]byte(msg))
	}
	l.writerMu.Unlock()

	// Write to console if enabled
	if cfg.ConsoleOutput {
		fmt.Fprint(l.consoleOut, msg)
	}
}

func (l *Logger) log(level Level, msg string) {
	if l.isClosed.Load() {
		return
	}

	cfg := l.config.Load().(Config)
	if level < cfg.Level {
		return
	}

	// Capture caller info
	// Skip 2: log() -> Debug/Info/... -> Caller
	fileName, funcName, line := getCallerInfo(3)

	entry := LogEntry{
		Time:       time.Now(),
		Level:      level,
		ThreadID:   getGID(),
		ClassName:  fileName,
		MethodName: funcName,
		Line:       line,
		Message:    msg,
	}

	// Non-blocking write if buffer is full, or block? 
	// Requirement: "avoid blocking main program".
	// Implementation: Try send, if full, maybe print to stderr or drop? 
	// For critical logs (FATAL/ERROR), we might want to ensure delivery.
	// Let's use a select with default for non-blocking, but maybe with a small timeout?
	// Strictly non-blocking:
	select {
	case l.logChan <- entry:
	default:
		// Buffer full. Fallback to console or drop.
		// Writing to stderr to warn about dropped logs is good practice.
		fmt.Fprintf(os.Stderr, "Logger buffer full, dropping message: %s\n", msg)
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

// Dynamic Configuration

func (l *Logger) SetLevel(level Level) {
	cfg := l.config.Load().(Config)
	cfg.Level = level
	l.config.Store(cfg)
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

// Close gracefully shuts down the logger
func (l *Logger) Close() {
	if l.isClosed.CompareAndSwap(false, true) {
		close(l.logChan)
		l.wg.Wait()
		l.writerMu.Lock()
		if l.writer != nil {
			l.writer.Close()
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
