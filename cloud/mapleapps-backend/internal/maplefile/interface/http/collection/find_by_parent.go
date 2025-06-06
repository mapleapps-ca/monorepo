// cloud/backend/internal/maplefile/interface/http/collection/find_by_parent.go
package collection

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FindCollectionsByParentHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_collection.FindCollectionsByParentService
	middleware middleware.Middleware
}

func NewFindCollectionsByParentHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_collection.FindCollectionsByParentService,
	middleware middleware.Middleware,
) *FindCollectionsByParentHTTPHandler {
	logger = logger.Named("FindCollectionsByParentHTTPHandler")
	return &FindCollectionsByParentHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*FindCollectionsByParentHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/collections-by-parent/{parent_id}"
}

func (h *FindCollectionsByParentHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *FindCollectionsByParentHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Extract parent ID from URL parameters
	parentIDStr := r.PathValue("parent_id")
	if parentIDStr == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("parent_id", "Parent ID is required"))
		return
	}

	// Convert string ID to ObjectID
	parentID, err := gocql.ParseUUID(parentIDStr)
	if err != nil {
		h.logger.Error("invalid parent ID format",
			zap.String("parent_id", parentIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("parent_id", "Invalid parent ID format"))
		return
	}

	// Create request DTO
	req := &svc_collection.FindByParentRequestDTO{
		ParentID: parentID,
	}

	// Call service
	resp, err := h.service.Execute(ctx, req)
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
