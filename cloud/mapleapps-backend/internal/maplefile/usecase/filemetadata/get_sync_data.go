// cloud/backend/internal/maplefile/usecase/filemetadata/get_sync_data.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFileMetadataSyncDataUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID, cursor *dom_file.FileSyncCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_file.FileSyncResponse, error)
}

type getFileMetadataSyncDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewGetFileMetadataSyncDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) GetFileMetadataSyncDataUseCase {
	logger = logger.Named("GetFileMetadataSyncDataUseCase")
	return &getFileMetadataSyncDataUseCaseImpl{config, logger, repo}
}

func (uc *getFileMetadataSyncDataUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID, cursor *dom_file.FileSyncCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_file.FileSyncResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(accessibleCollectionIDs) == 0 {
		e["accessible_collections"] = "At least one accessible collection is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get file sync data",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	uc.logger.Debug("Getting file sync data",
		zap.String("user_id", userID.String()),
		zap.Int("accessible_collections_count", len(accessibleCollectionIDs)),
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

	//
	// STEP 2: Get file sync data from repository for accessible collections.
	//

	result, err := uc.repo.GetSyncData(ctx, userID, cursor, limit, accessibleCollectionIDs)
	if err != nil {
		uc.logger.Error("Failed to get file sync data from repository",
			zap.Any("error", err),
			zap.String("user_id", userID.String()))
		return nil, err
	}

	return result, nil
}
