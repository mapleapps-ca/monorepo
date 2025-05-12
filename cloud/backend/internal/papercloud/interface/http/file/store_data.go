// cloud/backend/internal/papercloud/interface/http/file/store_data.go
package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type StoreFileDataHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.StoreFileDataService
	middleware middleware.Middleware
}

func NewStoreFileDataHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.StoreFileDataService,
	middleware middleware.Middleware,
) *StoreFileDataHTTPHandler {
	return &StoreFileDataHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*StoreFileDataHTTPHandler) Pattern() string {
	return "POST /papercloud/api/v1/files/{file_id}/data"
}

func (h *StoreFileDataHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *StoreFileDataHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("🔎 Starting upload file to save data...")

	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Extract file ID from the URL
	fileID := r.PathValue("file_id")
	if fileID == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required"))
		return
	}

	// Parse multipart form data (max 50MB)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		h.logger.Error("🔴 Failed to parse multipart form",
			zap.Any("error", err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file", "Failed to parse multipart form"))
		return
	}

	// Get the file from the form
	file, _, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("🔴 Failed to get file from form",
			zap.Any("error", err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file", "Failed to get file from form"))
		return
	}
	defer file.Close()

	// Read all the file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("🔴 Failed to read file data",
			zap.Any("error", err))
		httperror.ResponseError(w, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to read file data"))
		return
	}

	// Create request DTO
	dtoReq := &svc_file.StoreFileDataRequestDTO{
		ID:   fileID,
		Data: fileData,
	}

	// Start the transaction
	session, err := h.dbClient.StartSession()
	if err != nil {
		h.logger.Error("🔴 start session error",
			zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx context.Context) (interface{}, error) {
		// Call service
		response, err := h.service.Execute(sessCtx, dtoReq)
		if err != nil {
			h.logger.Error("🔴 failed to store file data",
				zap.Any("error", err))
			return nil, err
		}
		return response, nil
	}

	// Start a transaction
	result, txErr := session.WithTransaction(ctx, transactionFunc)
	if txErr != nil {
		h.logger.Error("🔴 session failed error",
			zap.Any("error", txErr))
		httperror.ResponseError(w, txErr)
		return
	}

	// Encode response
	if result != nil {
		resp := result.(*svc_file.StoreFileDataResponseDTO)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("🔴 failed to encode response",
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
