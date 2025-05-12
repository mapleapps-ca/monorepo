// cloud/backend/internal/vault/interface/http/encryptedfile.go
package encryptedfile

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/middleware"
	svc "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/service/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// GetEncryptedFileByFileIDHandler handles HTTP requests to get an encrypted file by file ID
type GetEncryptedFileByFileIDHandler struct {
	config             *config.Configuration
	logger             *zap.Logger
	getByFileIDService svc.GetEncryptedFileByFileIDService
	middleware         middleware.Middleware
}

// NewGetEncryptedFileByFileIDHandler creates a new handler for getting a file by file ID
func NewGetEncryptedFileByFileIDHandler(
	config *config.Configuration,
	logger *zap.Logger,
	getByFileIDService svc.GetEncryptedFileByFileIDService,
	middleware middleware.Middleware,
) *GetEncryptedFileByFileIDHandler {
	return &GetEncryptedFileByFileIDHandler{
		config:             config,
		logger:             logger.With(zap.String("handler", "get-encrypted-file-by-file-id")),
		getByFileIDService: getByFileIDService,
		middleware:         middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (h *GetEncryptedFileByFileIDHandler) Pattern() string {
	return "GET /vault/api/v1/files-by-client-id/{fileId}"
}

// ServeHTTP handles HTTP requests
func (h *GetEncryptedFileByFileIDHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.middleware.Attach(h.Execute)(w, r)
}

func (h *GetEncryptedFileByFileIDHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check authentication
	userIDValue := ctx.Value(constants.SessionFederatedUserID)
	if userIDValue == nil {
		httperror.ResponseError(w, httperror.NewForUnauthorizedWithSingleField("message", "Authentication required"))
		return
	}
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok || userID.IsZero() {
		httperror.ResponseError(w, httperror.NewForUnauthorizedWithSingleField("message", "Invalid authentication"))
		return
	}

	// Extract file ID from URL path - updated to match new pattern
	parts := r.URL.Path[len("/api/v1/files-by-client-id/"):]
	if parts == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required"))
		return
	}

	fileID := parts

	// Call service to get the file
	file, err := h.getByFileIDService.Execute(ctx, userID, fileID)
	if err != nil {
		h.logger.Error("Failed to get encrypted file by file ID", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := FileResponse{
		ID:                file.ID,
		UserID:            file.UserID,
		FileID:            file.FileID,
		EncryptedMetadata: file.EncryptedMetadata,
		EncryptionVersion: file.EncryptionVersion,
		EncryptedHash:     file.EncryptedHash,
		CreatedAt:         file.CreatedAt,
		ModifiedAt:        file.ModifiedAt,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}
