// cloud/mapleapps-backend/internal/iam/interface/http/gateway/recovery_verify.go
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

type VerifyRecoveryHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.VerifyRecoveryService
	middleware middleware.Middleware
}

func NewVerifyRecoveryHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.VerifyRecoveryService,
	middleware middleware.Middleware,
) *VerifyRecoveryHTTPHandler {
	logger = logger.Named("VerifyRecoveryHTTPHandler")
	return &VerifyRecoveryHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*VerifyRecoveryHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/recovery/verify"
}

func (h *VerifyRecoveryHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *VerifyRecoveryHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.VerifyRecoveryRequestDTO, error) {
	var requestData sv_gateway.VerifyRecoveryRequestDTO

	defer r.Body.Close()

	h.logger.Debug("beginning to decode json payload for api request ...",
		zap.String("api", "/iam/api/v1/recovery/verify"))

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
		zap.String("api", "/iam/api/v1/recovery/verify"))

	return &requestData, nil
}

func (h *VerifyRecoveryHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
