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

// Color returns ANSI color code for the log level
func (l Level) Color() string {
	switch l {
	case DEBUG:
		return "\033[36m" // Cyan
	case INFO:
		return "\033[32m" // Green
	case WARN:
		return "\033[33m" // Yellow
	case ERROR:
		return "\033[31m" // Red
	case FATAL:
		return "\033[35m" // Magenta
	default:
		return "\033[0m" // Reset
	}
}

const ColorReset = "\033[0m"

// Config holds the configuration for the logger
type Config struct {
	Level           Level  // Minimum log level
	LogDir          string // Directory to store log files
	FileName        string // Base file name (e.g., "app.log")
	MaxFileSize     int64  // Maximum size in bytes before rotation (default: 100MB)
	MaxBackups      int    // Maximum number of backup files to keep (default: 15 days)
	ConsoleOutput   bool   // Whether to also output to console
	ConsoleColor    bool   // Whether to use colors in console output
	AsyncBufferSize int    // Size of the asynchronous buffer (default: 1000)
	BatchSize       int    // Number of logs to batch before flushing (default: 10)
	FlushInterval   int    // Flush interval in milliseconds (default: 100ms)
	KafkaEnabled    bool   // Whether to send logs to Kafka
	KafkaBrokers    []string // Kafka broker addresses
	KafkaTopic      string   // Kafka topic name (default: "app-logs")
}

// LogEntry represents a single log message
type LogEntry struct {
	Time       time.Time
	Level      Level
	ThreadID   uint64
	TraceID    string            // Request trace ID for distributed tracing
	ClassName  string            // File name or package name
	MethodName string            // Function name
	Message    string
	Line       int
	Fields     map[string]interface{} // Additional context fields (userID, orderID, etc.)
	Duration   time.Duration     // Execution duration for performance tracking
}
