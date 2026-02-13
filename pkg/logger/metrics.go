package logger

import (
	"sync/atomic"
	"time"
)

// Metrics holds logger performance metrics
type Metrics struct {
	// Log counts by level
	debugCount  atomic.Uint64
	infoCount   atomic.Uint64
	warnCount   atomic.Uint64
	errorCount  atomic.Uint64
	fatalCount  atomic.Uint64
	
	// Performance metrics
	queueLength atomic.Int64  // Current queue length
	droppedLogs atomic.Uint64 // Total dropped logs
	totalLogs   atomic.Uint64 // Total logs processed
	
	// Latency tracking (in microseconds)
	lastFlushLatency atomic.Int64
	avgFlushLatency  atomic.Int64
	maxFlushLatency  atomic.Int64
	
	// Batch metrics
	batchCount atomic.Uint64
	avgBatchSize atomic.Int64
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncrementLevel increments the counter for a specific log level
func (m *Metrics) IncrementLevel(level Level) {
	switch level {
	case DEBUG:
		m.debugCount.Add(1)
	case INFO:
		m.infoCount.Add(1)
	case WARN:
		m.warnCount.Add(1)
	case ERROR:
		m.errorCount.Add(1)
	case FATAL:
		m.fatalCount.Add(1)
	}
	m.totalLogs.Add(1)
}

// IncrementDropped increments the dropped logs counter
func (m *Metrics) IncrementDropped() {
	m.droppedLogs.Add(1)
}

// UpdateQueueLength updates the current queue length
func (m *Metrics) UpdateQueueLength(length int) {
	m.queueLength.Store(int64(length))
}

// RecordFlushLatency records a flush operation latency
func (m *Metrics) RecordFlushLatency(duration time.Duration) {
	micros := duration.Microseconds()
	m.lastFlushLatency.Store(micros)
	
	// Update max
	for {
		oldMax := m.maxFlushLatency.Load()
		if micros <= oldMax {
			break
		}
		if m.maxFlushLatency.CompareAndSwap(oldMax, micros) {
			break
		}
	}
	
	// Update average (simple moving average)
	oldAvg := m.avgFlushLatency.Load()
	newAvg := (oldAvg*9 + micros) / 10 // Exponential moving average
	m.avgFlushLatency.Store(newAvg)
}

// RecordBatch records batch processing metrics
func (m *Metrics) RecordBatch(size int) {
	m.batchCount.Add(1)
	oldAvg := m.avgBatchSize.Load()
	newAvg := (oldAvg*9 + int64(size)) / 10
	m.avgBatchSize.Store(newAvg)
}

// GetSnapshot returns a snapshot of current metrics
type MetricsSnapshot struct {
	DebugCount       uint64
	InfoCount        uint64
	WarnCount        uint64
	ErrorCount       uint64
	FatalCount       uint64
	TotalLogs        uint64
	DroppedLogs      uint64
	QueueLength      int64
	LastFlushLatency int64 // microseconds
	AvgFlushLatency  int64 // microseconds
	MaxFlushLatency  int64 // microseconds
	BatchCount       uint64
	AvgBatchSize     int64
}

func (m *Metrics) GetSnapshot() MetricsSnapshot {
	return MetricsSnapshot{
		DebugCount:       m.debugCount.Load(),
		InfoCount:        m.infoCount.Load(),
		WarnCount:        m.warnCount.Load(),
		ErrorCount:       m.errorCount.Load(),
		FatalCount:       m.fatalCount.Load(),
		TotalLogs:        m.totalLogs.Load(),
		DroppedLogs:      m.droppedLogs.Load(),
		QueueLength:      m.queueLength.Load(),
		LastFlushLatency: m.lastFlushLatency.Load(),
		AvgFlushLatency:  m.avgFlushLatency.Load(),
		MaxFlushLatency:  m.maxFlushLatency.Load(),
		BatchCount:       m.batchCount.Load(),
		AvgBatchSize:     m.avgBatchSize.Load(),
	}
}

// Reset resets all metrics to zero
func (m *Metrics) Reset() {
	m.debugCount.Store(0)
	m.infoCount.Store(0)
	m.warnCount.Store(0)
	m.errorCount.Store(0)
	m.fatalCount.Store(0)
	m.totalLogs.Store(0)
	m.droppedLogs.Store(0)
	m.queueLength.Store(0)
	m.lastFlushLatency.Store(0)
	m.avgFlushLatency.Store(0)
	m.maxFlushLatency.Store(0)
	m.batchCount.Store(0)
	m.avgBatchSize.Store(0)
}
