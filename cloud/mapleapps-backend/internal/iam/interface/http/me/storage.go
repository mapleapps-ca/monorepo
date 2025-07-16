// github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/me/storage.go
package me

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/interface/http/middleware"
	svc_me "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/service/me"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// GetStorageUsageHTTPHandler handles GET /iam/api/v1/me/storage
type GetStorageUsageHTTPHandler struct {
	logger     *zap.Logger
	service    svc_me.GetStorageUsageService
	middleware middleware.Middleware
}

func NewGetStorageUsageHTTPHandler(
	logger *zap.Logger,
	service svc_me.GetStorageUsageService,
	middleware middleware.Middleware,
) *GetStorageUsageHTTPHandler {
	logger = logger.Named("GetStorageUsageHTTPHandler")
	return &GetStorageUsageHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GetStorageUsageHTTPHandler) Pattern() string {
	return "GET /iam/api/v1/me/storage"
}

func (h *GetStorageUsageHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *GetStorageUsageHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.service.Execute(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpgradePlanHTTPHandler handles POST /iam/api/v1/me/upgrade-plan
type UpgradePlanHTTPHandler struct {
	logger     *zap.Logger
	service    svc_me.UpgradePlanService
	middleware middleware.Middleware
}

func NewUpgradePlanHTTPHandler(
	logger *zap.Logger,
	service svc_me.UpgradePlanService,
	middleware middleware.Middleware,
) *UpgradePlanHTTPHandler {
	logger = logger.Named("UpgradePlanHTTPHandler")
	return &UpgradePlanHTTPHandler{
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*UpgradePlanHTTPHandler) Pattern() string {
	return "POST /iam/api/v1/me/upgrade-plan"
}

func (h *UpgradePlanHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *UpgradePlanHTTPHandler) unmarshalRequest(
	ctx context.Context,
	r *http.Request,
) (*svc_me.UpgradePlanRequestDTO, error) {
	var requestData svc_me.UpgradePlanRequestDTO

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Invalid request format")
	}

	return &requestData, nil
}

func (h *UpgradePlanHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := h.unmarshalRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	result, err := h.service.Execute(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
