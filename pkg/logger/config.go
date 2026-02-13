package logger

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ConfigWatcher watches for configuration changes
type ConfigWatcher struct {
	logger         *Logger
	checkInterval  time.Duration
	stopChan       chan struct{}
	wg             sync.WaitGroup
	lastLogLevel   string
	lastKafkaState string
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(logger *Logger, checkInterval time.Duration) *ConfigWatcher {
	if checkInterval == 0 {
		checkInterval = 10 * time.Second // Default check interval
	}

	return &ConfigWatcher{
		logger:        logger,
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
	}
}

// Start begins watching for configuration changes
func (cw *ConfigWatcher) Start() {
	cw.wg.Add(1)
	go cw.watch()
}

// Stop stops the configuration watcher
func (cw *ConfigWatcher) Stop() {
	close(cw.stopChan)
	cw.wg.Wait()
}

func (cw *ConfigWatcher) watch() {
	defer cw.wg.Done()

	ticker := time.NewTicker(cw.checkInterval)
	defer ticker.Stop()

	// Initial load
	cw.applyConfigFromEnv()

	for {
		select {
		case <-ticker.C:
			cw.applyConfigFromEnv()
		case <-cw.stopChan:
			return
		}
	}
}

func (cw *ConfigWatcher) applyConfigFromEnv() {
	// Check LOG_LEVEL environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" && logLevel != cw.lastLogLevel {
		level := parseLogLevel(logLevel)
		if level >= 0 {
			cw.logger.SetLevel(level)
			Info("Log level updated from environment: %s", logLevel)
			cw.lastLogLevel = logLevel
		}
	}

	// Check KAFKA_ENABLED environment variable
	kafkaEnabled := os.Getenv("KAFKA_ENABLED")
	if kafkaEnabled != "" && kafkaEnabled != cw.lastKafkaState {
		// Note: Dynamically changing Kafka state requires more complex logic
		// For now, we just log the change
		Info("Kafka enabled state changed in environment: %s (requires restart)", kafkaEnabled)
		cw.lastKafkaState = kafkaEnabled
	}

	// Check ASYNC_BUFFER_SIZE environment variable
	bufferSize := os.Getenv("LOG_BUFFER_SIZE")
	if bufferSize != "" {
		if size, err := strconv.Atoi(bufferSize); err == nil && size > 0 {
			// Note: Buffer size cannot be changed after initialization
			// This is just for validation
			Debug("Log buffer size in environment: %d (requires restart)", size)
		}
	}
}

func parseLogLevel(level string) Level {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return -1 // Invalid level
	}
}

// LoadConfigFromEnv loads logging configuration from environment variables
func LoadConfigFromEnv(base Config) Config {
	cfg := base

	// Load log level
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		if level := parseLogLevel(logLevel); level >= 0 {
			cfg.Level = level
		}
	}

	// Load log directory
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		cfg.LogDir = logDir
	}

	// Load file name
	if fileName := os.Getenv("LOG_FILE_NAME"); fileName != "" {
		cfg.FileName = fileName
	}

	// Load max file size (in MB)
	if maxSizeStr := os.Getenv("LOG_MAX_FILE_SIZE_MB"); maxSizeStr != "" {
		if maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64); err == nil {
			cfg.MaxFileSize = maxSize * 1024 * 1024
		}
	}

	// Load max backups
	if maxBackupsStr := os.Getenv("LOG_MAX_BACKUPS"); maxBackupsStr != "" {
		if maxBackups, err := strconv.Atoi(maxBackupsStr); err == nil {
			cfg.MaxBackups = maxBackups
		}
	}

	// Load console output
	if consoleStr := os.Getenv("LOG_CONSOLE_OUTPUT"); consoleStr != "" {
		cfg.ConsoleOutput = strings.ToLower(consoleStr) == "true" || consoleStr == "1"
	}

	// Load console color
	if colorStr := os.Getenv("LOG_CONSOLE_COLOR"); colorStr != "" {
		cfg.ConsoleColor = strings.ToLower(colorStr) == "true" || colorStr == "1"
	}

	// Load async buffer size
	if bufferSizeStr := os.Getenv("LOG_BUFFER_SIZE"); bufferSizeStr != "" {
		if bufferSize, err := strconv.Atoi(bufferSizeStr); err == nil {
			cfg.AsyncBufferSize = bufferSize
		}
	}

	// Load batch size
	if batchSizeStr := os.Getenv("LOG_BATCH_SIZE"); batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil {
			cfg.BatchSize = batchSize
		}
	}

	// Load flush interval
	if flushIntervalStr := os.Getenv("LOG_FLUSH_INTERVAL_MS"); flushIntervalStr != "" {
		if flushInterval, err := strconv.Atoi(flushIntervalStr); err == nil {
			cfg.FlushInterval = flushInterval
		}
	}

	// Load Kafka configuration
	if kafkaEnabledStr := os.Getenv("KAFKA_ENABLED"); kafkaEnabledStr != "" {
		cfg.KafkaEnabled = strings.ToLower(kafkaEnabledStr) == "true" || kafkaEnabledStr == "1"
	}

	if kafkaBrokers := os.Getenv("KAFKA_BROKERS"); kafkaBrokers != "" {
		cfg.KafkaBrokers = strings.Split(kafkaBrokers, ",")
		for i := range cfg.KafkaBrokers {
			cfg.KafkaBrokers[i] = strings.TrimSpace(cfg.KafkaBrokers[i])
		}
	}

	if kafkaTopic := os.Getenv("KAFKA_TOPIC"); kafkaTopic != "" {
		cfg.KafkaTopic = kafkaTopic
	}

	return cfg
}
