// cloud/backend/internal/maplefile/service/file/get_sync_data.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetFileSyncDataService interface {
	Execute(ctx context.Context, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error)
}

type getFileSyncDataServiceImpl struct {
	config                 *config.Configuration
	logger                 *zap.Logger
	getFileSyncDataUseCase uc_filemetadata.GetFileMetadataSyncDataUseCase
	collectionRepository   dom_collection.CollectionRepository // ADD: Collection repository
}

func NewGetFileSyncDataService(
	config *config.Configuration,
	logger *zap.Logger,
	getFileSyncDataUseCase uc_filemetadata.GetFileMetadataSyncDataUseCase,
	collectionRepository dom_collection.CollectionRepository, // ADD: Collection repository
) GetFileSyncDataService {
	logger = logger.Named("GetFileSyncDataService")
	return &getFileSyncDataServiceImpl{
		config:                 config,
		logger:                 logger,
		getFileSyncDataUseCase: getFileSyncDataUseCase,
		collectionRepository:   collectionRepository, // ADD: Collection repository
	}
}

func (svc *getFileSyncDataServiceImpl) Execute(ctx context.Context, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 2: Get accessible collections for the user
	//
	svc.logger.Debug("Getting accessible collections for file sync",
		zap.String("user_id", userID.Hex()))

	// Get collections where user is owner
	ownedCollections, err := svc.collectionRepository.GetAllByUserID(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get owned collections",
			zap.String("user_id", userID.Hex()),
			zap.Error(err))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to get accessible collections")
	}

	// Get collections shared with user
	sharedCollections, err := svc.collectionRepository.GetCollectionsSharedWithUser(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get shared collections",
			zap.String("user_id", userID.Hex()),
			zap.Error(err))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to get accessible collections")
	}

	// Combine owned and shared collections
	var accessibleCollectionIDs []primitive.ObjectID
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
		zap.String("user_id", userID.Hex()),
		zap.Int("owned_count", len(ownedCollections)),
		zap.Int("shared_count", len(sharedCollections)),
		zap.Int("total_accessible", len(accessibleCollectionIDs)))

	// If no accessible collections, return empty response
	if len(accessibleCollectionIDs) == 0 {
		svc.logger.Info("User has no accessible collections for file sync",
			zap.String("user_id", userID.Hex()))
		return &dom_file.FileSyncResponse{
			Files:      []dom_file.FileSyncItem{},
			NextCursor: nil,
			HasMore:    false,
		}, nil
	}

	//
	// STEP 3: Get file sync data for accessible collections
	//
	syncData, err := svc.getFileSyncDataUseCase.Execute(ctx, userID, cursor, limit, accessibleCollectionIDs)
	if err != nil {
		svc.logger.Error("Failed to get file sync data",
			zap.Any("error", err),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	if syncData == nil {
		svc.logger.Debug("File sync data not found",
			zap.String("user_id", userID.Hex()))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File sync results not found")
	}

	svc.logger.Debug("File sync data successfully retrieved",
		zap.String("user_id", userID.Hex()),
		zap.Any("next_cursor.", syncData.NextCursor),
		zap.Int("files_count", len(syncData.Files)))

	return syncData, nil
}
