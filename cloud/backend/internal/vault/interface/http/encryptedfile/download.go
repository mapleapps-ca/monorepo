// cloud/backend/internal/vault/interface/http/encryptedfile/download.go
package encryptedfile

import (
	"io"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/middleware"
	svc "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/service/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// DownloadEncryptedFileHandler handles HTTP requests to download an encrypted file
type DownloadEncryptedFileHandler struct {
	config          *config.Configuration
	logger          *zap.Logger
	downloadService svc.DownloadEncryptedFileService
	getByIDService  svc.GetEncryptedFileByIDService
	middleware      middleware.Middleware
}

// NewDownloadEncryptedFileHandler creates a new handler for file downloads
func NewDownloadEncryptedFileHandler(
	config *config.Configuration,
	logger *zap.Logger,
	downloadService svc.DownloadEncryptedFileService,
	getByIDService svc.GetEncryptedFileByIDService,
	middleware middleware.Middleware,
) *DownloadEncryptedFileHandler {
	return &DownloadEncryptedFileHandler{
		config:          config,
		logger:          logger.With(zap.String("handler", "download-encrypted-file")),
		downloadService: downloadService,
		getByIDService:  getByIDService,
		middleware:      middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (h *DownloadEncryptedFileHandler) Pattern() string {
	return "GET /vault/api/v1/encrypted-files/{id}/download"
}

// ServeHTTP handles HTTP requests
func (h *DownloadEncryptedFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply MaplesSend middleware before handling the request
	h.middleware.Attach(h.Execute)(w, r)
}

func (h *DownloadEncryptedFileHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract file ID from URL path
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 5 {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("id", "File ID is required"))
		return
	}
	idStr := path[3]

	// Convert string ID to ObjectID
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("id", "Invalid file ID format"))
		return
	}

	// First get the file metadata
	file, err := h.getByIDService.Execute(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get file metadata for download", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Call service to download the file content
	content, err := h.downloadService.Execute(ctx, id)
	if err != nil {
		h.logger.Error("Failed to download encrypted file", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}
	defer content.Close()

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+file.FileID)

	// Stream the file content to the response
	if _, err := io.Copy(w, content); err != nil {
		h.logger.Error("Failed to stream file content", zap.Error(err))
		// Can't really respond with an error here since we've already started writing the response
	}
}
