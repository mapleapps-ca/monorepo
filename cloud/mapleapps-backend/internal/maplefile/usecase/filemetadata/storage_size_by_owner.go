// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata/storage_size.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// StorageSizeResponse contains storage size information
type StorageSizeResponse struct {
	TotalSizeBytes int64 `json:"total_size_bytes"`
}

// Use case interfaces
type GetStorageSizeByOwnerUseCase interface {
	Execute(ctx context.Context, ownerID gocql.UUID) (*StorageSizeResponse, error)
}

// Use case implementations
type getStorageSizeByOwnerUseCaseImpl struct {
	config   *config.Configuration
	logger   *zap.Logger
	fileRepo dom_file.FileMetadataRepository
}

// Constructors
func NewGetStorageSizeByOwnerUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileMetadataRepository,
) GetStorageSizeByOwnerUseCase {
	logger = logger.Named("GetStorageSizeByOwnerUseCase")
	return &getStorageSizeByOwnerUseCaseImpl{config, logger, fileRepo}
}

// Use case implementations

func (uc *getStorageSizeByOwnerUseCaseImpl) Execute(ctx context.Context, ownerID gocql.UUID) (*StorageSizeResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if ownerID.String() == "" {
		e["owner_id"] = "Owner ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get storage size by owner",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Calculate storage size.
	//

	totalSize, err := uc.fileRepo.GetTotalStorageSizeByOwner(ctx, ownerID)
	if err != nil {
		uc.logger.Error("Failed to get storage size by owner",
			zap.String("owner_id", ownerID.String()),
			zap.Error(err))
		return nil, err
	}

	response := &StorageSizeResponse{
		TotalSizeBytes: totalSize,
	}

	uc.logger.Debug("Successfully calculated storage size by owner",
		zap.String("owner_id", ownerID.String()),
		zap.Int64("total_size_bytes", totalSize))

	return response, nil
}
