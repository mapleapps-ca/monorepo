// cloud/backend/internal/maplefile/interface/http/file/sync.go
package file

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/middleware"
	file_service "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type FileSyncHTTPHandler struct {
	config          *config.Configuration
	logger          *zap.Logger
	fileSyncService file_service.GetFileSyncDataService // CHANGED: Use service instead of repository
	middleware      middleware.Middleware
}

func NewFileSyncHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	fileSyncService file_service.GetFileSyncDataService, // CHANGED: Inject service instead of repository
	middleware middleware.Middleware,
) *FileSyncHTTPHandler {
	logger = logger.Named("FileSyncHTTPHandler")
	return &FileSyncHTTPHandler{
		config:          config,
		logger:          logger,
		fileSyncService: fileSyncService, // CHANGED: Use service
		middleware:      middleware,
	}
}

func (*FileSyncHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/sync/files"
}

func (h *FileSyncHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *FileSyncHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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

	// Parse limit parameter (default: 5000, max: 10000)
	limit := int64(5000)
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			if parsedLimit > 0 && parsedLimit <= 10000 {
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
	var cursor *dom_file.FileSyncCursor
	if cursorStr := queryParams.Get("cursor"); cursorStr != "" {
		var parsedCursor dom_file.FileSyncCursor
		if err := json.Unmarshal([]byte(cursorStr), &parsedCursor); err != nil {
			h.logger.Error("Failed to parse cursor parameter",
				zap.String("cursor", cursorStr),
				zap.Error(err))
			httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("cursor", "Invalid cursor format"))
			return
		}
		cursor = &parsedCursor
	}

	h.logger.Debug("Processing file sync request",
		zap.Any("user_id", userID),
		zap.Int64("limit", limit),
		zap.Any("cursor", cursor))

	// CHANGED: Call service instead of repository directly (userID comes from context)
	response, err := h.fileSyncService.Execute(ctx, cursor, limit)
	if err != nil {
		h.logger.Error("Failed to get file sync data",
			zap.Any("user_id", userID),
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Encode and return response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode file sync response",
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	h.logger.Info("Successfully served file sync data",
		zap.Any("user_id", userID),
		zap.Int("files_count", len(response.Files)),
		zap.Bool("has_more", response.HasMore))
}
