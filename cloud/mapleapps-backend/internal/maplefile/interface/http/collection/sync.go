// cloud/backend/internal/maplefile/interface/http/collection/sync.go
package collection

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_sync "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CollectionSyncHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	repository dom_collection.CollectionRepository
	middleware middleware.Middleware
}

func NewCollectionSyncHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	repository dom_collection.CollectionRepository,
	middleware middleware.Middleware,
) *CollectionSyncHTTPHandler {
	logger = logger.Named("CollectionSyncHTTPHandler")
	return &CollectionSyncHTTPHandler{
		config:     config,
		logger:     logger,
		repository: repository,
		middleware: middleware,
	}
}

func (*CollectionSyncHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/sync/collections"
}

func (h *CollectionSyncHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *CollectionSyncHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		h.logger.Error("Failed getting user ID from context")
		httperror.ResponseError(w, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error"))
		return
	}

	// Parse query parameters
	queryParams := r.URL.Query()

	// Parse limit parameter (default: 1000, max: 5000)
	limit := int64(1000)
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			if parsedLimit > 0 && parsedLimit <= 5000 {
				limit = parsedLimit
			} else {
				h.logger.Warn("Invalid limit parameter, using default",
					zap.String("limit", limitStr),
					zap.Int64("default", limit))
			}
		} else {
			h.logger.Warn("Failed to parse limit parameter, using default",
				zap.String("limit", limitStr),
				zap.Error(err))
		}
	}

	// Parse cursor parameter
	var cursor *dom_sync.CollectionSyncCursor
	if cursorStr := queryParams.Get("cursor"); cursorStr != "" {
		var parsedCursor dom_sync.CollectionSyncCursor
		if err := json.Unmarshal([]byte(cursorStr), &parsedCursor); err != nil {
			h.logger.Error("Failed to parse cursor parameter",
				zap.String("cursor", cursorStr),
				zap.Error(err))
			httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("cursor", "Invalid cursor format"))
			return
		}
		cursor = &parsedCursor
	}

	h.logger.Debug("Processing collection sync request",
		zap.Any("user_id", userID),
		zap.Int64("limit", limit),
		zap.Any("cursor", cursor))

	// Call repository to get sync data
	response, err := h.repository.GetCollectionSyncData(ctx, userID, cursor, limit)
	if err != nil {
		h.logger.Error("Failed to get collection sync data",
			zap.Any("user_id", userID),
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Encode and return response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode collection sync response",
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	h.logger.Info("Successfully served collection sync data",
		zap.Any("user_id", userID),
		zap.Int("collections_count", len(response.Collections)),
		zap.Bool("has_more", response.HasMore))
}
