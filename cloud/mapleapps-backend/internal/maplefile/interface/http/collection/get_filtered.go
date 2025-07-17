// monorepo/cloud/backend/internal/maplefile/interface/http/collection/get_filtered.go
package collection

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/interface/http/middleware"
	svc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/service/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFilteredCollectionsHTTPHandler struct {
	config     *config.Configuration
	logger     *zap.Logger
	service    svc_collection.GetFilteredCollectionsService
	middleware middleware.Middleware
}

func NewGetFilteredCollectionsHTTPHandler(
	config *config.Configuration,
	logger *zap.Logger,
	service svc_collection.GetFilteredCollectionsService,
	middleware middleware.Middleware,
) *GetFilteredCollectionsHTTPHandler {
	logger = logger.Named("GetFilteredCollectionsHTTPHandler")
	return &GetFilteredCollectionsHTTPHandler{
		config:     config,
		logger:     logger,
		service:    service,
		middleware: middleware,
	}
}

func (*GetFilteredCollectionsHTTPHandler) Pattern() string {
	return "GET /maplefile/api/v1/collections/filtered"
}

func (h *GetFilteredCollectionsHTTPHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Apply middleware before handling the request
	h.middleware.Attach(h.Execute)(w, req)
}

func (h *GetFilteredCollectionsHTTPHandler) Execute(w http.ResponseWriter, r *http.Request) {
	// Set response content type
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	// Parse query parameters for filter options
	req, err := h.parseFilterOptions(r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	resp, err := h.service.Execute(ctx, req)
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

// parseFilterOptions parses the query parameters to create the request DTO
func (h *GetFilteredCollectionsHTTPHandler) parseFilterOptions(r *http.Request) (*svc_collection.GetFilteredCollectionsRequestDTO, error) {
	req := &svc_collection.GetFilteredCollectionsRequestDTO{
		IncludeOwned:  true,  // Default to including owned collections
		IncludeShared: false, // Default to not including shared collections
	}

	// Parse include_owned parameter
	if includeOwnedStr := r.URL.Query().Get("include_owned"); includeOwnedStr != "" {
		includeOwned, err := strconv.ParseBool(includeOwnedStr)
		if err != nil {
			h.logger.Warn("Invalid include_owned parameter",
				zap.String("value", includeOwnedStr),
				zap.Error(err))
			return nil, httperror.NewForBadRequestWithSingleField("include_owned", "Invalid boolean value for include_owned parameter")
		}
		req.IncludeOwned = includeOwned
	}

	// Parse include_shared parameter
	if includeSharedStr := r.URL.Query().Get("include_shared"); includeSharedStr != "" {
		includeShared, err := strconv.ParseBool(includeSharedStr)
		if err != nil {
			h.logger.Warn("Invalid include_shared parameter",
				zap.String("value", includeSharedStr),
				zap.Error(err))
			return nil, httperror.NewForBadRequestWithSingleField("include_shared", "Invalid boolean value for include_shared parameter")
		}
		req.IncludeShared = includeShared
	}

	// Validate that at least one option is enabled
	if !req.IncludeOwned && !req.IncludeShared {
		return nil, httperror.NewForBadRequestWithSingleField("filter_options", "At least one filter option (include_owned or include_shared) must be enabled")
	}

	h.logger.Debug("Parsed filter options",
		zap.Bool("include_owned", req.IncludeOwned),
		zap.Bool("include_shared", req.IncludeShared))

	return req, nil
}
