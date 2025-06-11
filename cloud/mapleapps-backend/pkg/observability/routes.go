// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability/routes.go
package observability

import (
	"net/http"

	"go.uber.org/zap"
)

// HealthRoute provides detailed health check endpoint
type HealthRoute struct {
	checker *HealthChecker
	logger  *zap.Logger
}

func NewHealthRoute(checker *HealthChecker, logger *zap.Logger) *HealthRoute {
	return &HealthRoute{
		checker: checker,
		logger:  logger,
	}
}

func (h *HealthRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.checker.HealthHandler()(w, r)
}

func (h *HealthRoute) Pattern() string {
	return "/health"
}

// ReadinessRoute provides readiness probe endpoint
type ReadinessRoute struct {
	checker *HealthChecker
	logger  *zap.Logger
}

func NewReadinessRoute(checker *HealthChecker, logger *zap.Logger) *ReadinessRoute {
	return &ReadinessRoute{
		checker: checker,
		logger:  logger,
	}
}

func (r *ReadinessRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.checker.ReadinessHandler()(w, req)
}

func (r *ReadinessRoute) Pattern() string {
	return "/health/ready"
}

// LivenessRoute provides liveness probe endpoint
type LivenessRoute struct {
	checker *HealthChecker
	logger  *zap.Logger
}

func NewLivenessRoute(checker *HealthChecker, logger *zap.Logger) *LivenessRoute {
	return &LivenessRoute{
		checker: checker,
		logger:  logger,
	}
}

func (l *LivenessRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.checker.LivenessHandler()(w, r)
}

func (l *LivenessRoute) Pattern() string {
	return "/health/live"
}

// MetricsRoute provides metrics endpoint
type MetricsRoute struct {
	server *MetricsServer
	logger *zap.Logger
}

func NewMetricsRoute(server *MetricsServer, logger *zap.Logger) *MetricsRoute {
	return &MetricsRoute{
		server: server,
		logger: logger,
	}
}

func (m *MetricsRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.server.Handler()(w, r)
}

func (m *MetricsRoute) Pattern() string {
	return "/metrics"
}
