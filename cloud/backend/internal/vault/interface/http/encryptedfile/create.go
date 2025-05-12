// cloud/backend/internal/vault/interface/http/encryptedfile/create.go
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

// CreateEncryptedFileHandler handles HTTP requests to create a new encrypted file
type CreateEncryptedFileHandler struct {
	config        *config.Configuration
	logger        *zap.Logger
	createService svc.CreateEncryptedFileService
	middleware    middleware.Middleware
}

// NewCreateEncryptedFileHandler creates a new handler for file creation
func NewCreateEncryptedFileHandler(
	config *config.Configuration,
	logger *zap.Logger,
	createService svc.CreateEncryptedFileService,
	middleware middleware.Middleware,
) *CreateEncryptedFileHandler {
	return &CreateEncryptedFileHandler{
		config:        config,
		logger:        logger.With(zap.String("handler", "create-encrypted-file")),
		createService: createService,
		middleware:    middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (h *CreateEncryptedFileHandler) Pattern() string {
	return "POST /vault/api/v1/encrypted-files"
}

// ServeHTTP handles HTTP requests
func (h *CreateEncryptedFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply MaplesSend middleware before handling the request
	h.middleware.Attach(h.Execute)(w, r)
}

func (h *CreateEncryptedFileHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check authentication
	userIDValue := ctx.Value(constants.SessionFederatedUserID)
	if userIDValue == nil {
		h.logger.Error("--> anonymous user detected")
		httperror.ResponseError(w, httperror.NewForUnauthorizedWithSingleField("message", "Authentication required"))
		return
	}
	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok || userID.IsZero() {
		httperror.ResponseError(w, httperror.NewForUnauthorizedWithSingleField("message", "Invalid authentication"))
		return
	}

	// Parse multipart form to get file and metadata
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("content", "Invalid multipart form"))
		return
	}

	// Extract form fields
	fileID := r.FormValue("file_id")
	if fileID == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required"))
		return
	}

	encryptedMetadata := r.FormValue("encrypted_metadata")
	encryptedHash := r.FormValue("encrypted_hash")
	encryptionVersion := r.FormValue("encryption_version")
	if encryptionVersion == "" {
		encryptionVersion = "1.0" // Default version
	}

	// Get file content
	file, _, err := r.FormFile("encrypted_content")
	if err != nil {
		h.logger.Error("Failed to get file from form", zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("encrypted_content", "File content is required"))
		return
	}
	defer file.Close()

	// Call service to create the file
	result, err := h.createService.Execute(
		ctx,
		userID,
		fileID,
		encryptedMetadata,
		encryptedHash,
		encryptionVersion,
		file,
	)
	if err != nil {
		h.logger.Error("Failed to create encrypted file", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := FileResponse{
		ID:                result.ID,
		UserID:            result.UserID,
		FileID:            result.FileID,
		EncryptedMetadata: result.EncryptedMetadata,
		EncryptionVersion: result.EncryptionVersion,
		EncryptedHash:     result.EncryptedHash,
		CreatedAt:         result.CreatedAt,
		ModifiedAt:        result.ModifiedAt,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}
