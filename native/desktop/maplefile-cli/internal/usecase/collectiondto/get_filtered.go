// monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto/get_filtered.go
package collectiondto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// GetFilteredCollectionsFromCloudUseCase defines the interface for getting filtered collections from cloud
type GetFilteredCollectionsFromCloudUseCase interface {
	Execute(ctx context.Context, request *collectiondto.GetFilteredCollectionsRequest) (*collectiondto.GetFilteredCollectionsResponse, error)
}

// getFilteredCollectionsFromCloudUseCase implements the GetFilteredCollectionsFromCloudUseCase interface
type getFilteredCollectionsFromCloudUseCase struct {
	logger     *zap.Logger
	repository collectiondto.CollectionDTORepository
}

// NewGetFilteredCollectionsFromCloudUseCase creates a new use case for getting filtered collections from cloud
func NewGetFilteredCollectionsFromCloudUseCase(
	logger *zap.Logger,
	repository collectiondto.CollectionDTORepository,
) GetFilteredCollectionsFromCloudUseCase {
	logger = logger.Named("GetFilteredCollectionsFromCloudUseCase")
	return &getFilteredCollectionsFromCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute gets filtered collections from the cloud
func (uc *getFilteredCollectionsFromCloudUseCase) Execute(ctx context.Context, request *collectiondto.GetFilteredCollectionsRequest) (*collectiondto.GetFilteredCollectionsResponse, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	if request == nil {
		e["request"] = "Request is required"
	} else {
		// At least one filter option must be enabled
		if !request.IncludeOwned && !request.IncludeShared {
			e["filter_options"] = "At least one filter option (include_owned or include_shared) must be enabled"
		}
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Call repository to get filtered collections from cloud
	//

	uc.logger.Debug("Executing get filtered collections from cloud use case",
		zap.Bool("include_owned", request.IncludeOwned),
		zap.Bool("include_shared", request.IncludeShared))

	response, err := uc.repository.GetFilteredCollectionsFromCloud(ctx, request)
	if err != nil {
		uc.logger.Error("Failed to get filtered collections from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get filtered collections from the cloud", err)
	}

	//
	// STEP 3: Return response
	//

	uc.logger.Info("Successfully retrieved filtered collections from cloud",
		zap.Int("owned_count", len(response.OwnedCollections)),
		zap.Int("shared_count", len(response.SharedCollections)),
		zap.Int("total_count", response.TotalCount))

	return response, nil
}
