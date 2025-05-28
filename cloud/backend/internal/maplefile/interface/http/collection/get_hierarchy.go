// cloud/backend/internal/maplefile/interface/http/collection/get_hierarchy.go
package collection

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
	svc_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetCollectionHierarchyHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_collection.GetCollectionHierarchyService
	middleware middleware.Middleware
}

func NewGetCollectionHierarchyHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_collection.GetCollectionHierarchyService,
	middleware middleware.Middleware,
) *GetCollectionHierarchyHTTPHandler {
	logger = logger.Named("GetCollectionHierarchyHTTPHandler")
	return &GetCollectionHierarchyHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*GetCollectionHierarchyHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/collections/{collection_id}/hierarchy"
}

func (h *GetCollectionHierarchyHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *GetCollectionHierarchyHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
	rootID, err := primitive.ObjectIDFromHex(collectionIDStr)
	if err != nil {
		h.logger.Error("invalid collection ID format",
			zap.String("collection_id", collectionIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("collection_id", "Invalid collection ID format"))
		return
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
		response, err := h.service.Execute(sessCtx, rootID)
		if err != nil {
			h.logger.Error("failed to get collection hierarchy",
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
		resp := result.(*svc_collection.CollectionResponseDTO)
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
