// cloud/backend/internal/iam/interface/http/gateway/requestott.go
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	_ "time/tzdata"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GatewayRequestLoginOTTHTTPHandler struct {
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    sv_gateway.GatewayRequestLoginOTTService
	middleware middleware.Middleware
}

func NewGatewayRequestLoginOTTHTTPHandler(
	logger *zap.Logger,
	dbClient *mongo.Client,
	service sv_gateway.GatewayRequestLoginOTTService,
	middleware middleware.Middleware,
) *GatewayRequestLoginOTTHTTPHandler {
	return &GatewayRequestLoginOTTHTTPHandler{
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*GatewayRequestLoginOTTHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/request-ott"
}

func (r *GatewayRequestLoginOTTHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *GatewayRequestLoginOTTHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*sv_gateway.GatewayRequestLoginOTTRequestIDO, error) {
	var requestData sv_gateway.GatewayRequestLoginOTTRequestIDO

	defer r.Body.Close()

	h.logger.Debug("beginning to decode json payload for api request ...",
		zap.String("api", "/iam/api/v1/request-login-ott"))

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

	// Defensive Code: For security purposes we need to remove all whitespaces from the email and lower the characters.
	requestData.Email = strings.ToLower(requestData.Email)
	requestData.Email = strings.ReplaceAll(requestData.Email, " ", "")

	h.logger.Debug("successfully decoded json payload api request",
		zap.String("api", "/iam/api/v1/request-login-ott"))

	return &requestData, nil
}

func (h *GatewayRequestLoginOTTHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Start the transaction
	session, err := h.dbClient.StartSession()
	if err != nil {
		h.logger.Error("start session error", zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}
	defer session.EndSession(ctx)

	// Define a transaction function
	transactionFunc := func(sessCtx context.Context) (interface{}, error) {
		resp, err := h.service.Execute(sessCtx, data)
		if err != nil {
			h.logger.Error("service error", zap.Any("err", err))
			return nil, err
		}
		return resp, nil
	}

	// Start the transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		h.logger.Error("session failed error", zap.Any("error", err))
		httperror.ResponseError(w, err)
		return
	}

	resp := result.(*sv_gateway.GatewayRequestLoginOTTResponseIDO)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
