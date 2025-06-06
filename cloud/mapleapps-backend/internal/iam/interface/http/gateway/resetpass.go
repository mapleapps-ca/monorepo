package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	_ "time/tzdata"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GatewayResetPasswordHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.GatewayResetPasswordService
	middleware middleware.Middleware
}

func NewGatewayResetPasswordHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.GatewayResetPasswordService,
	middleware middleware.Middleware,
) *GatewayResetPasswordHTTPHandler {
	logger = logger.Named("GatewayResetPasswordHTTPHandler")
	return &GatewayResetPasswordHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GatewayResetPasswordHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/reset-password"
}

func (r *GatewayResetPasswordHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *GatewayResetPasswordHTTPHandler) unmarshalLoginRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.GatewayResetPasswordRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData sv_gateway.GatewayResetPasswordRequestIDO

	defer r.Body.Close()

	h.logger.Debug("beginning to decode json payload for api request ...",
		zap.String("api", "/iam/api/v1/reset-password"))

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

	h.logger.Debug("successfully decoded json payload api request",
		zap.String("api", "/iam/api/v1/reset-password"))

	return &requestData, nil
}

func (h *GatewayResetPasswordHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalLoginRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	resp, err := h.service.Execute(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
