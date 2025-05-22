// cloud/backend/internal/maplefile/service/file/get.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetFileService interface {
	Execute(ctx context.Context, fileID primitive.ObjectID) (*FileResponseDTO, error)
}

type getFileServiceImpl struct {
	config             *config.Configuration
	logger             *zap.Logger
	collectionRepo     dom_collection.CollectionRepository
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase
}

func NewGetFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
) GetFileService {
	return &getFileServiceImpl{
		config:             config,
		logger:             logger,
		collectionRepo:     collectionRepo,
		getMetadataUseCase: getMetadataUseCase,
	}
}

func (svc *getFileServiceImpl) Execute(ctx context.Context, fileID primitive.ObjectID) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if fileID.IsZero() {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file metadata
	//
	file, err := svc.getMetadataUseCase.Execute(fileID)
	if err != nil {
		svc.logger.Error("Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", fileID))
		return nil, err
	}

	//
	// STEP 4: Check if user has access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadOnly)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file access attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", fileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to access this file")
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	response := mapFileToDTO(file)

	return response, nil
}
