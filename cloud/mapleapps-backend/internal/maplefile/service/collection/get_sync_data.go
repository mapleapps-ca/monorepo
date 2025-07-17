// monorepo/cloud/backend/internal/maplefile/service/collection/get.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	uc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetCollectionSyncDataService interface {
	Execute(ctx context.Context, userID gocql.UUID, cursor *dom_collection.CollectionSyncCursor, limit int64, accessType string) (*dom_collection.CollectionSyncResponse, error)
}

type getCollectionSyncDataServiceImpl struct {
	config                       *config.Configuration
	logger                       *zap.Logger
	getCollectionSyncDataUseCase uc_collection.GetCollectionSyncDataUseCase
}

func NewGetCollectionSyncDataService(
	config *config.Configuration,
	logger *zap.Logger,
	getCollectionSyncDataUseCase uc_collection.GetCollectionSyncDataUseCase,
) GetCollectionSyncDataService {
	logger = logger.Named("GetCollectionSyncDataService")
	return &getCollectionSyncDataServiceImpl{
		config:                       config,
		logger:                       logger,
		getCollectionSyncDataUseCase: getCollectionSyncDataUseCase,
	}
}

func (svc *getCollectionSyncDataServiceImpl) Execute(ctx context.Context, userID gocql.UUID, cursor *dom_collection.CollectionSyncCursor, limit int64, accessType string) (*dom_collection.CollectionSyncResponse, error) {
	//
	// STEP 1: Validation
	//
	if userID.String() == "" {
		svc.logger.Warn("Empty user ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("user_id", "User ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get related data based on access type.
	//

	syncData, err := svc.getCollectionSyncDataUseCase.Execute(ctx, userID, cursor, limit, accessType)
	if err != nil {
		svc.logger.Error("Failed to get collection sync data",
			zap.Any("error", err),
			zap.Any("user_id", userID))
		return nil, err
	}

	if syncData == nil {
		svc.logger.Debug("Collection sync data not found",
			zap.Any("user_id", userID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection sync results not found")
	}

	// //
	// // STEP 4: Check if the user has access to this collection
	// //
	// // First check if user is owner
	// hasAccess := collection.OwnerID == userID

	// // If not owner, check if user is a member
	// if !hasAccess {
	// 	for _, member := range collection.Members {
	// 		if member.RecipientID == userID {
	// 			hasAccess = true
	// 			break
	// 		}
	// 	}
	// }

	// if !hasAccess {
	// 	svc.logger.Warn("Unauthorized collection access attempt",
	// 		zap.Any("user_id", userID),
	// 		zap.Any("collection_id", collectionID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this collection")
	// }
	//
	svc.logger.Debug("Collection sync data successfully retrieved",
		zap.Any("user_id", userID),
		zap.Any("sync_data", syncData))

	return syncData, nil
}
