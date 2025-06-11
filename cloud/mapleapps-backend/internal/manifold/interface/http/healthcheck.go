// cloud/mapleapps-backend/internal/manifold/interface/http/healthcheck.go
package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability"
)

// GetHealthCheckHTTPHandler provides comprehensive health check endpoint using observability package
type GetHealthCheckHTTPHandler struct {
	log           *zap.Logger
	healthChecker *observability.HealthChecker
}

func NewGetHealthCheckHTTPHandler(
	log *zap.Logger,
	healthChecker *observability.HealthChecker,
) *GetHealthCheckHTTPHandler {
	return &GetHealthCheckHTTPHandler{
		log:           log,
		healthChecker: healthChecker,
	}
}

// ServeHTTP handles comprehensive health check requests using the observability system
func (h *GetHealthCheckHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Delegate to the comprehensive health checker from observability package
	h.healthChecker.HealthHandler()(w, r)
}

func (*GetHealthCheckHTTPHandler) Pattern() string {
	return "/healthcheck"
}
