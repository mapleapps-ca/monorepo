// monorepo/cloud/mapleapps-backend/internal/iam/interface/http/gateway/recovery_initiate.go
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

type InitiateRecoveryHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.InitiateRecoveryService
	middleware middleware.Middleware
}

func NewInitiateRecoveryHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.InitiateRecoveryService,
	middleware middleware.Middleware,
) *InitiateRecoveryHTTPHandler {
	logger = logger.Named("InitiateRecoveryHTTPHandler")
	return &InitiateRecoveryHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*InitiateRecoveryHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/recovery/initiate"
}

func (h *InitiateRecoveryHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *InitiateRecoveryHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.InitiateRecoveryRequestDTO, error) {
	var requestData sv_gateway.InitiateRecoveryRequestDTO

	defer r.Body.Close()

	h.logger.Debug("beginning to decode json payload for api request ...",
		zap.String("api", "/iam/api/v1/recovery/initiate"))

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
		zap.String("api", "/iam/api/v1/recovery/initiate"))

	return &requestData, nil
}

func (h *InitiateRecoveryHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
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
