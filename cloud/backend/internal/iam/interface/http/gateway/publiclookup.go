// github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/gateway/register.go
package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	_ "time/tzdata"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GatewayFederatedUserPublicLookupHTTPHandler struct {
	logger     *zap.Logger
	dbClient   *mongo.Client
	service    sv_gateway.GatewayFederatedUserPublicLookupService
	middleware middleware.Middleware
}

func NewGatewayFederatedUserPublicLookupHTTPHandler(
	logger *zap.Logger,
	dbClient *mongo.Client,
	service sv_gateway.GatewayFederatedUserPublicLookupService,
	middleware middleware.Middleware,
) *GatewayFederatedUserPublicLookupHTTPHandler {
	logger = logger.Named("GatewayFederatedUserPublicLookupHTTPHandler")
	return &GatewayFederatedUserPublicLookupHTTPHandler{
		logger:     logger,
		dbClient:   dbClient,
		service:    service,
		middleware: middleware,
	}
}

func (*GatewayFederatedUserPublicLookupHTTPHandler) Pattern() string {
	return "GET /iam/api/v1/users/lookup"
}

func (r *GatewayFederatedUserPublicLookupHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *GatewayFederatedUserPublicLookupHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email parameter required", http.StatusBadRequest)
		return
	}

	var req sv_gateway.GatewayFederatedUserPublicLookupRequestDTO
	req.Email = email

	////
	//// Start the transaction.
	////

	session, err := h.dbClient.StartSession()
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx context.Context) (any, error) {
		result, err := h.service.Execute(sessCtx, &req)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	// Start a transaction
	result, txErr := session.WithTransaction(ctx, transactionFunc)
	if txErr != nil {
		httperror.ResponseError(w, txErr)
		return
	}

	// If transaction succeeds, return success response
	response := result.(*sv_gateway.GatewayFederatedUserPublicLookupResponseDTO)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
