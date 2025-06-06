// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/me/get.go
package me

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_me "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetMeHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_me.GetMeService
	middleware middleware.Middleware
}

func NewGetMeHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_me.GetMeService,
	middleware middleware.Middleware,
) *GetMeHTTPHandler {
	logger = logger.With(zap.String("module", "maplefile"))
	logger = logger.Named("GetMeHTTPHandler")
	return &GetMeHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GetMeHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/me"
}

func (r *GetMeHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply MaplesSend middleware before handling the request
	r.middleware.Attach(r.Execute)(w, req)
}

func (h *GetMeHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	resp, err := h.service.Execute(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	// Encode response
	if resp != nil {
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			h.logger.Error("failed to encode response",
				zap.Any("error", err))
			httperror.ResponseError(w, err)
			return
		}
	} else {
		err := errors.New("no result")
		httperror.ResponseError(w, err)
		return
	}

}
