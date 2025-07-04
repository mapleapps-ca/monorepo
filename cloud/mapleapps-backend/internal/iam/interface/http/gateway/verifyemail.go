package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	_ "time/tzdata"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GatewayVerifyEmailHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.GatewayVerifyEmailService
	middleware middleware.Middleware
}

func NewGatewayVerifyEmailHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.GatewayVerifyEmailService,
	middleware middleware.Middleware,
) *GatewayVerifyEmailHTTPHandler {
	logger = logger.Named("GatewayVerifyEmailHTTPHandler")
	return &GatewayVerifyEmailHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GatewayVerifyEmailHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/verify-email-code"
}

func (r *GatewayVerifyEmailHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *GatewayVerifyEmailHTTPHandler) unmarshalVerifyRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.GatewayVerifyEmailRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData sv_gateway.GatewayVerifyEmailRequestIDO

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

	// Defensive Code: For security purposes we need to remove all whitespaces from the email and lower the characters.
	requestData.Code = strings.ReplaceAll(requestData.Code, " ", "")

	return &requestData, nil
}

func (h *GatewayVerifyEmailHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalVerifyRequest(ctx, r)
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
