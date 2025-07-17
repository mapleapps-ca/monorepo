// monorepo/cloud/backend/internal/maplefile/service/file/list_sync_data.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListFileSyncDataService interface {
	Execute(ctx context.Context, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error)
}

type listFileSyncDataServiceImpl struct {
	config                  *config.Configuration
	logger                  *zap.Logger
	listFileSyncDataUseCase uc_filemetadata.ListFileMetadataSyncDataUseCase
	collectionRepository    dom_collection.CollectionRepository
}

func NewListFileSyncDataService(
	config *config.Configuration,
	logger *zap.Logger,
	listFileSyncDataUseCase uc_filemetadata.ListFileMetadataSyncDataUseCase,
	collectionRepository dom_collection.CollectionRepository,
) ListFileSyncDataService {
	logger = logger.Named("ListFileSyncDataService")
	return &listFileSyncDataServiceImpl{
		config:                  config,
		logger:                  logger,
		listFileSyncDataUseCase: listFileSyncDataUseCase,
		collectionRepository:    collectionRepository,
	}
}

func (svc *listFileSyncDataServiceImpl) Execute(ctx context.Context, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 2: Get accessible collections for the user
	//
	svc.logger.Debug("Getting accessible collections for file sync",
		zap.String("user_id", userID.String()))

	// Get collections where user is owner
	ownedCollections, err := svc.collectionRepository.GetAllByUserID(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get owned collections",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to get accessible collections")
	}

	// Get collections shared with user
	sharedCollections, err := svc.collectionRepository.GetCollectionsSharedWithUser(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get shared collections",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to get accessible collections")
	}

	// Combine owned and shared collections
	var accessibleCollectionIDs []gocql.UUID
	for _, coll := range ownedCollections {
		if coll.State == "active" { // Only include active collections
			accessibleCollectionIDs = append(accessibleCollectionIDs, coll.ID)
		}
	}
	for _, coll := range sharedCollections {
		if coll.State == "active" { // Only include active collections
			accessibleCollectionIDs = append(accessibleCollectionIDs, coll.ID)
		}
	}

	svc.logger.Debug("Found accessible collections for file sync",
		zap.String("user_id", userID.String()),
		zap.Int("owned_count", len(ownedCollections)),
		zap.Int("shared_count", len(sharedCollections)),
		zap.Int("total_accessible", len(accessibleCollectionIDs)))

	// If no accessible collections, return empty response
	if len(accessibleCollectionIDs) == 0 {
		svc.logger.Info("User has no accessible collections for file sync",
			zap.String("user_id", userID.String()))
		return &dom_file.FileSyncResponse{
			Files:      []dom_file.FileSyncItem{},
			NextCursor: nil,
			HasMore:    false,
		}, nil
	}

	//
	// STEP 3: List file sync data for accessible collections
	//
	syncData, err := svc.listFileSyncDataUseCase.Execute(ctx, userID, cursor, limit, accessibleCollectionIDs)
	if err != nil {
		svc.logger.Error("Failed to list file sync data",
			zap.Any("error", err),
			zap.String("user_id", userID.String()))
		return nil, err
	}

	if syncData == nil {
		svc.logger.Debug("File sync data not found",
			zap.String("user_id", userID.String()))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File sync results not found")
	}

	// Log sync data with all fields including EncryptedFileSizeInBytes
	svc.logger.Debug("File sync data successfully retrieved",
		zap.String("user_id", userID.String()),
		zap.Any("next_cursor", syncData.NextCursor),
		zap.Int("files_count", len(syncData.Files)))

	// Verify each item has all fields populated including EncryptedFileSizeInBytes
	for i, item := range syncData.Files {
		svc.logger.Debug("Returning file sync item",
			zap.Int("index", i),
			zap.String("file_id", item.ID.String()),
			zap.String("collection_id", item.CollectionID.String()),
			zap.Uint64("version", item.Version),
			zap.String("state", item.State),
			zap.Int64("encrypted_file_size_in_bytes", item.EncryptedFileSizeInBytes))
	}

	return syncData, nil
}
