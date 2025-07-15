// cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata/storage_size.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// StorageSizeBreakdownResponse contains detailed storage breakdown
type StorageSizeBreakdownResponse struct {
	OwnedSizeBytes           int64            `json:"owned_size_bytes"`
	SharedSizeBytes          int64            `json:"shared_size_bytes"`
	TotalSizeBytes           int64            `json:"total_size_bytes"`
	CollectionBreakdownBytes map[string]int64 `json:"collection_breakdown_bytes"`
	OwnedCollectionsCount    int              `json:"owned_collections_count"`
	SharedCollectionsCount   int              `json:"shared_collections_count"`
}

// Use case interfaces

type GetStorageSizeByCollectionUseCase interface {
	Execute(ctx context.Context, collectionID gocql.UUID) (*StorageSizeResponse, error)
}

// Use case implementations

type getStorageSizeByCollectionUseCaseImpl struct {
	config   *config.Configuration
	logger   *zap.Logger
	fileRepo dom_file.FileMetadataRepository
}

// Constructors

func NewGetStorageSizeByCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileMetadataRepository,
) GetStorageSizeByCollectionUseCase {
	logger = logger.Named("GetStorageSizeByCollectionUseCase")
	return &getStorageSizeByCollectionUseCaseImpl{config, logger, fileRepo}
}

// Use case implementations

func (uc *getStorageSizeByCollectionUseCaseImpl) Execute(ctx context.Context, collectionID gocql.UUID) (*StorageSizeResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID.String() == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage size by collection",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Calculate storage size.
	//

	totalSize, err := uc.fileRepo.GetTotalStorageSizeByCollection(ctx, collectionID)
	if err != nil {
		uc.logger.Error("Failed to get storage size by collection",
			zap.String("collection_id", collectionID.String()),
			zap.Error(err))
		return nil, err
	}

	response := &StorageSizeResponse{
		TotalSizeBytes: totalSize,
	}

	uc.logger.Debug("Successfully calculated storage size by collection",
		zap.String("collection_id", collectionID.String()),
		zap.Int64("total_size_bytes", totalSize))

	return response, nil
}
