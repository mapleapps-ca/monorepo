// cloud/backend/internal/vault/interface/http/encryptedfile/list.go
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

// ListEncryptedFilesHandler handles HTTP requests to list encrypted files
type ListEncryptedFilesHandler struct {
	config      *config.Configuration
	logger      *zap.Logger
	listService svc.ListEncryptedFilesService
	middleware  middleware.Middleware
}

// NewListEncryptedFilesHandler creates a new handler for listing files
func NewListEncryptedFilesHandler(
	config *config.Configuration,
	logger *zap.Logger,
	listService svc.ListEncryptedFilesService,
	middleware middleware.Middleware,
) *ListEncryptedFilesHandler {
	return &ListEncryptedFilesHandler{
		config:      config,
		logger:      logger.With(zap.String("handler", "list-encrypted-files")),
		listService: listService,
		middleware:  middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (h *ListEncryptedFilesHandler) Pattern() string {
	return "GET /vault/api/v1/encrypted-files"
}

// ServeHTTP handles HTTP requests
func (h *ListEncryptedFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply MaplesSend middleware before handling the request
	h.middleware.Attach(h.Execute)(w, r)
}

func (h *ListEncryptedFilesHandler) Execute(w http.ResponseWriter, r *http.Request) {
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

	// Call service to list files
	files, err := h.listService.Execute(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to list encrypted files", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Convert domain files to response format
	filesResponse := make([]FileResponse, len(files))
	for i, file := range files {
		filesResponse[i] = FileResponse{
			ID:                file.ID,
			UserID:            file.UserID,
			FileID:            file.FileID,
			EncryptedMetadata: file.EncryptedMetadata,
			EncryptionVersion: file.EncryptionVersion,
			EncryptedHash:     file.EncryptedHash,
			CreatedAt:         file.CreatedAt,
			ModifiedAt:        file.ModifiedAt,
		}
	}

	response := FilesListResponse{
		Files: filesResponse,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}
