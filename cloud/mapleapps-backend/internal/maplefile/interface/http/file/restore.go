// monorepo/cloud/backend/internal/maplefile/interface/http/file/restore.go
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

type RestoreFileHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_file.RestoreFileService
	middleware middleware.Middleware
}

func NewRestoreFileHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_file.RestoreFileService,
	middleware middleware.Middleware,
) *RestoreFileHTTPHandler {
	logger = logger.Named("RestoreFileHTTPHandler")
	return &RestoreFileHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*RestoreFileHTTPHandler) Pattern() string {
	return "POST /maplefile/api/v1/files/{file_id}/restore"
}

func (h *RestoreFileHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *RestoreFileHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Extract file ID from the URL
	fileIDStr := r.PathValue("file_id")
	if fileIDStr == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required"))
		return
	}

	// Convert string ID to ObjectID
	fileID, err := gocql.ParseUUID(fileIDStr)
	if err != nil {
		h.logger.Error("invalid file ID format",
			zap.String("file_id", fileIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "Invalid file ID format"))
		return
	}

	// Create request DTO
	dtoReq := &svc_file.RestoreFileRequestDTO{
		FileID: fileID,
	}

	resp, err := h.service.Execute(ctx, dtoReq)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Encode response
	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("failed to encode response",
				zap.String("file_id", fileIDStr),
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
