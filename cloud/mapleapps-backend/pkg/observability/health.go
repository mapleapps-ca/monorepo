// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability/health.go
package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/cache/twotiercache"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/storage/object/s3"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Duration  string       `json:"duration,omitempty"`
	Component string       `json:"component"`
	Details   interface{}  `json:"details,omitempty"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status    HealthStatus                 `json:"status"`
	Timestamp time.Time                    `json:"timestamp"`
	Services  map[string]HealthCheckResult `json:"services"`
	Version   string                       `json:"version"`
	Uptime    string                       `json:"uptime"`
}

// HealthChecker manages health checks for various components
type HealthChecker struct {
	checks    map[string]HealthCheck
	mu        sync.RWMutex
	logger    *zap.Logger
	startTime time.Time
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) HealthCheckResult

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		checks:    make(map[string]HealthCheck),
		logger:    logger,
		startTime: time.Now(),
	}
}

// RegisterCheck registers a health check for a service
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// CheckHealth performs all registered health checks
func (hc *HealthChecker) CheckHealth(ctx context.Context) HealthResponse {
	hc.mu.RLock()
	checks := make(map[string]HealthCheck, len(hc.checks))
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mu.RUnlock()

	results := make(map[string]HealthCheckResult)
	overallStatus := HealthStatusHealthy

	// Run health checks with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for name, check := range checks {
		start := time.Now()
		result := check(checkCtx)
		result.Duration = time.Since(start).String()

		results[name] = result

		// Determine overall status
		if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  results,
		Version:   "1.0.0", // Could be injected
		Uptime:    time.Since(hc.startTime).String(),
	}
}

// HealthHandler creates an HTTP handler for health checks
func (hc *HealthChecker) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		health := hc.CheckHealth(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Set appropriate status code
		switch health.Status {
		case HealthStatusHealthy:
			w.WriteHeader(http.StatusOK)
		case HealthStatusDegraded:
			w.WriteHeader(http.StatusOK) // 200 but degraded
		case HealthStatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(health); err != nil {
			hc.logger.Error("Failed to encode health response", zap.Error(err))
		}
	}
}

// ReadinessHandler creates a simple readiness probe
func (hc *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		health := hc.CheckHealth(ctx)

		// For readiness, we're more strict - any unhealthy component means not ready
		if health.Status == HealthStatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NOT READY"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	}
}

// LivenessHandler creates a simple liveness probe
func (hc *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// For liveness, we just check if the service can respond
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ALIVE"))
	}
}

// CassandraHealthCheck creates a health check for Cassandra database connectivity
func CassandraHealthCheck(session *gocql.Session, logger *zap.Logger) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		start := time.Now()

		// Check if session is nil
		if session == nil {
			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Cassandra session is nil",
				Timestamp: time.Now(),
				Component: "cassandra",
				Details:   map[string]interface{}{"error": "session_nil"},
			}
		}

		// Try to execute a simple query with context
		var result string
		query := session.Query("SELECT uuid() FROM system.local")

		// Create a channel to handle the query execution
		done := make(chan error, 1)
		go func() {
			done <- query.Scan(&result)
		}()

		// Wait for either completion or context cancellation
		select {
		case err := <-done:
			duration := time.Since(start)

			if err != nil {
				logger.Warn("Cassandra health check failed",
					zap.Error(err),
					zap.Duration("duration", duration))

				return HealthCheckResult{
					Status:    HealthStatusUnhealthy,
					Message:   "Cassandra query failed: " + err.Error(),
					Timestamp: time.Now(),
					Component: "cassandra",
					Details: map[string]interface{}{
						"error":    err.Error(),
						"duration": duration.String(),
					},
				}
			}

			return HealthCheckResult{
				Status:    HealthStatusHealthy,
				Message:   "Cassandra connection healthy",
				Timestamp: time.Now(),
				Component: "cassandra",
				Details: map[string]interface{}{
					"query_result": result,
					"duration":     duration.String(),
				},
			}

		case <-ctx.Done():
			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Cassandra health check timed out",
				Timestamp: time.Now(),
				Component: "cassandra",
				Details: map[string]interface{}{
					"error":    "timeout",
					"duration": time.Since(start).String(),
				},
			}
		}
	}
}

// TwoTierCacheHealthCheck creates a health check for the two-tier cache system
func TwoTierCacheHealthCheck(cache twotiercache.TwoTierCacher, logger *zap.Logger) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		start := time.Now()

		if cache == nil {
			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Cache instance is nil",
				Timestamp: time.Now(),
				Component: "two_tier_cache",
				Details:   map[string]interface{}{"error": "cache_nil"},
			}
		}

		// Test cache functionality with a health check key
		healthKey := "health_check_" + time.Now().Format("20060102150405")
		testValue := []byte("health_check_value")

		// Test Set operation
		if err := cache.Set(ctx, healthKey, testValue); err != nil {
			duration := time.Since(start)
			logger.Warn("Cache health check SET failed",
				zap.Error(err),
				zap.Duration("duration", duration))

			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Cache SET operation failed: " + err.Error(),
				Timestamp: time.Now(),
				Component: "two_tier_cache",
				Details: map[string]interface{}{
					"error":     err.Error(),
					"operation": "set",
					"duration":  duration.String(),
				},
			}
		}

		// Test Get operation
		retrievedValue, err := cache.Get(ctx, healthKey)
		if err != nil {
			duration := time.Since(start)
			logger.Warn("Cache health check GET failed",
				zap.Error(err),
				zap.Duration("duration", duration))

			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "Cache GET operation failed: " + err.Error(),
				Timestamp: time.Now(),
				Component: "two_tier_cache",
				Details: map[string]interface{}{
					"error":     err.Error(),
					"operation": "get",
					"duration":  duration.String(),
				},
			}
		}

		// Verify the value
		if string(retrievedValue) != string(testValue) {
			duration := time.Since(start)
			return HealthCheckResult{
				Status:    HealthStatusDegraded,
				Message:   "Cache value mismatch",
				Timestamp: time.Now(),
				Component: "two_tier_cache",
				Details: map[string]interface{}{
					"expected": string(testValue),
					"actual":   string(retrievedValue),
					"duration": duration.String(),
				},
			}
		}

		// Clean up test key
		_ = cache.Delete(ctx, healthKey)

		duration := time.Since(start)
		return HealthCheckResult{
			Status:    HealthStatusHealthy,
			Message:   "Two-tier cache healthy",
			Timestamp: time.Now(),
			Component: "two_tier_cache",
			Details: map[string]interface{}{
				"operations_tested": []string{"set", "get", "delete"},
				"duration":          duration.String(),
			},
		}
	}
}

// S3HealthCheck creates a health check for S3 object storage
func S3HealthCheck(s3Storage s3.S3ObjectStorage, logger *zap.Logger) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		start := time.Now()

		if s3Storage == nil {
			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "S3 storage instance is nil",
				Timestamp: time.Now(),
				Component: "s3_storage",
				Details:   map[string]interface{}{"error": "storage_nil"},
			}
		}

		// Test basic S3 connectivity by listing objects (lightweight operation)
		_, err := s3Storage.ListAllObjects(ctx)
		duration := time.Since(start)

		if err != nil {
			logger.Warn("S3 health check failed",
				zap.Error(err),
				zap.Duration("duration", duration))

			return HealthCheckResult{
				Status:    HealthStatusUnhealthy,
				Message:   "S3 connectivity failed: " + err.Error(),
				Timestamp: time.Now(),
				Component: "s3_storage",
				Details: map[string]interface{}{
					"error":     err.Error(),
					"operation": "list_objects",
					"duration":  duration.String(),
				},
			}
		}

		return HealthCheckResult{
			Status:    HealthStatusHealthy,
			Message:   "S3 storage healthy",
			Timestamp: time.Now(),
			Component: "s3_storage",
			Details: map[string]interface{}{
				"operation": "list_objects",
				"duration":  duration.String(),
			},
		}
	}
}

// registerRealHealthChecks registers health checks for actual infrastructure components
func registerRealHealthChecks(
	hc *HealthChecker,
	logger *zap.Logger,
	cassandraSession *gocql.Session,
	cache twotiercache.TwoTierCacher,
	s3Storage s3.S3ObjectStorage,
) {
	// Register Cassandra health check
	hc.RegisterCheck("cassandra", CassandraHealthCheck(cassandraSession, logger))

	// Register two-tier cache health check
	hc.RegisterCheck("cache", TwoTierCacheHealthCheck(cache, logger))

	// Register S3 storage health check
	hc.RegisterCheck("s3_storage", S3HealthCheck(s3Storage, logger))

	logger.Info("Real infrastructure health checks registered",
		zap.Strings("components", []string{"cassandra", "cache", "s3_storage"}))
}

// startObservabilityServer starts the observability HTTP server on a separate port
func startObservabilityServer(
	lc fx.Lifecycle,
	hc *HealthChecker,
	ms *MetricsServer,
	logger *zap.Logger,
) {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", hc.HealthHandler())
	mux.HandleFunc("/health/ready", hc.ReadinessHandler())
	mux.HandleFunc("/health/live", hc.LivenessHandler())

	// Metrics endpoint
	mux.Handle("/metrics", ms.Handler())

	server := &http.Server{
		Addr:         ":8080", // Separate port for observability
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting observability server on :8080")
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("Observability server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping observability server")
			return server.Shutdown(ctx)
		},
	})
}
