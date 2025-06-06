// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/gateway/publiclookup.go
package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
	_ "time/tzdata"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GatewayFederatedUserPublicLookupHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.GatewayFederatedUserPublicLookupService
	middleware middleware.Middleware
}

func NewGatewayFederatedUserPublicLookupHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.GatewayFederatedUserPublicLookupService,
	middleware middleware.Middleware,
) *GatewayFederatedUserPublicLookupHTTPHandler {
	logger = logger.Named("GatewayFederatedUserPublicLookupHTTPHandler")
	return &GatewayFederatedUserPublicLookupHTTPHandler{
		logger:     logger,
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

	// 🔍 DEBUG: Log the raw query string to see what's actually received
	h.logger.Debug("🔍 Raw query string", zap.String("raw_query", r.URL.RawQuery))

	// r.URL.Query().Get() already URL-decodes the parameter automatically
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email parameter required", http.StatusBadRequest)
		return
	}

	// 🔍 DEBUG: Log what we got from Query().Get()
	h.logger.Debug("🔍 Email from Query().Get()", zap.String("email", email))
	h.logger.Debug("received email", zap.String("email", email))

	// Basic email validation
	if !strings.Contains(email, "@") {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	var req sv_gateway.GatewayFederatedUserPublicLookupRequestDTO
	req.Email = email

	response, err := h.service.Execute(ctx, &req)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
