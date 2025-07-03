// cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata/list_sync_data.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListFileMetadataSyncDataUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID, cursor *dom_file.FileSyncCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_file.FileSyncResponse, error)
}

type listFileMetadataSyncDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileMetadataRepository
}

func NewListFileMetadataSyncDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileMetadataRepository,
) ListFileMetadataSyncDataUseCase {
	logger = logger.Named("ListFileMetadataSyncDataUseCase")
	return &listFileMetadataSyncDataUseCaseImpl{config, logger, repo}
}

func (uc *listFileMetadataSyncDataUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID, cursor *dom_file.FileSyncCursor, limit int64, accessibleCollectionIDs []gocql.UUID) (*dom_file.FileSyncResponse, error) {
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
		uc.logger.Warn("Failed validating list file sync data",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	uc.logger.Debug("Listing file sync data",
		zap.String("user_id", userID.String()),
		zap.Int("accessible_collections_count", len(accessibleCollectionIDs)),
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

	//
	// STEP 2: List file sync data from repository for accessible collections.
	//

	result, err := uc.repo.ListSyncData(ctx, userID, cursor, limit, accessibleCollectionIDs)
	if err != nil {
		uc.logger.Error("Failed to list file sync data from repository",
			zap.Any("error", err),
			zap.String("user_id", userID.String()))
		return nil, err
	}

	// Log the sync items for debugging
	uc.logger.Debug("File sync data retrieved from repository",
		zap.String("user_id", userID.String()),
		zap.Int("files_count", len(result.Files)),
		zap.Bool("has_more", result.HasMore))

	// Log each sync item to verify all fields are populated
	for i, item := range result.Files {
		uc.logger.Debug("File sync item",
			zap.Int("index", i),
			zap.String("file_id", item.ID.String()),
			zap.String("collection_id", item.CollectionID.String()),
			zap.Uint64("version", item.Version),
			zap.Time("modified_at", item.ModifiedAt),
			zap.String("state", item.State),
			zap.Uint64("tombstone_version", item.TombstoneVersion),
			zap.Time("tombstone_expiry", item.TombstoneExpiry),
			zap.Int64("encrypted_file_size_in_bytes", item.EncryptedFileSizeInBytes))
	}

	return result, nil
}
