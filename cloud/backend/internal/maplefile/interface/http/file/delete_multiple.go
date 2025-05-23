// cloud/backend/internal/maplefile/interface/http/file/delete_multiple.go
package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/interface/http/middleware"
	svc_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/service/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteMultipleFilesHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    svc_file.DeleteMultipleFilesService
	middleware middleware.Middleware
}

func NewDeleteMultipleFilesHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
	service svc_file.DeleteMultipleFilesService,
	middleware middleware.Middleware,
) *DeleteMultipleFilesHTTPHandler {
	return &DeleteMultipleFilesHTTPHandler{
		config:     config,
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*DeleteMultipleFilesHTTPHandler) Pattern() string {
	return "DELETE /maplefile/api/v1/files/multiple"
}

func (h *DeleteMultipleFilesHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *DeleteMultipleFilesHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*svc_file.DeleteMultipleFilesRequestDTO, error) {
	// Initialize our structure which will store the parsed request data
	var requestData svc_file.DeleteMultipleFilesRequestDTO

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

	return &requestData, nil
}

func (h *DeleteMultipleFilesHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	req, err := h.unmarshalRequest(ctx, r)
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
			h.logger.Error("failed to delete multiple files",
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
		resp := result.(*svc_file.DeleteMultipleFilesResponseDTO)
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
