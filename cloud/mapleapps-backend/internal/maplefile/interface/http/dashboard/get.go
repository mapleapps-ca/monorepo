// cloud/mapleapps-backend/internal/maplefile/interface/http/dashboard/get.go
package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_dashboard "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/dashboard"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetDashboardHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_dashboard.GetDashboardService
	middleware middleware.Middleware
}

func NewGetDashboardHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_dashboard.GetDashboardService,
	middleware middleware.Middleware,
) *GetDashboardHTTPHandler {
	logger = logger.With(zap.String("module", "maplefile"))
	logger = logger.Named("GetDashboardHTTPHandler")
	return &GetDashboardHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GetDashboardHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/dashboard"
}

func (h *GetDashboardHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *GetDashboardHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	//
	// STEP 1: Execute service
	//
	resp, err := h.service.Execute(ctx)
	if err != nil {
		h.logger.Error("Failed to get dashboard data",
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	//
	// STEP 2: Encode and return response
	//
	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("Failed to encode dashboard response",
				zap.Error(err))
			httperror.ResponseError(w, err)
			return
		}
	} else {
		err := errors.New("no dashboard data available")
		h.logger.Error("No dashboard data returned from service")
		httperror.ResponseError(w, err)
		return
	}

	h.logger.Debug("Dashboard data successfully returned")
}
