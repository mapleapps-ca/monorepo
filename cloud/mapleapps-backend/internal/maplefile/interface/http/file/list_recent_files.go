// cloud/mapleapps-backend/internal/maplefile/interface/http/file/list_recent_files.go
package file

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	file_service "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListRecentFilesHTTPHandler struct {
	config                 *config.Configuration
	logger                 *zap.Logger
	listRecentFilesService file_service.ListRecentFilesService
	middleware             middleware.Middleware
}

func NewListRecentFilesHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	listRecentFilesService file_service.ListRecentFilesService,
	middleware middleware.Middleware,
) *ListRecentFilesHTTPHandler {
	logger = logger.Named("ListRecentFilesHTTPHandler")
	return &ListRecentFilesHTTPHandler{
		config:                 config,
		logger:                 logger,
		listRecentFilesService: listRecentFilesService,
		middleware:             middleware,
	}
}

func (*ListRecentFilesHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/files/recent"
}

func (h *ListRecentFilesHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *ListRecentFilesHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Parse query parameters
	queryParams := r.URL.Query()

	// Parse limit parameter (default: 30, max: 100)
	limit := int64(30)
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			if parsedLimit > 0 && parsedLimit <= 100 {
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
	var cursor *string
	if cursorStr := queryParams.Get("cursor"); cursorStr != "" {
		cursor = &cursorStr
	}

	h.logger.Debug("Processing recent files request",
		zap.Int64("limit", limit),
		zap.Any("cursor", cursor))

	// Call service to get recent files
	response, err := h.listRecentFilesService.Execute(ctx, cursor, limit)
	if err != nil {
		h.logger.Error("Failed to get recent files",
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Encode and return response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode recent files response",
			zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	h.logger.Info("Successfully served recent files",
		zap.Int("files_count", len(response.Files)),
		zap.Bool("has_more", response.HasMore),
		zap.Any("next_cursor", response.NextCursor))
}
