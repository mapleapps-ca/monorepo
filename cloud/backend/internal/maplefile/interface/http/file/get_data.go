// cloud/backend/internal/maplefile/interface/http/file/get_data.go
package file

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileDataHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.GetFileDataService
	middleware middleware.Middleware
}

func NewGetFileDataHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.GetFileDataService,
	middleware middleware.Middleware,
) *GetFileDataHTTPHandler {
	return &GetFileDataHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*GetFileDataHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/files/{file_id}/data"
}

func (h *GetFileDataHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *GetFileDataHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("🔎 Starting file data download...")

	ctx := r.Context()

	// Extract file ID from the URL
	fileID := r.PathValue("file_id")
	if fileID == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required"))
		return
	}

	// Start the transaction
	session, err := h.dbClient.StartSession()
	if err != nil {
		h.logger.Error("start session error",
			zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx context.Context) (interface{}, error) {
		// Call service
		response, err := h.service.Execute(sessCtx, fileID)
		if err != nil {
			h.logger.Error("failed to get file data",
				zap.Any("error", err))
			return nil, err
		}
		return response, nil
	}

	// Start a transaction
	result, txErr := session.WithTransaction(ctx, transactionFunc)
	if txErr != nil {
		h.logger.Error("session failed error",
			zap.Any("error", txErr))
		httperror.ResponseError(w, txErr)
		return
	}

	// Get the result
	fileData, ok := result.([]byte)
	if !ok || fileData == nil {
		err := errors.New("no file data found")
		h.logger.Error("failed to get file data",
			zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}

	// Set appropriate headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=encrypted_file.bin")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileData)))

	// Write the file data to the response
	if _, err := w.Write(fileData); err != nil {
		h.logger.Error("failed to write file data to response",
			zap.Any("error", err))
		// Can't send error response at this point as we've already started writing the response
	}
}
