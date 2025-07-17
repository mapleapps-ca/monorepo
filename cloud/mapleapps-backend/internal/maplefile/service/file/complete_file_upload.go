// monorepo/cloud/backend/internal/maplefile/service/file/complete_file_upload.go
package file

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	uc_feduser "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/usecase/federateduser"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CompleteFileUploadRequestDTO struct {
	FileID gocql.UUID `json:"file_id"`
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
	StorageAdjustment   int64            `json:"storage_adjustment"` // Positive if more space used, negative if less
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
	storageQuotaHelperUseCase uc_feduser.FederatedUserStorageQuotaHelperUseCase
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
	storageQuotaHelperUseCase uc_feduser.FederatedUserStorageQuotaHelperUseCase,
) CompleteFileUploadService {
	logger = logger.Named("CompleteFileUploadService")
	return &completeFileUploadServiceImpl{
		config:                    config,
		logger:                    logger,
		collectionRepo:            collectionRepo,
		getMetadataUseCase:        getMetadataUseCase,
		updateMetadataUseCase:     updateMetadataUseCase,
		verifyObjectExistsUseCase: verifyObjectExistsUseCase,
		getObjectSizeUseCase:      getObjectSizeUseCase,
		deleteDataUseCase:         deleteDataUseCase,
		storageQuotaHelperUseCase: storageQuotaHelperUseCase,
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

	if req.FileID.String() == "" {
		svc.logger.Warn("⚠️ Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("🔴 Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file metadata
	//
	// Developers note: Use `ExecuteWithAnyState` because initially created `FileMetadata` object has state set to `pending`.
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
	if file.State != dom_file.FileStatePending {
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
	// STEP 8: Calculate storage adjustment and update quota
	//
	expectedTotalSize := file.EncryptedFileSizeInBytes + file.EncryptedThumbnailSizeInBytes
	actualTotalSize := actualFileSize + actualThumbnailSize
	storageAdjustment := actualTotalSize - expectedTotalSize

	svc.logger.Info("📊 Storage size comparison",
		zap.Any("file_id", req.FileID),
		zap.Int64("expected_file_size", file.EncryptedFileSizeInBytes),
		zap.Int64("actual_file_size", actualFileSize),
		zap.Int64("expected_thumbnail_size", file.EncryptedThumbnailSizeInBytes),
		zap.Int64("actual_thumbnail_size", actualThumbnailSize),
		zap.Int64("expected_total", expectedTotalSize),
		zap.Int64("actual_total", actualTotalSize),
		zap.Int64("adjustment", storageAdjustment))

	// Handle storage quota adjustment
	if storageAdjustment != 0 {
		if storageAdjustment > 0 {
			// Need more quota than originally reserved
			err = svc.storageQuotaHelperUseCase.CheckAndReserveQuota(ctx, userID, storageAdjustment)
			if err != nil {
				svc.logger.Error("🔴 Failed to reserve additional storage quota",
					zap.String("user_id", userID.String()),
					zap.Int64("additional_size", storageAdjustment),
					zap.Error(err))

				// Clean up the uploaded file since we can't complete due to quota
				if deleteErr := svc.deleteDataUseCase.Execute(file.EncryptedFileObjectKey); deleteErr != nil {
					svc.logger.Error("🔴 Failed to clean up file after quota exceeded", zap.Error(deleteErr))
				}
				if file.EncryptedThumbnailObjectKey != "" {
					if deleteErr := svc.deleteDataUseCase.Execute(file.EncryptedThumbnailObjectKey); deleteErr != nil {
						svc.logger.Error("🔴 Failed to clean up thumbnail after quota exceeded", zap.Error(deleteErr))
					}
				}

				return nil, err
			}
		} else {
			// Used less quota than originally reserved, release the difference
			err = svc.storageQuotaHelperUseCase.ReleaseQuota(ctx, userID, -storageAdjustment)
			if err != nil {
				svc.logger.Warn("⚠️ Failed to release excess quota, but continuing",
					zap.String("user_id", userID.String()),
					zap.Int64("excess_size", -storageAdjustment),
					zap.Error(err))
			}
		}
	}

	//
	// STEP 9: Validate file size if client provided it
	//
	if req.ActualFileSizeInBytes > 0 && req.ActualFileSizeInBytes != actualFileSize {
		svc.logger.Warn("⚠️ File size mismatch between client and storage",
			zap.Any("file_id", req.FileID),
			zap.Int64("client_reported_size", req.ActualFileSizeInBytes),
			zap.Int64("storage_actual_size", actualFileSize))
		// Continue with storage size as authoritative
	}

	//
	// STEP 10: Update file metadata to active state
	//
	file.EncryptedFileSizeInBytes = actualFileSize
	file.EncryptedThumbnailSizeInBytes = actualThumbnailSize
	file.State = dom_file.FileStateActive
	file.ModifiedAt = time.Now()
	file.ModifiedByUserID = userID
	file.Version++ // Every mutation we need to keep a track of.

	err = svc.updateMetadataUseCase.Execute(ctx, file)
	if err != nil {
		svc.logger.Error("🔴 Failed to update file metadata to active state",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 11: Prepare response
	//
	response := &CompleteFileUploadResponseDTO{
		File:                mapFileToDTO(file),
		Success:             true,
		Message:             "File upload completed successfully with storage quota updated",
		ActualFileSize:      actualFileSize,
		ActualThumbnailSize: actualThumbnailSize,
		UploadVerified:      true,
		ThumbnailVerified:   thumbnailVerified,
		StorageAdjustment:   storageAdjustment,
	}

	svc.logger.Info("✅ File upload completed successfully with quota adjustment",
		zap.Any("file_id", req.FileID),
		zap.Any("collection_id", file.CollectionID),
		zap.Any("owner_id", userID),
		zap.Int64("actual_file_size", actualFileSize),
		zap.Int64("actual_thumbnail_size", actualThumbnailSize),
		zap.Int64("storage_adjustment", storageAdjustment),
		zap.Bool("thumbnail_verified", thumbnailVerified))

	return response, nil
}
