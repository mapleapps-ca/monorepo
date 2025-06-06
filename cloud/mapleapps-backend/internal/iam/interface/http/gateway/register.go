// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/gateway/register.go
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

type GatewayFederatedUserRegisterHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.GatewayFederatedUserRegisterService
	middleware middleware.Middleware
}

func NewGatewayFederatedUserRegisterHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.GatewayFederatedUserRegisterService,
	middleware middleware.Middleware,
) *GatewayFederatedUserRegisterHTTPHandler {
	logger = logger.Named("GatewayFederatedUserRegisterHTTPHandler")
	return &GatewayFederatedUserRegisterHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GatewayFederatedUserRegisterHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/register"
}

func (r *GatewayFederatedUserRegisterHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *GatewayFederatedUserRegisterHTTPHandler) unmarshalRegisterCustomerRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.RegisterCustomerRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData sv_gateway.RegisterCustomerRequestIDO

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
	requestData.Email = strings.ToLower(requestData.Email)
	requestData.Email = strings.ReplaceAll(requestData.Email, " ", "")

	return &requestData, nil
}

func (h *GatewayFederatedUserRegisterHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalRegisterCustomerRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := h.service.Execute(ctx, data); err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// If transaction succeeds, return success response
	response := map[string]any{
		"message":           "Registration successful. Please check your email for verification.",
		"recovery_key_info": "IMPORTANT: Please ensure you have saved your recovery key. It cannot be retrieved later.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
