// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability/metrics.go
package observability

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"go.uber.org/zap"
)

// MetricsServer provides basic metrics endpoint
type MetricsServer struct {
	logger    *zap.Logger
	startTime time.Time
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		logger:    logger,
		startTime: time.Now(),
	}
}

// Handler returns an HTTP handler that serves basic metrics
func (ms *MetricsServer) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		metrics := ms.collectMetrics()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		for _, metric := range metrics {
			fmt.Fprintf(w, "%s\n", metric)
		}
	}
}

// collectMetrics collects basic application metrics
func (ms *MetricsServer) collectMetrics() []string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(ms.startTime).Seconds()

	metrics := []string{
		fmt.Sprintf("# HELP mapleapps_uptime_seconds Total uptime of the service in seconds"),
		fmt.Sprintf("# TYPE mapleapps_uptime_seconds counter"),
		fmt.Sprintf("mapleapps_uptime_seconds %.2f", uptime),

		fmt.Sprintf("# HELP mapleapps_memory_alloc_bytes Currently allocated memory in bytes"),
		fmt.Sprintf("# TYPE mapleapps_memory_alloc_bytes gauge"),
		fmt.Sprintf("mapleapps_memory_alloc_bytes %d", m.Alloc),

		fmt.Sprintf("# HELP mapleapps_memory_total_alloc_bytes Total allocated memory in bytes"),
		fmt.Sprintf("# TYPE mapleapps_memory_total_alloc_bytes counter"),
		fmt.Sprintf("mapleapps_memory_total_alloc_bytes %d", m.TotalAlloc),

		fmt.Sprintf("# HELP mapleapps_memory_sys_bytes Memory obtained from system in bytes"),
		fmt.Sprintf("# TYPE mapleapps_memory_sys_bytes gauge"),
		fmt.Sprintf("mapleapps_memory_sys_bytes %d", m.Sys),

		fmt.Sprintf("# HELP mapleapps_gc_runs_total Total number of GC runs"),
		fmt.Sprintf("# TYPE mapleapps_gc_runs_total counter"),
		fmt.Sprintf("mapleapps_gc_runs_total %d", m.NumGC),

		fmt.Sprintf("# HELP mapleapps_goroutines Current number of goroutines"),
		fmt.Sprintf("# TYPE mapleapps_goroutines gauge"),
		fmt.Sprintf("mapleapps_goroutines %d", runtime.NumGoroutine()),
	}

	return metrics
}

// RecordMetric records a custom metric (placeholder for future implementation)
func (ms *MetricsServer) RecordMetric(name string, value float64, labels map[string]string) {
	ms.logger.Debug("Recording metric",
		zap.String("name", name),
		zap.Float64("value", value),
		zap.Any("labels", labels),
	)
}
