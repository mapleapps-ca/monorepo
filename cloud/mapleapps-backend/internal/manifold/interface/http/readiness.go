// monorepo/cloud/mapleapps-backend/internal/manifold/interface/http/readiness.go
package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability"
)

// GetReadinessHTTPHandler provides readiness probe endpoint
type GetReadinessHTTPHandler struct {
	log           *zap.Logger
	healthChecker *observability.HealthChecker
}

func NewGetReadinessHTTPHandler(
	log *zap.Logger,
	healthChecker *observability.HealthChecker,
) *GetReadinessHTTPHandler {
	return &GetReadinessHTTPHandler{
		log:           log,
		healthChecker: healthChecker,
	}
}

// ServeHTTP handles readiness probe requests
func (h *GetReadinessHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Delegate to the readiness handler from observability package
	h.healthChecker.ReadinessHandler()(w, r)
}

func (*GetReadinessHTTPHandler) Pattern() string {
	return "/health/ready"
}
