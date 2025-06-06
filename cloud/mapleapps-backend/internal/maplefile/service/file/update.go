// cloud/backend/internal/maplefile/service/file/update.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UpdateFileRequestDTO struct {
	ID                primitive.ObjectID    `json:"id"`
	EncryptedMetadata string                `json:"encrypted_metadata,omitempty"`
	EncryptedFileKey  keys.EncryptedFileKey `json:"encrypted_file_key,omitempty"`
	EncryptionVersion string                `json:"encryption_version,omitempty"`
	EncryptedHash     string                `json:"encrypted_hash,omitempty"`
	Version           uint64                `json:"version,omitempty"`
}

type UpdateFileService interface {
	Execute(ctx context.Context, req *UpdateFileRequestDTO) (*FileResponseDTO, error)
}

type updateFileServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	collectionRepo        dom_collection.CollectionRepository
	getMetadataUseCase    uc_filemetadata.GetFileMetadataUseCase
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase
}

func NewUpdateFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase,
) UpdateFileService {
	logger = logger.Named("UpdateFileService")
	return &updateFileServiceImpl{
		config:                config,
		logger:                logger,
		collectionRepo:        collectionRepo,
		getMetadataUseCase:    getMetadataUseCase,
		updateMetadataUseCase: updateMetadataUseCase,
	}
}

func (svc *updateFileServiceImpl) Execute(ctx context.Context, req *UpdateFileRequestDTO) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File update details are required")
	}

	if req.ID.IsZero() {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID is required")
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
	// STEP 3: Get existing file metadata
	//
	file, err := svc.getMetadataUseCase.Execute(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.ID))
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
		svc.logger.Warn("Unauthorized file update attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.ID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to update this file")
	}

	//
	// STEP 5: Check if submitted collection request is in-sync with our backend's collection copy.
	//

	// Developers note:
	// What is the purpose of this check?
	// Our server has multiple clients sharing data and hence our backend needs to ensure that the file being updated is the most recent version.
	if file.Version != req.Version {
		svc.logger.Warn("Outdated collection update attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.ID),
			zap.Any("submitted_version", req.Version),
			zap.Any("current_version", file.Version))
		return nil, httperror.NewForBadRequestWithSingleField("message", "Collection has been updated since you last fetched it")
	}

	//
	// STEP 6: Update file metadata
	//
	updated := false

	if req.EncryptedMetadata != "" {
		file.EncryptedMetadata = req.EncryptedMetadata
		updated = true
	}
	if req.EncryptedFileKey.Ciphertext != nil && len(req.EncryptedFileKey.Ciphertext) > 0 {
		file.EncryptedFileKey = req.EncryptedFileKey
		updated = true
	}
	if req.EncryptionVersion != "" {
		file.EncryptionVersion = req.EncryptionVersion
		updated = true
	}
	if req.EncryptedHash != "" {
		file.EncryptedHash = req.EncryptedHash
		updated = true
	}

	if !updated {
		svc.logger.Warn("No fields to update provided")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "At least one field must be provided for update")
	}

	file.Version++ // Mutation means we increment version.
	file.ModifiedAt = time.Now()
	file.ModifiedByUserID = userID

	//
	// STEP 6: Save updated file
	//
	err = svc.updateMetadataUseCase.Execute(ctx, file)
	if err != nil {
		svc.logger.Error("Failed to update file metadata",
			zap.Any("error", err),
			zap.Any("file_id", file.ID))
		return nil, err
	}

	//
	// STEP 7: Map domain model to response DTO
	//
	response := mapFileToDTO(file)

	svc.logger.Debug("File updated successfully",
		zap.Any("file_id", file.ID))

	return response, nil
}
