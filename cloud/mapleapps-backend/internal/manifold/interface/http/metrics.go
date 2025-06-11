// cloud/mapleapps-backend/internal/manifold/interface/http/metrics.go
package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/observability"
)

// GetMetricsHTTPHandler provides metrics endpoint
type GetMetricsHTTPHandler struct {
	log           *zap.Logger
	metricsServer *observability.MetricsServer
}

func NewGetMetricsHTTPHandler(
	log *zap.Logger,
	metricsServer *observability.MetricsServer,
) *GetMetricsHTTPHandler {
	return &GetMetricsHTTPHandler{
		log:           log,
		metricsServer: metricsServer,
	}
}

// ServeHTTP handles metrics requests
func (h *GetMetricsHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Delegate to the metrics handler from observability package
	h.metricsServer.Handler()(w, r)
}

func (*GetMetricsHTTPHandler) Pattern() string {
	return "/metrics"
}
