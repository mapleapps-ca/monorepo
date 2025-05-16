// cloud/backend/internal/maplefile/service/file/get_data.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileDataService interface {
	Execute(sessCtx context.Context, fileID string) ([]byte, error)
}

type getFileDataServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewGetFileDataService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) GetFileDataService {
	return &getFileDataServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *getFileDataServiceImpl) Execute(sessCtx context.Context, fileID string) ([]byte, error) {
	//
	// STEP 1: Validation
	//
	if fileID == "" {
		svc.logger.Warn("Empty file ID")
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Retrieve existing file
	//
	file, err := svc.fileRepo.Get(fileID)
	if err != nil {
		svc.logger.Error("Failed to get file",
			zap.Any("error", err),
			zap.String("file_id", fileID))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found",
			zap.String("file_id", fileID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	//
	// STEP 4: Check if user has rights to download this file
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(
		file.CollectionID,
		userID.Hex(),
		dom_collection.CollectionPermissionReadOnly,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.String("collection_id", file.CollectionID),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file data download attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("file_id", fileID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to download this file")
	}

	//
	// STEP 5: Get encrypted data
	//
	encryptedData, err := svc.fileRepo.GetEncryptedData(fileID)
	if err != nil {
		svc.logger.Error("Failed to get encrypted file data",
			zap.Any("error", err),
			zap.String("file_id", fileID))
		return nil, err
	}

	if encryptedData == nil || len(encryptedData) == 0 {
		svc.logger.Warn("No data found for file",
			zap.String("file_id", fileID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "No data found for this file")
	}

	svc.logger.Info("File data retrieved successfully",
		zap.String("file_id", fileID),
		zap.Int("size", len(encryptedData)))

	return encryptedData, nil
}
