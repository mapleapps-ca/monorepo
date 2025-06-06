// cloud/backend/internal/maplefile/interface/http/file/archive.go
package file

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ArchiveFileHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.ArchiveFileService
	middleware middleware.Middleware
}

func NewArchiveFileHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.ArchiveFileService,
	middleware middleware.Middleware,
) *ArchiveFileHTTPHandler {
	logger = logger.Named("ArchiveFileHTTPHandler")
	return &ArchiveFileHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*ArchiveFileHTTPHandler) Pattern() string {
	return "POST /maplefile/api/v1/files/{file_id}/archive"
}

func (h *ArchiveFileHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *ArchiveFileHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
	fileID, err := gocql.UUIDFromHex(fileIDStr)
	if err != nil {
		h.logger.Error("invalid file ID format",
			zap.String("file_id", fileIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "Invalid file ID format"))
		return
	}

	// Create request DTO
	dtoReq := &svc_file.ArchiveFileRequestDTO{
		FileID: fileID,
	}

	// Start the transaction
	session, err := h.dbClient.StartSession()
	if err != nil {
		h.logger.Error("start session error",
			zap.String("file_id", fileIDStr),
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
			h.logger.Error("failed to archive file",
				zap.String("file_id", fileIDStr),
				zap.Any("error", err))
			return nil, err
		}
		return response, nil
	}

	// Start a transaction
	result, txErr := session.WithTransaction(ctx, transactionFunc)
	if txErr != nil {
		h.logger.Error("session failed error",
			zap.String("file_id", fileIDStr),
			zap.Any("error", txErr))
		httperror.ResponseError(w, txErr)
		return
	}

	// Encode response
	if result != nil {
		resp := result.(*svc_file.ArchiveFileResponseDTO)
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
