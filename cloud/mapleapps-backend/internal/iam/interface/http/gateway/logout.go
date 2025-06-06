package gateway

import (
	"net/http"
	_ "time/tzdata"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	sv_gateway "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/gateway"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GatewayLogoutHTTPHandler struct {
	logger     *zap.Logger
	service    sv_gateway.GatewayLogoutService
	middleware middleware.Middleware
}

func NewGatewayLogoutHTTPHandler(
	logger *zap.Logger,
	service sv_gateway.GatewayLogoutService,
	middleware middleware.Middleware,
) *GatewayLogoutHTTPHandler {
	logger = logger.Named("GatewayLogoutHTTPHandler")
	return &GatewayLogoutHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (r *GatewayLogoutHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (*GatewayLogoutHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/logout"
}

func (h *GatewayLogoutHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.service.Execute(ctx); err != nil {
		httperror.ResponseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
