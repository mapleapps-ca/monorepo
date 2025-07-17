// monorepo/cloud/backend/internal/maplefile/service/collection/get_filtered.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFilteredCollectionsRequestDTO struct {
	IncludeOwned  bool `json:"include_owned"`
	IncludeShared bool `json:"include_shared"`
}

type FilteredCollectionsResponseDTO struct {
	OwnedCollections  []*CollectionResponseDTO `json:"owned_collections"`
	SharedCollections []*CollectionResponseDTO `json:"shared_collections"`
	TotalCount        int                      `json:"total_count"`
}

type GetFilteredCollectionsService interface {
	Execute(ctx context.Context, req *GetFilteredCollectionsRequestDTO) (*FilteredCollectionsResponseDTO, error)
}

type getFilteredCollectionsServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetFilteredCollectionsService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetFilteredCollectionsService {
	logger = logger.Named("GetFilteredCollectionsService")
	return &getFilteredCollectionsServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *getFilteredCollectionsServiceImpl) Execute(ctx context.Context, req *GetFilteredCollectionsRequestDTO) (*FilteredCollectionsResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Request details are required")
	}

	e := make(map[string]string)
	if !req.IncludeOwned && !req.IncludeShared {
		e["filter_options"] = "At least one filter option (include_owned or include_shared) must be enabled"
	}
	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Create filter options
	//
	filterOptions := dom_collection.CollectionFilterOptions{
		IncludeOwned:  req.IncludeOwned,
		IncludeShared: req.IncludeShared,
		UserID:        userID,
	}

	//
	// STEP 4: Get filtered collections from repository
	//
	result, err := svc.repo.GetCollectionsWithFilter(ctx, filterOptions)
	if err != nil {
		svc.logger.Error("Failed to get filtered collections",
			zap.Any("error", err),
			zap.Any("user_id", userID),
			zap.Any("filter_options", filterOptions))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	response := &FilteredCollectionsResponseDTO{
		OwnedCollections:  make([]*CollectionResponseDTO, len(result.OwnedCollections)),
		SharedCollections: make([]*CollectionResponseDTO, len(result.SharedCollections)),
		TotalCount:        result.TotalCount,
	}

	// Map owned collections
	for i, collection := range result.OwnedCollections {
		response.OwnedCollections[i] = mapCollectionToDTO(collection)
	}

	// Map shared collections
	for i, collection := range result.SharedCollections {
		response.SharedCollections[i] = mapCollectionToDTO(collection)
	}

	svc.logger.Debug("Retrieved filtered collections successfully",
		zap.Int("owned_count", len(response.OwnedCollections)),
		zap.Int("shared_count", len(response.SharedCollections)),
		zap.Int("total_count", response.TotalCount),
		zap.Any("user_id", userID))

	return response, nil
}
