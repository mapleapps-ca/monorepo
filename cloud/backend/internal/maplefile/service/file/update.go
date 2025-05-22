// cloud/backend/internal/maplefile/service/file/update.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UpdateFileRequestDTO struct {
	ID                 primitive.ObjectID    `json:"id"`
	EncryptedMetadata  string                `json:"encrypted_metadata,omitempty"`
	EncryptedFileKey   keys.EncryptedFileKey `json:"encrypted_file_key,omitempty"`
	ThumbnailObjectKey string                `json:"thumbnail_object_key,omitempty"`
}

type UpdateFileService interface {
	Execute(sessCtx context.Context, req *UpdateFileRequestDTO) (*FileResponseDTO, error)
}

type updateFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewUpdateFileService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) UpdateFileService {
	return &updateFileServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *updateFileServiceImpl) Execute(sessCtx context.Context, req *UpdateFileRequestDTO) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File details are required")
	}

	if req.ID.IsZero() {
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
	file, err := svc.fileRepo.Get(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get file",
			zap.Any("error", err),
			zap.Any("file_id", req.ID))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found",
			zap.Any("file_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	//
	// STEP 4: Check if user has rights to update this file
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(
		sessCtx,
		file.CollectionID,
		userID,
		dom_collection.CollectionPermissionReadWrite,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file update attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to update this file")
	}

	//
	// STEP 5: Update file
	//
	file.ModifiedAt = time.Now()

	// Only update optional fields if they are provided
	if req.EncryptedMetadata != "" {
		file.EncryptedMetadata = req.EncryptedMetadata
	}
	if req.EncryptedFileKey.Ciphertext != nil && len(req.EncryptedFileKey.Ciphertext) > 0 &&
		req.EncryptedFileKey.Nonce != nil && len(req.EncryptedFileKey.Nonce) > 0 {
		file.EncryptedFileKey = req.EncryptedFileKey
	}
	if req.ThumbnailObjectKey != "" {
		file.ThumbnailObjectKey = req.ThumbnailObjectKey
	}

	//
	// STEP 6: Save updated file
	//
	err = svc.fileRepo.Update(file)
	if err != nil {
		svc.logger.Error("Failed to update file",
			zap.Any("error", err),
			zap.Any("file_id", file.ID))
		return nil, err
	}

	//
	// STEP 7: Map domain model to response DTO
	//
	response := &FileResponseDTO{
		ID:                 file.ID,
		CollectionID:       file.CollectionID,
		OwnerID:            file.OwnerID,
		FileObjectKey:      file.FileObjectKey,
		EncryptedFileSize:  file.EncryptedFileSize,
		EncryptedMetadata:  file.EncryptedMetadata,
		EncryptionVersion:  file.EncryptionVersion,
		EncryptedHash:      file.EncryptedHash,
		EncryptedFileKey:   file.EncryptedFileKey,
		ThumbnailObjectKey: file.ThumbnailObjectKey,
		CreatedAt:          file.CreatedAt,
		ModifiedAt:         file.ModifiedAt,
	}

	svc.logger.Debug("File updated successfully",
		zap.Any("file_id", file.ID))

	return response, nil
}
