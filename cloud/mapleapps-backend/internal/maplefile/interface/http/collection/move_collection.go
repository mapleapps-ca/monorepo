// cloud/backend/internal/maplefile/interface/http/collection/move_collection.go
package collection

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
	svc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type MoveCollectionHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_collection.MoveCollectionService
	middleware middleware.Middleware
}

func NewMoveCollectionHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_collection.MoveCollectionService,
	middleware middleware.Middleware,
) *MoveCollectionHTTPHandler {
	logger = logger.Named("MoveCollectionHTTPHandler")
	return &MoveCollectionHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*MoveCollectionHTTPHandler) Pattern() string {
	return "POST /maplefile/api/v1/collections/{collection_id}/move"
}

func (h *MoveCollectionHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *MoveCollectionHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
	collectionID gocql.UUID,
) (*svc_collection.MoveCollectionRequestDTO, error) {
	// Initialize our structure which will store the parsed request data
	var requestData svc_collection.MoveCollectionRequestDTO

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

	// Set the collection ID from the URL parameter
	requestData.CollectionID = collectionID

	return &requestData, nil
}

func (h *MoveCollectionHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Extract collection ID from URL parameters
	collectionIDStr := r.PathValue("collection_id")
	if collectionIDStr == "" {
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required"))
		return
	}

	// Convert string ID to ObjectID
	collectionID, err := gocql.ParseUUID(collectionIDStr)
	if err != nil {
		h.logger.Error("invalid collection ID format",
			zap.String("collection_id", collectionIDStr),
			zap.Error(err))
		httperror.ResponseError(w, httperror.NewForBadRequestWithSingleField("collection_id", "Invalid collection ID format"))
		return
	}

	req, err := h.unmarshalRequest(ctx, r, collectionID)
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
