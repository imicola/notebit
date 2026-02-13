package logger

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MetricsHandler returns an HTTP handler for exposing logger metrics
// in Prometheus-compatible format
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if defaultLogger == nil {
			http.Error(w, "Logger not initialized", http.StatusInternalServerError)
			return
		}

		metrics := defaultLogger.GetMetrics()

		// Check Accept header for format preference
		accept := r.Header.Get("Accept")
		if accept == "application/json" {
			// JSON format
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(metrics)
			return
		}

		// Prometheus text format
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP logger_debug_total Total number of DEBUG logs\n")
		fmt.Fprintf(w, "# TYPE logger_debug_total counter\n")
		fmt.Fprintf(w, "logger_debug_total %d\n", metrics.DebugCount)

		fmt.Fprintf(w, "# HELP logger_info_total Total number of INFO logs\n")
		fmt.Fprintf(w, "# TYPE logger_info_total counter\n")
		fmt.Fprintf(w, "logger_info_total %d\n", metrics.InfoCount)

		fmt.Fprintf(w, "# HELP logger_warn_total Total number of WARN logs\n")
		fmt.Fprintf(w, "# TYPE logger_warn_total counter\n")
		fmt.Fprintf(w, "logger_warn_total %d\n", metrics.WarnCount)

		fmt.Fprintf(w, "# HELP logger_error_total Total number of ERROR logs\n")
		fmt.Fprintf(w, "# TYPE logger_error_total counter\n")
		fmt.Fprintf(w, "logger_error_total %d\n", metrics.ErrorCount)

		fmt.Fprintf(w, "# HELP logger_fatal_total Total number of FATAL logs\n")
		fmt.Fprintf(w, "# TYPE logger_fatal_total counter\n")
		fmt.Fprintf(w, "logger_fatal_total %d\n", metrics.FatalCount)

		fmt.Fprintf(w, "# HELP logger_total_logs Total number of logs processed\n")
		fmt.Fprintf(w, "# TYPE logger_total_logs counter\n")
		fmt.Fprintf(w, "logger_total_logs %d\n", metrics.TotalLogs)

		fmt.Fprintf(w, "# HELP logger_dropped_total Total number of dropped logs\n")
		fmt.Fprintf(w, "# TYPE logger_dropped_total counter\n")
		fmt.Fprintf(w, "logger_dropped_total %d\n", metrics.DroppedLogs)

		fmt.Fprintf(w, "# HELP logger_queue_length Current queue length\n")
		fmt.Fprintf(w, "# TYPE logger_queue_length gauge\n")
		fmt.Fprintf(w, "logger_queue_length %d\n", metrics.QueueLength)

		fmt.Fprintf(w, "# HELP logger_last_flush_latency_microseconds Last flush latency in microseconds\n")
		fmt.Fprintf(w, "# TYPE logger_last_flush_latency_microseconds gauge\n")
		fmt.Fprintf(w, "logger_last_flush_latency_microseconds %d\n", metrics.LastFlushLatency)

		fmt.Fprintf(w, "# HELP logger_avg_flush_latency_microseconds Average flush latency in microseconds\n")
		fmt.Fprintf(w, "# TYPE logger_avg_flush_latency_microseconds gauge\n")
		fmt.Fprintf(w, "logger_avg_flush_latency_microseconds %d\n", metrics.AvgFlushLatency)

		fmt.Fprintf(w, "# HELP logger_max_flush_latency_microseconds Maximum flush latency in microseconds\n")
		fmt.Fprintf(w, "# TYPE logger_max_flush_latency_microseconds gauge\n")
		fmt.Fprintf(w, "logger_max_flush_latency_microseconds %d\n", metrics.MaxFlushLatency)

		fmt.Fprintf(w, "# HELP logger_batch_count_total Total number of batches processed\n")
		fmt.Fprintf(w, "# TYPE logger_batch_count_total counter\n")
		fmt.Fprintf(w, "logger_batch_count_total %d\n", metrics.BatchCount)

		fmt.Fprintf(w, "# HELP logger_avg_batch_size Average batch size\n")
		fmt.Fprintf(w, "# TYPE logger_avg_batch_size gauge\n")
		fmt.Fprintf(w, "logger_avg_batch_size %d\n", metrics.AvgBatchSize)
	}
}

// RegisterMetricsEndpoint registers the /metrics/log endpoint on the provided mux
func RegisterMetricsEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/metrics/log", MetricsHandler())
}
