// cloud/backend/internal/maplefile/usecase/collection/get_filtered.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFilteredCollectionsUseCase interface {
	Execute(ctx context.Context, options dom_collection.CollectionFilterOptions) (*dom_collection.CollectionFilterResult, error)
}

type getFilteredCollectionsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetFilteredCollectionsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetFilteredCollectionsUseCase {
	logger = logger.Named("GetFilteredCollectionsUseCase")
	return &getFilteredCollectionsUseCaseImpl{config, logger, repo}
}

func (uc *getFilteredCollectionsUseCaseImpl) Execute(ctx context.Context, options dom_collection.CollectionFilterOptions) (*dom_collection.CollectionFilterResult, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if options.UserID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if !options.IsValid() {
		e["filter_options"] = "At least one filter option (include_owned or include_shared) must be enabled"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get filtered collections",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get filtered collections from repository.
	//

	result, err := uc.repo.GetCollectionsWithFilter(ctx, options)
	if err != nil {
		uc.logger.Error("Failed to get filtered collections from repository",
			zap.Any("error", err),
			zap.Any("options", options))
		return nil, err
	}

	uc.logger.Debug("Successfully retrieved filtered collections",
		zap.Int("owned_count", len(result.OwnedCollections)),
		zap.Int("shared_count", len(result.SharedCollections)),
		zap.Int("total_count", result.TotalCount),
		zap.Any("user_id", options.UserID))

	return result, nil
}
