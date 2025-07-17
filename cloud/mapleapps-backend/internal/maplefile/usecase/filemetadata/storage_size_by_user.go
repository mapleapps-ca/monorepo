// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata/storage_size.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// Use case interfaces

type GetStorageSizeByUserUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) (*StorageSizeResponse, error)
}

// Use case implementations

type getStorageSizeByUserUseCaseImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileMetadataRepository
	collectionRepo dom_collection.CollectionRepository
}

// Constructors

func NewGetStorageSizeByUserUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileMetadataRepository,
	collectionRepo dom_collection.CollectionRepository,
) GetStorageSizeByUserUseCase {
	logger = logger.Named("GetStorageSizeByUserUseCase")
	return &getStorageSizeByUserUseCaseImpl{config, logger, fileRepo, collectionRepo}
}

// Use case implementations

func (uc *getStorageSizeByUserUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) (*StorageSizeResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage size by user",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get accessible collections for the user.
	//

	filterOptions := dom_collection.CollectionFilterOptions{
		UserID:        userID,
		IncludeOwned:  true,
		IncludeShared: true,
	}

	collectionResult, err := uc.collectionRepo.GetCollectionsWithFilter(ctx, filterOptions)
	if err != nil {
		uc.logger.Error("Failed to get accessible collections for storage size calculation",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	// Extract collection IDs
	allCollections := collectionResult.GetAllCollections()
	accessibleCollectionIDs := make([]gocql.UUID, 0, len(allCollections))
	for _, collection := range allCollections {
		accessibleCollectionIDs = append(accessibleCollectionIDs, collection.ID)
	}

	//
	// STEP 3: Calculate storage size.
	//

	totalSize, err := uc.fileRepo.GetTotalStorageSizeByUser(ctx, userID, accessibleCollectionIDs)
	if err != nil {
		uc.logger.Error("Failed to get storage size by user",
			zap.String("user_id", userID.String()),
			zap.Int("accessible_collections", len(accessibleCollectionIDs)),
			zap.Error(err))
		return nil, err
	}

	response := &StorageSizeResponse{
		TotalSizeBytes: totalSize,
	}

	uc.logger.Debug("Successfully calculated storage size by user",
		zap.String("user_id", userID.String()),
		zap.Int("accessible_collections", len(accessibleCollectionIDs)),
		zap.Int64("total_size_bytes", totalSize))

	return response, nil
}
