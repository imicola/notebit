package logger

import (
	"time"
)

// Level defines the log level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Config holds the configuration for the logger
type Config struct {
	Level           Level  // Minimum log level
	LogDir          string // Directory to store log files
	FileName        string // Base file name (e.g., "app.log")
	MaxFileSize     int64  // Maximum size in bytes before rotation
	MaxBackups      int    // Maximum number of backup files to keep (archiving strategy)
	ConsoleOutput   bool   // Whether to also output to console
	AsyncBufferSize int    // Size of the asynchronous buffer
}

// LogEntry represents a single log message
type LogEntry struct {
	Time       time.Time
	Level      Level
	ThreadID   uint64
	ClassName  string // File name or package name
	MethodName string // Function name
	Message    string
	Line       int
}
