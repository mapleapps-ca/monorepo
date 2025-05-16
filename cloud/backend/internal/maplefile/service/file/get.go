// cloud/backend/internal/maplefile/service/file/get.go
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

type GetFileService interface {
	Execute(sessCtx context.Context, fileID string) (*FileResponseDTO, error)
}

type getFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewGetFileService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) GetFileService {
	return &getFileServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *getFileServiceImpl) Execute(sessCtx context.Context, fileID string) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if fileID == "" {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
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
	// STEP 3: Get file from repository
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
	// STEP 4: Check if the user has access to the file's collection
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
		svc.logger.Warn("Unauthorized file access attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("file_id", fileID),
			zap.String("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this file")
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	response := &FileResponseDTO{
		ID:                    file.ID,
		CollectionID:          file.CollectionID,
		OwnerID:               file.OwnerID,
		FileID:                file.FileID,
		StoragePath:           file.StoragePath,
		EncryptedSize:         file.EncryptedSize,
		EncryptedOriginalSize: file.EncryptedOriginalSize,
		EncryptedMetadata:     file.EncryptedMetadata,
		EncryptionVersion:     file.EncryptionVersion,
		EncryptedHash:         file.EncryptedHash,
		EncryptedFileKey:      file.EncryptedFileKey,
		EncryptedThumbnail:    file.EncryptedThumbnail,
		CreatedAt:             file.CreatedAt,
		ModifiedAt:            file.ModifiedAt,
	}

	return response, nil
}
