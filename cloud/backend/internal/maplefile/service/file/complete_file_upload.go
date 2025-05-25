// cloud/backend/internal/maplefile/service/file/complete_file_upload.go
package file

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CompleteFileUploadRequestDTO struct {
	FileID primitive.ObjectID `json:"file_id"`
	// Optional: Client can provide actual file size for validation
	ActualFileSizeInBytes int64 `json:"actual_file_size_in_bytes,omitempty"`
	// Optional: Client can provide actual thumbnail size for validation
	ActualThumbnailSizeInBytes int64 `json:"actual_thumbnail_size_in_bytes,omitempty"`
	// Optional: Client can confirm successful upload
	UploadConfirmed          bool `json:"upload_confirmed,omitempty"`
	ThumbnailUploadConfirmed bool `json:"thumbnail_upload_confirmed,omitempty"`
}

type CompleteFileUploadResponseDTO struct {
	File                *FileResponseDTO `json:"file"`
	Success             bool             `json:"success"`
	Message             string           `json:"message"`
	ActualFileSize      int64            `json:"actual_file_size"`
	ActualThumbnailSize int64            `json:"actual_thumbnail_size"`
	UploadVerified      bool             `json:"upload_verified"`
	ThumbnailVerified   bool             `json:"thumbnail_verified"`
}

type CompleteFileUploadService interface {
	Execute(ctx context.Context, req *CompleteFileUploadRequestDTO) (*CompleteFileUploadResponseDTO, error)
}

type completeFileUploadServiceImpl struct {
	config                    *config.Configuration
	logger                    *zap.Logger
	collectionRepo            dom_collection.CollectionRepository
	getMetadataUseCase        uc_filemetadata.GetFileMetadataUseCase
	updateMetadataUseCase     uc_filemetadata.UpdateFileMetadataUseCase
	verifyObjectExistsUseCase uc_fileobjectstorage.VerifyObjectExistsUseCase
	getObjectSizeUseCase      uc_fileobjectstorage.GetObjectSizeUseCase
	deleteDataUseCase         uc_fileobjectstorage.DeleteEncryptedDataUseCase
}

func NewCompleteFileUploadService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	updateMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase,
	verifyObjectExistsUseCase uc_fileobjectstorage.VerifyObjectExistsUseCase,
	getObjectSizeUseCase uc_fileobjectstorage.GetObjectSizeUseCase,
	deleteDataUseCase uc_fileobjectstorage.DeleteEncryptedDataUseCase,
) CompleteFileUploadService {
	return &completeFileUploadServiceImpl{
		config:                    config,
		logger:                    logger,
		collectionRepo:            collectionRepo,
		getMetadataUseCase:        getMetadataUseCase,
		updateMetadataUseCase:     updateMetadataUseCase,
		verifyObjectExistsUseCase: verifyObjectExistsUseCase,
		getObjectSizeUseCase:      getObjectSizeUseCase,
		deleteDataUseCase:         deleteDataUseCase,
	}
}

func (svc *completeFileUploadServiceImpl) Execute(ctx context.Context, req *CompleteFileUploadRequestDTO) (*CompleteFileUploadResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("⚠️ Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File completion details are required")
	}

	if req.FileID.IsZero() {
		svc.logger.Warn("⚠️ Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("🔴 Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file metadata
	//
	file, err := svc.getMetadataUseCase.Execute(req.FileID)
	if err != nil {
		svc.logger.Error("🔴 Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 4: Verify user has write access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("🔴 Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("⚠️ Unauthorized file completion attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to complete this file upload")
	}

	//
	// STEP 5: Verify file is in pending state
	//
	if file.State != dom_file.StatePending {
		svc.logger.Warn("⚠️ File is not in pending state",
			zap.Any("file_id", req.FileID),
			zap.String("current_state", file.State))
		return nil, httperror.NewForBadRequestWithSingleField("file_id", fmt.Sprintf("File is not in pending state (current state: %s)", file.State))
	}

	//
	// STEP 6: Verify file exists in object storage and get actual size
	//
	fileExists, err := svc.verifyObjectExistsUseCase.Execute(file.EncryptedFileObjectKey)
	if err != nil {
		svc.logger.Error("🔴 Failed to verify file exists in storage",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to verify file upload")
	}

	if !fileExists {
		svc.logger.Warn("⚠️ File does not exist in storage",
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File has not been uploaded yet")
	}

	// Get actual file size from storage
	actualFileSize, err := svc.getObjectSizeUseCase.Execute(file.EncryptedFileObjectKey)
	if err != nil {
		svc.logger.Error("🔴 Failed to get file size from storage",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to verify file size")
	}

	//
	// STEP 7: Verify thumbnail if expected
	//
	var actualThumbnailSize int64 = 0
	var thumbnailVerified bool = true

	if file.EncryptedThumbnailObjectKey != "" {
		thumbnailExists, err := svc.verifyObjectExistsUseCase.Execute(file.EncryptedThumbnailObjectKey)
		if err != nil {
			svc.logger.Warn("⚠️ Failed to verify thumbnail exists, continuing without it",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID),
				zap.String("thumbnail_storage_path", file.EncryptedThumbnailObjectKey))
			thumbnailVerified = false
		} else if thumbnailExists {
			actualThumbnailSize, err = svc.getObjectSizeUseCase.Execute(file.EncryptedThumbnailObjectKey)
			if err != nil {
				svc.logger.Warn("⚠️ Failed to get thumbnail size, continuing without it",
					zap.Any("error", err),
					zap.Any("file_id", req.FileID),
					zap.String("thumbnail_storage_path", file.EncryptedThumbnailObjectKey))
				thumbnailVerified = false
			}
		} else {
			// Thumbnail was expected but not uploaded - clear the path
			file.EncryptedThumbnailObjectKey = ""
			thumbnailVerified = false
		}
	}

	//
	// STEP 8: Validate file size if client provided it
	//
	if req.ActualFileSizeInBytes > 0 && req.ActualFileSizeInBytes != actualFileSize {
		svc.logger.Warn("⚠️ File size mismatch between client and storage",
			zap.Any("file_id", req.FileID),
			zap.Int64("client_reported_size", req.ActualFileSizeInBytes),
			zap.Int64("storage_actual_size", actualFileSize))
		// Continue with storage size as authoritative
	}

	//
	// STEP 9: Update file metadata to active state
	//
	file.EncryptedFileSizeInBytes = actualFileSize
	file.EncryptedThumbnailSizeInBytes = actualThumbnailSize
	file.State = dom_file.StateActive
	file.ModifiedAt = time.Now()
	file.ModifiedByUserID = userID
	file.Version++

	err = svc.updateMetadataUseCase.Execute(file)
	if err != nil {
		svc.logger.Error("🔴 Failed to update file metadata to active state",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 10: Prepare response
	//
	response := &CompleteFileUploadResponseDTO{
		File:                mapFileToDTO(file),
		Success:             true,
		Message:             "File upload completed successfully",
		ActualFileSize:      actualFileSize,
		ActualThumbnailSize: actualThumbnailSize,
		UploadVerified:      true,
		ThumbnailVerified:   thumbnailVerified,
	}

	svc.logger.Info("✅ File upload completed successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("collection_id", file.CollectionID),
		zap.Any("owner_id", userID),
		zap.Int64("actual_file_size", actualFileSize),
		zap.Int64("actual_thumbnail_size", actualThumbnailSize),
		zap.Bool("thumbnail_verified", thumbnailVerified))

	return response, nil
}
