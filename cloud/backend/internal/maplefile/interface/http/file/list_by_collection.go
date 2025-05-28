// cloud/backend/internal/maplefile/interface/http/file/list_by_collection.go
package file

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type ListFilesByCollectionHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.ListFilesByCollectionService
	middleware middleware.Middleware
}

func NewListFilesByCollectionHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.ListFilesByCollectionService,
	middleware middleware.Middleware,
) *ListFilesByCollectionHTTPHandler {
	logger = logger.Named("ListFilesByCollectionHTTPHandler")
	return &ListFilesByCollectionHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
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
	collectionID, err := primitive.ObjectIDFromHex(collectionIDStr)
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

	// Start the transaction
	session, err := h.dbClient.StartSession()
	if err != nil {
		h.logger.Error("start session error",
			zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx context.Context) (interface{}, error) {
		// Call service
		response, err := h.service.Execute(sessCtx, req)
		if err != nil {
			h.logger.Error("failed to list files by collection",
				zap.Any("error", err))
			return nil, err
		}
		return response, nil
	}

	// Start a transaction
	result, txErr := session.WithTransaction(ctx, transactionFunc)
	if txErr != nil {
		h.logger.Error("session failed error",
			zap.Any("error", txErr))
		httperror.ResponseError(w, txErr)
		return
	}

	// Encode response
	if result != nil {
		resp := result.(*svc_file.FilesResponseDTO)
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
