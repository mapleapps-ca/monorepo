// monorepo/cloud/backend/internal/maplefile/interface/http/file/complete_file_upload.go
package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CompleteFileUploadHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_file.CompleteFileUploadService
	middleware middleware.Middleware
}

func NewCompleteFileUploadHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_file.CompleteFileUploadService,
	middleware middleware.Middleware,
) *CompleteFileUploadHTTPHandler {
	logger = logger.Named("CompleteFileUploadHTTPHandler")
	return &CompleteFileUploadHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*CompleteFileUploadHTTPHandler) Pattern() string {
	return "POST /maplefile/api/v1/files/{file_id}/complete"
}

func (h *CompleteFileUploadHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *CompleteFileUploadHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
	fileID gocql.UUID,
) (*svc_file.CompleteFileUploadRequestDTO, error) {
	// Initialize our structure which will store the parsed request data
	var requestData svc_file.CompleteFileUploadRequestDTO

	defer r.Body.Close()

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON) // TeeReader allows you to read the JSON and capture it

	// Read the JSON string and convert it into our golang struct
	err := json.NewDecoder(teeReader).Decode(&requestData)
	if err != nil {
		h.logger.Error("decoding error",
			zap.Any("err", err),
			zap.String("json", rawJSON.String()),
		)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Set the file ID from the URL parameter
	requestData.FileID = fileID

	return &requestData, nil
}

func (h *CompleteFileUploadHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
	fileID, err := gocql.ParseUUID(fileIDStr)
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

	resp, err := h.service.Execute(ctx, req)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Encode response
	if resp != nil {
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
