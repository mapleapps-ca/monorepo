// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/me/get.go
package me

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_me "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type PutUpdateMeHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_me.UpdateMeService
	middleware middleware.Middleware
}

func NewPutUpdateMeHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_me.UpdateMeService,
	middleware middleware.Middleware,
) *PutUpdateMeHTTPHandler {
	logger = logger.With(zap.String("module", "maplefile"))
	logger = logger.Named("PutUpdateMeHTTPHandler")
	return &PutUpdateMeHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*PutUpdateMeHTTPHandler) Pattern() string {
	return "PUT /maplefile/api/v1/me"
}

func (r *PutUpdateMeHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *PutUpdateMeHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*svc_me.UpdateMeRequestDTO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData svc_me.UpdateMeRequestDTO

	defer r.Body.Close()

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON) // TeeReader allows you to read the JSON and capture it

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(teeReader).Decode(&requestData) // [1]
	if err != nil {
		h.logger.Error("decoding error",
			zap.Any("err", err),
			zap.String("json", rawJSON.String()),
		)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	return &requestData, nil
}

func (h *PutUpdateMeHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	req, err := h.unmarshalRequest(ctx, r)
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
