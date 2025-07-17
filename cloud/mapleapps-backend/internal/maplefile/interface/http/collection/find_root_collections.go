// monorepo/cloud/backend/internal/maplefile/interface/http/collection/find_root_collections.go
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

type FindRootCollectionsHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_collection.FindRootCollectionsService
	middleware middleware.Middleware
}

func NewFindRootCollectionsHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_collection.FindRootCollectionsService,
	middleware middleware.Middleware,
) *FindRootCollectionsHTTPHandler {
	logger = logger.Named("FindRootCollectionsHTTPHandler")
	return &FindRootCollectionsHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*FindRootCollectionsHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/collections/root"
}

func (h *FindRootCollectionsHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *FindRootCollectionsHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
