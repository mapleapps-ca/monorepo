// monorepo/cloud/backend/internal/maplefile/interface/http/file/list_by_collection.go
package file

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListFilesByCollectionHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_file.ListFilesByCollectionService
	middleware middleware.Middleware
}

func NewListFilesByCollectionHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_file.ListFilesByCollectionService,
	middleware middleware.Middleware,
) *ListFilesByCollectionHTTPHandler {
	logger = logger.Named("ListFilesByCollectionHTTPHandler")
	return &ListFilesByCollectionHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*ListFilesByCollectionHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/collections/{collection_id}/files"
}

func (h *ListFilesByCollectionHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *ListFilesByCollectionHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Extract collection ID from URL parameters
	collectionIDStr := r.PathValue("collection_id")
	if collectionIDStr == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required"))
		return
	}

	// Convert string ID to ObjectID
	collectionID, err := gocql.ParseUUID(collectionIDStr)
	if err != nil {
		h.logger.Error("invalid collection ID format",
			zap.String("collection_id", collectionIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("collection_id", "Invalid collection ID format"))
		return
	}

	// Create request DTO
	req := &svc_file.ListFilesByCollectionRequestDTO{
		CollectionID: collectionID,
	}

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
