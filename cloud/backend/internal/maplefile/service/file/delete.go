// cloud/backend/internal/maplefile/service/file/delete.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteFileRequestDTO struct {
	FileID primitive.ObjectID `json:"file_id"`
}

type DeleteFileResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteFileService interface {
	Execute(ctx context.Context, req *DeleteFileRequestDTO) (*DeleteFileResponseDTO, error)
}

type deleteFileServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	collectionRepo        dom_collection.CollectionRepository
	getMetadataUseCase    uc_filemetadata.GetFileMetadataUseCase
	deleteMetadataUseCase uc_filemetadata.DeleteFileMetadataUseCase
	deleteDataUseCase     uc_fileobjectstorage.DeleteEncryptedDataUseCase
}

func NewDeleteFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	deleteMetadataUseCase uc_filemetadata.DeleteFileMetadataUseCase,
	deleteDataUseCase uc_fileobjectstorage.DeleteEncryptedDataUseCase,
) DeleteFileService {
	return &deleteFileServiceImpl{
		config:                config,
		logger:                logger,
		collectionRepo:        collectionRepo,
		getMetadataUseCase:    getMetadataUseCase,
		deleteMetadataUseCase: deleteMetadataUseCase,
		deleteDataUseCase:     deleteDataUseCase,
	}
}

func (svc *deleteFileServiceImpl) Execute(ctx context.Context, req *DeleteFileRequestDTO) (*DeleteFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File ID is required")
	}

	if req.FileID.IsZero() {
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
	file, err := svc.getMetadataUseCase.Execute(req.FileID)
	if err != nil {
		svc.logger.Error("Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 4: Check if user has write access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file deletion attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to delete this file")
	}

	//
	// STEP 5: Delete encrypted file data from object storage
	//
	err = svc.deleteDataUseCase.Execute(file.EncryptedFileObjectKey)
	if err != nil {
		svc.logger.Error("Failed to delete encrypted file data",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		// Continue with metadata deletion even if object storage deletion fails
	}

	// Delete thumbnail if it exists
	if file.EncryptedThumbnailObjectKey != "" {
		err = svc.deleteDataUseCase.Execute(file.EncryptedThumbnailObjectKey)
		if err != nil {
			svc.logger.Warn("Failed to delete encrypted thumbnail data",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID),
				zap.String("thumbnail_storage_path", file.EncryptedThumbnailObjectKey))
		}
	}

	//
	// STEP 6: Delete file metadata
	//
	err = svc.deleteMetadataUseCase.Execute(req.FileID)
	if err != nil {
		svc.logger.Error("Failed to delete file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	svc.logger.Info("File deleted successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("collection_id", file.CollectionID),
		zap.Any("user_id", userID))

	return &DeleteFileResponseDTO{
		Success: true,
		Message: "File deleted successfully",
	}, nil
}
