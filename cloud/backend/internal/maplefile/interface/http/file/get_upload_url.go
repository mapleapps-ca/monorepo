// cloud/backend/internal/maplefile/interface/http/file/get_upload_url.go
package file

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

// GetFileUploadURLHTTPHandler handles requests for generating file upload URLs
type GetFileUploadURLHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.GetFileUploadURLService
	middleware middleware.Middleware
}

// NewGetFileUploadURLHTTPHandler creates a new instance of GetFileUploadURLHTTPHandler
func NewGetFileUploadURLHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.GetFileUploadURLService,
	middleware middleware.Middleware,
) *GetFileUploadURLHTTPHandler {
	return &GetFileUploadURLHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

// Pattern returns the URL pattern for this handler
func (*GetFileUploadURLHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/files/{file_id}/upload-url"
}

// ServeHTTP handles HTTP requests to this endpoint
func (h *GetFileUploadURLHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

// Execute handles the request logic
func (h *GetFileUploadURLHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		h.logger.Error("invalid file ID format", zap.String("file_id", fileIDStr), zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "Invalid file ID format"))
		return
	}

	// Get user ID from context
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		h.logger.Error("Failed getting user ID from context")
		httperror.ResponseError(w, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error"))
		return
	}

	// Start the transaction
	session, err := h.dbClient.StartSession()
	if err != nil {
		h.logger.Error("start session error", zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}
	defer session.EndSession(ctx)

	// Define a transaction function
	transactionFunc := func(sessCtx context.Context) (interface{}, error) {
		// Call service to get the upload URL
		response, err := h.service.Execute(sessCtx, fileID, userID)
		if err != nil {
			h.logger.Error("failed to get upload URL", zap.Any("error", err))
			return nil, err
		}
		return response, nil
	}

	// Start a transaction
	result, txErr := session.WithTransaction(ctx, transactionFunc)
	if txErr != nil {
		h.logger.Error("session failed error", zap.Any("error", txErr))
		httperror.ResponseError(w, txErr)
		return
	}

	// Encode response
	if result != nil {
		resp := result.(*svc_file.FileUploadURLResponseDTO)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("failed to encode response", zap.Any("error", err))
			httperror.ResponseError(w, err)
			return
		}
	} else {
		err := errors.New("no result")
		httperror.ResponseError(w, err)
		return
	}
}
