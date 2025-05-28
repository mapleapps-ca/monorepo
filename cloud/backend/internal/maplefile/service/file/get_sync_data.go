// cloud/backend/internal/maplefile/service/file/get.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileSyncDataService interface {
	Execute(ctx context.Context, userID primitive.ObjectID, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error)
}

type getFileSyncDataServiceImpl struct {
	config                 *config.Configuration
	logger                 *zap.Logger
	getFileSyncDataUseCase uc_filemetadata.GetFileMetadataSyncDataUseCase
}

func NewGetFileSyncDataService(
	config *config.Configuration,
	logger *zap.Logger,
	getFileSyncDataUseCase uc_filemetadata.GetFileMetadataSyncDataUseCase,
) GetFileSyncDataService {
	return &getFileSyncDataServiceImpl{
		config:                 config,
		logger:                 logger,
		getFileSyncDataUseCase: getFileSyncDataUseCase,
	}
}

func (svc *getFileSyncDataServiceImpl) Execute(ctx context.Context, userID primitive.ObjectID, cursor *dom_file.FileSyncCursor, limit int64) (*dom_file.FileSyncResponse, error) {
	//
	// STEP 1: Validation
	//
	// if options.UserID.IsZero() {
	// 	svc.logger.Warn("Empty user ID provided")
	// 	return nil, httperror.NewForBadRequestWithSingleField("user_id", "User ID is required")
	// }

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get related data.
	//
	syncData, err := svc.getFileSyncDataUseCase.Execute(ctx, userID, cursor, limit)
	if err != nil {
		svc.logger.Error("Failed to get file sync data",
			zap.Any("error", err),
			zap.Any("user_id", userID))
		return nil, err
	}

	if syncData == nil {
		svc.logger.Debug("File sync data not found",
			zap.Any("user_id", userID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File sync results not found")
	}

	// //
	// // STEP 4: Check if the user has access to this file
	// //
	// // First check if user is owner
	// hasAccess := file.OwnerID == userID

	// // If not owner, check if user is a member
	// if !hasAccess {
	// 	for _, member := range file.Members {
	// 		if member.RecipientID == userID {
	// 			hasAccess = true
	// 			break
	// 		}
	// 	}
	// }

	// if !hasAccess {
	// 	svc.logger.Warn("Unauthorized file access attempt",
	// 		zap.Any("user_id", userID),
	// 		zap.Any("file_id", fileID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this file")
	// }

	svc.logger.Debug("File sync data successfully retrieved",
		zap.Any("user_id", userID),
		zap.Any("sync_data", syncData))

	return syncData, nil
}
