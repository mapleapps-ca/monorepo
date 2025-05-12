// cloud/backend/internal/vault/interface/http/encryptedfile/delete.go
package encryptedfile

import (
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/middleware"
	svc "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/service/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// DeleteEncryptedFileHandler handles HTTP requests to delete an encrypted file
type DeleteEncryptedFileHandler struct {
	config        *config.Configuration
	logger        *zap.Logger
	deleteService svc.DeleteEncryptedFileService
	middleware    middleware.Middleware
}

// NewDeleteEncryptedFileHandler creates a new handler for file deletion
func NewDeleteEncryptedFileHandler(
	config *config.Configuration,
	logger *zap.Logger,
	deleteService svc.DeleteEncryptedFileService,
	middleware middleware.Middleware,
) *DeleteEncryptedFileHandler {
	return &DeleteEncryptedFileHandler{
		config:        config,
		logger:        logger.With(zap.String("handler", "delete-encrypted-file")),
		deleteService: deleteService,
		middleware:    middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (h *DeleteEncryptedFileHandler) Pattern() string {
	return "DELETE /vault/api/v1/encrypted-files/{id}"
}

// ServeHTTP handles HTTP requests
func (h *DeleteEncryptedFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply MaplesSend middleware before handling the request
	h.middleware.Attach(h.Execute)(w, r)
}

func (h *DeleteEncryptedFileHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract file ID from URL path
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 4 {
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

	// Call service to delete the file
	if err := h.deleteService.Execute(ctx, id); err != nil {
		h.logger.Error("Failed to delete encrypted file", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}
