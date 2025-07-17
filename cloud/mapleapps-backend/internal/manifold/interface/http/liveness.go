// monorepo/cloud/mapleapps-backend/internal/manifold/interface/http/liveness.go
package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability"
)

// GetLivenessHTTPHandler provides liveness probe endpoint
type GetLivenessHTTPHandler struct {
	log           *zap.Logger
	healthChecker *observability.HealthChecker
}

func NewGetLivenessHTTPHandler(
	log *zap.Logger,
	healthChecker *observability.HealthChecker,
) *GetLivenessHTTPHandler {
	return &GetLivenessHTTPHandler{
		log:           log,
		healthChecker: healthChecker,
	}
}

// ServeHTTP handles liveness probe requests
func (h *GetLivenessHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Delegate to the liveness handler from observability package
	h.healthChecker.LivenessHandler()(w, r)
}

func (*GetLivenessHTTPHandler) Pattern() string {
	return "/health/live"
}
