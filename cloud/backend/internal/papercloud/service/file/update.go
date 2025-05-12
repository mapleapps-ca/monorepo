// cloud/backend/internal/papercloud/service/file/update.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UpdateFileRequestDTO struct {
	ID                 string                `json:"id"`
	EncryptedMetadata  string                `json:"encrypted_metadata,omitempty"`
	EncryptedFileKey   keys.EncryptedFileKey `json:"encrypted_file_key,omitempty"`
	EncryptedThumbnail string                `json:"encrypted_thumbnail,omitempty"`
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

	e := make(map[string]string)
	if req.ID == "" {
		e["id"] = "File ID is required"
	}

	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
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
			zap.String("file_id", req.ID))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found",
			zap.String("file_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	//
	// STEP 4: Check if user has rights to update this file
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(
		file.CollectionID,
		userID.Hex(),
		dom_collection.CollectionPermissionReadWrite,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.String("collection_id", file.CollectionID),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file update attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("file_id", req.ID))
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
	if req.EncryptedThumbnail != "" {
		file.EncryptedThumbnail = req.EncryptedThumbnail
	}

	//
	// STEP 6: Save updated file
	//
	err = svc.fileRepo.Update(file)
	if err != nil {
		svc.logger.Error("Failed to update file",
			zap.Any("error", err),
			zap.String("file_id", file.ID))
		return nil, err
	}

	//
	// STEP 7: Map domain model to response DTO
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
		EncryptedThumbnail:    file.EncryptedThumbnail,
		CreatedAt:             file.CreatedAt,
		ModifiedAt:            file.ModifiedAt,
	}

	svc.logger.Debug("File updated successfully",
		zap.String("file_id", file.ID))

	return response, nil
}
