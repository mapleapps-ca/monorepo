// cloud/backend/internal/maplefile/interface/http/collection/list_by_user.go
package collection

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListUserCollectionsHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_collection.ListUserCollectionsService
	middleware middleware.Middleware
}

func NewListUserCollectionsHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_collection.ListUserCollectionsService,
	middleware middleware.Middleware,
) *ListUserCollectionsHTTPHandler {
	logger = logger.Named("ListUserCollectionsHTTPHandler")
	return &ListUserCollectionsHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*ListUserCollectionsHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/collections"
}

func (h *ListUserCollectionsHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *ListUserCollectionsHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	resp, err := h.service.Execute(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Encode response
	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("failed to encode response",
				zap.Any("error", err))
			httperror.ResponseError(w, err)
			return
		}
	} else {
		err := errors.New("no result")
		httperror.ResponseError(w, err)
		return
	}
}
