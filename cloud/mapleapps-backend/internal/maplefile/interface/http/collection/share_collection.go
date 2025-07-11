// cloud/backend/internal/maplefile/interface/http/collection/share_collection.go
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

type ShareCollectionHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_collection.ShareCollectionService
	middleware middleware.Middleware
}

func NewShareCollectionHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_collection.ShareCollectionService,
	middleware middleware.Middleware,
) *ShareCollectionHTTPHandler {
	logger = logger.Named("ShareCollectionHTTPHandler")
	return &ShareCollectionHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*ShareCollectionHTTPHandler) Pattern() string {
	return "POST /maplefile/api/v1/collections/{collection_id}/share"
}

func (h *ShareCollectionHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *ShareCollectionHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
	collectionID gocql.UUID,
) (*svc_collection.ShareCollectionRequestDTO, error) {
	// Initialize our structure which will store the parsed request data
	var requestData svc_collection.ShareCollectionRequestDTO

	defer r.Body.Close()

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON) // TeeReader allows you to read the JSON and capture it

	// Read the JSON string and convert it into our golang struct
	err := json.NewDecoder(teeReader).Decode(&requestData)
	if err != nil {
		h.logger.Error("JSON decoding error",
			zap.Any("err", err),
			zap.String("raw_json", rawJSON.String()),
		)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	// Log the decoded request for debugging
	h.logger.Info("decoded share collection request",
		zap.String("collection_id_from_url", collectionID.String()),
		zap.String("collection_id_from_body", requestData.CollectionID.String()),
		zap.String("recipient_id", requestData.RecipientID.String()),
		zap.String("recipient_email", requestData.RecipientEmail),
		zap.String("permission_level", requestData.PermissionLevel),
		zap.Int("encrypted_key_length", len(requestData.EncryptedCollectionKey)),
		zap.Bool("share_with_descendants", requestData.ShareWithDescendants),
		zap.String("raw_json", rawJSON.String()))

	// CRITICAL: Check if encrypted collection key is present in the request
	if len(requestData.EncryptedCollectionKey) == 0 {
		h.logger.Error("FRONTEND BUG: encrypted_collection_key is missing from request",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", requestData.RecipientID.String()),
			zap.String("recipient_email", requestData.RecipientEmail),
			zap.String("raw_json", rawJSON.String()))
	} else {
		h.logger.Info("encrypted_collection_key found in request",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", requestData.RecipientID.String()),
			zap.Int("encrypted_key_length", len(requestData.EncryptedCollectionKey)))
	}

	// Set the collection ID from the URL parameter
	requestData.CollectionID = collectionID

	return &requestData, nil
}

func (h *ShareCollectionHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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

	h.logger.Info("processing share collection request",
		zap.String("collection_id", collectionID.String()),
		zap.String("method", r.Method),
		zap.String("content_type", r.Header.Get("Content-Type")))

	req, err := h.unmarshalRequest(ctx, r, collectionID)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Call service
	resp, err := h.service.Execute(ctx, req)
	if err != nil {
		h.logger.Error("share collection service failed",
			zap.String("collection_id", collectionID.String()),
			zap.String("recipient_id", req.RecipientID.String()),
			zap.Error(err))
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
