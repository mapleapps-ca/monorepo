// cloud/backend/internal/maplefile/interface/http/file/get_presigned_upload_url.go
package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetPresignedUploadURLHTTPRequestDTO struct {
	URLDurationStr string `json:"url_duration,omitempty"` // Optional, duration as string of nanoseconds, defaults to 1 hour
}

type GetPresignedUploadURLHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.GetPresignedUploadURLService
	middleware middleware.Middleware
}

func NewGetPresignedUploadURLHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.GetPresignedUploadURLService,
	middleware middleware.Middleware,
) *GetPresignedUploadURLHTTPHandler {
	logger = logger.Named("GetPresignedUploadURLHTTPHandler")
	return &GetPresignedUploadURLHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*GetPresignedUploadURLHTTPHandler) Pattern() string {
	return "POST /maplefile/api/v1/files/{file_id}/upload-url"
}

func (h *GetPresignedUploadURLHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *GetPresignedUploadURLHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
	fileID primitive.ObjectID,
) (*svc_file.GetPresignedUploadURLRequestDTO, error) {
	// Initialize our structure which will store the parsed request data
	var httpRequestData GetPresignedUploadURLHTTPRequestDTO

	defer r.Body.Close()

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON) // TeeReader allows you to read the JSON and capture it

	// Read the JSON string and convert it into our golang struct
	err := json.NewDecoder(teeReader).Decode(&httpRequestData)
	if err != nil {
		h.logger.Error("decoding error",
			zap.Any("err", err),
			zap.String("json", rawJSON.String()),
		)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Set default URL duration if not provided (1 hour in nanoseconds)
	var urlDuration time.Duration
	if httpRequestData.URLDurationStr == "" {
		urlDuration = 1 * time.Hour
	} else {
		// Parse the string to int64 (nanoseconds)
		durationNanos, err := strconv.ParseInt(httpRequestData.URLDurationStr, 10, 64)
		if err != nil {
			return nil, httperror.NewForSingleField(http.StatusBadRequest, "url_duration", "Invalid duration format")
		}
		urlDuration = time.Duration(durationNanos)
	}

	// Convert to service DTO
	serviceRequest := &svc_file.GetPresignedUploadURLRequestDTO{
		FileID:      fileID,
		URLDuration: urlDuration,
	}

	return serviceRequest, nil
}

func (h *GetPresignedUploadURLHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Extract file ID from URL parameters
	fileIDStr := r.PathValue("file_id")
	if fileIDStr == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required"))
		return
	}

	// Convert string ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(fileIDStr)
	if err != nil {
		h.logger.Error("invalid file ID format",
			zap.String("file_id", fileIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("file_id", "Invalid file ID format"))
		return
	}

	req, err := h.unmarshalRequest(ctx, r, fileID)
	if err != nil {
		httperror.ResponseError(w, err)
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
		response, err := h.service.Execute(sessCtx, req)
		if err != nil {
			h.logger.Error("failed to generate presigned upload URLs",
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

	// Encode response
	if result != nil {
		resp := result.(*svc_file.GetPresignedUploadURLResponseDTO)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("failed to encode response",
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
