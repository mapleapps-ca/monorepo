// cloud/backend/internal/vault/interface/http/encryptedfile/getdownloadurl.go
package encryptedfile

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/middleware"
	svc "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/service/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// GetEncryptedFileDownloadURLHandler handles HTTP requests to get a download URL for an encrypted file
type GetEncryptedFileDownloadURLHandler struct {
	config        *config.Configuration
	logger        *zap.Logger
	getURLService svc.GetEncryptedFileDownloadURLService
	middleware    middleware.Middleware
}

// NewGetEncryptedFileDownloadURLHandler creates a new handler for getting download URLs
func NewGetEncryptedFileDownloadURLHandler(
	config *config.Configuration,
	logger *zap.Logger,
	getURLService svc.GetEncryptedFileDownloadURLService,
	middleware middleware.Middleware,
) *GetEncryptedFileDownloadURLHandler {
	return &GetEncryptedFileDownloadURLHandler{
		config:        config,
		logger:        logger.With(zap.String("handler", "get-encrypted-file-download-url")),
		getURLService: getURLService,
		middleware:    middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (h *GetEncryptedFileDownloadURLHandler) Pattern() string {
	return "GET /vault/api/v1/encrypted-files/{id}/url"
}

// ServeHTTP handles HTTP requests
func (h *GetEncryptedFileDownloadURLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply MaplesSend middleware before handling the request
	h.middleware.Attach(h.Execute)(w, r)
}

func (h *GetEncryptedFileDownloadURLHandler) Execute(w http.ResponseWriter, r *http.Request) {
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

	// Get expiry duration from query parameter (optional)
	expiryDuration := 15 * time.Minute // Default expiry
	expiryStr := r.URL.Query().Get("expiry")
	if expiryStr != "" {
		expiryMinutes, err := strconv.Atoi(expiryStr)
		if err != nil {
			httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("expiry", "Invalid expiry value, must be a number of minutes"))
			return
		}

		if expiryMinutes <= 0 || expiryMinutes > 1440 { // Max 24 hours (1440 minutes)
			httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("expiry", "Expiry must be between 1 and 1440 minutes"))
			return
		}

		expiryDuration = time.Duration(expiryMinutes) * time.Minute
	}

	// Call service to get the download URL
	url, err := h.getURLService.Execute(ctx, id, expiryDuration)
	if err != nil {
		h.logger.Error("Failed to get download URL", zap.Error(err))
		httperror.ResponseError(w, err)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := FileURLResponse{
		URL:       url,
		ExpiresAt: time.Now().Add(expiryDuration),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}
