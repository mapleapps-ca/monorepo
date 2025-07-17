// monorepo/cloud/mapleapps-backend/internal/iam/interface/http/gateway/recovery_complete.go
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CompleteRecoveryHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.CompleteRecoveryService
	middleware middleware.Middleware
}

func NewCompleteRecoveryHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.CompleteRecoveryService,
	middleware middleware.Middleware,
) *CompleteRecoveryHTTPHandler {
	logger = logger.Named("CompleteRecoveryHTTPHandler")
	return &CompleteRecoveryHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*CompleteRecoveryHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/recovery/complete"
}

func (h *CompleteRecoveryHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *CompleteRecoveryHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.CompleteRecoveryRequestDTO, error) {
	var requestData sv_gateway.CompleteRecoveryRequestDTO

	defer r.Body.Close()

	h.logger.Debug("beginning to decode json payload for api request ...",
		zap.String("api", "/iam/api/v1/recovery/complete"))

	var rawJSON bytes.Buffer
	teeReader := io.TeeReader(r.Body, &rawJSON)

	// Read the JSON string and convert it into our golang struct
	err := json.NewDecoder(teeReader).Decode(&requestData)
	if err != nil {
		h.logger.Error("decoding error",
			zap.Any("err", err),
			zap.String("json", rawJSON.String()),
		)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}

	h.logger.Debug("successfully decoded json payload api request",
		zap.Any("requestData:", requestData),
		zap.String("api", "/iam/api/v1/recovery/complete"))

	return &requestData, nil
}

func (h *CompleteRecoveryHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	result, err := h.service.Execute(ctx, data)
	if err != nil {
		h.logger.Error("service error", zap.Any("err", err))
		httperror.ResponseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
