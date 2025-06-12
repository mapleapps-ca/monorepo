// monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto/list.go
package collectiondto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// ListCollectionsFromCloudUseCase defines the interface for listing collections from cloud
type ListCollectionsFromCloudUseCase interface {
	Execute(ctx context.Context, filter collectiondto.CollectionFilter) ([]*collectiondto.CollectionDTO, error)
}

// listCollectionsFromCloudUseCase implements the ListCollectionsFromCloudUseCase interface
type listCollectionsFromCloudUseCase struct {
	logger     *zap.Logger
	repository collectiondto.CollectionDTORepository
}

// NewListCollectionsFromCloudUseCase creates a new use case for listing collections from cloud
func NewListCollectionsFromCloudUseCase(
	logger *zap.Logger,
	repository collectiondto.CollectionDTORepository,
) ListCollectionsFromCloudUseCase {
	logger = logger.Named("ListCollectionsFromCloudUseCase")
	return &listCollectionsFromCloudUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists collections from the cloud with optional filtering
func (uc *listCollectionsFromCloudUseCase) Execute(ctx context.Context, filter collectiondto.CollectionFilter) ([]*collectiondto.CollectionDTO, error) {
	//
	// STEP 1: Validate the input
	//

	e := make(map[string]string)

	// Validate parent ID format if provided
	if filter.ParentID != nil && filter.ParentID.String() == "" {
		e["parent_id"] = "Parent ID cannot be empty if provided"
	}

	// Validate collection type if provided
	if filter.CollectionType != "" {
		// Validate against known collection types
		validTypes := map[string]bool{
			"folder": true,
			"album":  true,
		}
		if !validTypes[filter.CollectionType] {
			e["collection_type"] = "Collection type must be either 'folder' or 'album'"
		}
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Call repository to list collections from cloud
	//

	uc.logger.Debug("Executing list collections from cloud use case",
		zap.Any("filter", filter))

	collections, err := uc.repository.ListFromCloud(ctx, filter)
	if err != nil {
		uc.logger.Error("Failed to list collections from cloud",
			zap.Error(err),
			zap.Any("filter", filter))
		return nil, errors.NewAppError("failed to list collections from the cloud", err)
	}

	//
	// STEP 3: Return collections response
	//

	uc.logger.Info("Successfully retrieved collections list from cloud",
		zap.Int("count", len(collections)),
		zap.Any("filter", filter))

	// Return empty slice instead of nil for consistency
	if collections == nil {
		collections = make([]*collectiondto.CollectionDTO, 0)
	}

	return collections, nil
}
