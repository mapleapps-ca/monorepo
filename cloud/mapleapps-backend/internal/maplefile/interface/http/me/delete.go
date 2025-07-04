// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/me/delete.go
package me

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_me "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type DeleteMeHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_me.DeleteMeService
	middleware middleware.Middleware
}

func NewDeleteMeHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_me.DeleteMeService,
	middleware middleware.Middleware,
) *DeleteMeHTTPHandler {
	logger = logger.With(zap.String("module", "maplefile"))
	logger = logger.Named("DeleteMeHTTPHandler")
	return &DeleteMeHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*DeleteMeHTTPHandler) Pattern() string {
	return "DELETE /maplefile/api/v1/me"
}

func (r *DeleteMeHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *DeleteMeHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*svc_me.DeleteMeRequestDTO, error) {
	// Initialize our structure which will store the parsed request data
	var requestData svc_me.DeleteMeRequestDTO

	defer r.Body.Close()

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON) // TeeReader allows you to read the JSON and capture it

	// Read the JSON string and convert it into our golang struct else we need
	// to send a `400 Bad Request` error message back to the client
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

func (h *DeleteMeHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	req, err := h.unmarshalRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := h.service.Execute(ctx, req); err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Return successful no content response since the account was deleted
	w.WriteHeader(http.StatusNoContent)
}
