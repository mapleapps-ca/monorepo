// cloud/backend/internal/maplefile/interface/http/collection/find_by_parent.go
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

type FindCollectionsByParentHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_collection.FindCollectionsByParentService
	middleware middleware.Middleware
}

func NewFindCollectionsByParentHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_collection.FindCollectionsByParentService,
	middleware middleware.Middleware,
) *FindCollectionsByParentHTTPHandler {
	logger = logger.Named("FindCollectionsByParentHTTPHandler")
	return &FindCollectionsByParentHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
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
	parentID, err := primitive.ObjectIDFromHex(parentIDStr)
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
			h.logger.Error("failed to find collections by parent",
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
		resp := result.(*svc_collection.CollectionsResponseDTO)
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
