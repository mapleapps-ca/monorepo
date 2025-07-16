// cloud/backend/internal/maplefile/service/file/softdelete.go
package file

import (
	"context"
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

type SoftDeleteFileRequestDTO struct {
	FileID gocql.UUID `json:"file_id"`
}

type SoftDeleteFileResponseDTO struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	ReleasedBytes int64  `json:"released_bytes"` // Amount of storage quota released
}

type SoftDeleteFileService interface {
	Execute(ctx context.Context, req *SoftDeleteFileRequestDTO) (*SoftDeleteFileResponseDTO, error)
}

type softDeleteFileServiceImpl struct {
	config                    *config.Configuration
	logger                    *zap.Logger
	collectionRepo            dom_collection.CollectionRepository
	getMetadataUseCase        uc_filemetadata.GetFileMetadataUseCase
	updateFileMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase
	softDeleteMetadataUseCase uc_filemetadata.SoftDeleteFileMetadataUseCase
	deleteDataUseCase         uc_fileobjectstorage.DeleteEncryptedDataUseCase
	listFilesByOwnerIDService ListFilesByOwnerIDService
	// Storage quota management
	storageQuotaHelperUseCase uc_feduser.FederatedUserStorageQuotaHelperUseCase
}

func NewSoftDeleteFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	updateFileMetadataUseCase uc_filemetadata.UpdateFileMetadataUseCase,
	softDeleteMetadataUseCase uc_filemetadata.SoftDeleteFileMetadataUseCase,
	deleteDataUseCase uc_fileobjectstorage.DeleteEncryptedDataUseCase,
	listFilesByOwnerIDService ListFilesByOwnerIDService,
	storageQuotaHelperUseCase uc_feduser.FederatedUserStorageQuotaHelperUseCase,
) SoftDeleteFileService {
	logger = logger.Named("SoftDeleteFileService")
	return &softDeleteFileServiceImpl{
		config:                    config,
		logger:                    logger,
		collectionRepo:            collectionRepo,
		getMetadataUseCase:        getMetadataUseCase,
		updateFileMetadataUseCase: updateFileMetadataUseCase,
		softDeleteMetadataUseCase: softDeleteMetadataUseCase,
		deleteDataUseCase:         deleteDataUseCase,
		listFilesByOwnerIDService: listFilesByOwnerIDService,
		storageQuotaHelperUseCase: storageQuotaHelperUseCase,
	}
}

func (svc *softDeleteFileServiceImpl) Execute(ctx context.Context, req *SoftDeleteFileRequestDTO) (*SoftDeleteFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File ID is required")
	}

	if req.FileID.String() == "" {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
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

		svc.logger.Debug("Debugging started, will list all files that belong to the authenticated user")
		currentFiles, err := svc.listFilesByOwnerIDService.Execute(ctx, &ListFilesByOwnerIDRequestDTO{OwnerID: userID})
		if err != nil {
			svc.logger.Error("Failed to list files by owner ID",
				zap.Any("error", err),
				zap.Any("user_id", userID))
			return nil, err
		}
		for _, file := range currentFiles.Files {
			svc.logger.Debug("File",
				zap.Any("id", file.ID))
		}

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

	// Check valid transitions.
	if err := dom_file.IsValidStateTransition(file.State, dom_file.FileStateDeleted); err != nil {
		svc.logger.Warn("Invalid file state transition",
			zap.Any("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Calculate storage space to be released
	//
	totalFileSize := file.EncryptedFileSizeInBytes + file.EncryptedThumbnailSizeInBytes

	svc.logger.Info("📊 Preparing to delete file and release storage quota",
		zap.Any("file_id", req.FileID),
		zap.Int64("file_size", file.EncryptedFileSizeInBytes),
		zap.Int64("thumbnail_size", file.EncryptedThumbnailSizeInBytes),
		zap.Int64("total_size_to_release", totalFileSize))

	//
	// STEP 6: Delete encrypted file data from object storage
	//
	err = svc.deleteDataUseCase.Execute(file.EncryptedFileObjectKey)
	if err != nil {
		svc.logger.Error("Failed to hard delete encrypted file data",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		// Continue with metadata deletion even if object storage deletion fails
	}

	// Delete thumbnail if it exists
	if file.EncryptedThumbnailObjectKey != "" {
		err = svc.deleteDataUseCase.Execute(file.EncryptedThumbnailObjectKey)
		if err != nil {
			svc.logger.Warn("Failed to hard delete encrypted thumbnail data",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID),
				zap.String("thumbnail_storage_path", file.EncryptedThumbnailObjectKey))
		}
	}

	//
	// STEP 7: Update file metadata for soft deletion
	//
	file.Version++ // Mutation means we increment version.
	file.ModifiedAt = time.Now()
	file.ModifiedByUserID = userID
	file.TombstoneVersion = file.Version
	file.TombstoneExpiry = time.Now().Add(time.Hour * 24 * 30)
	if err := svc.updateFileMetadataUseCase.Execute(ctx, file); err != nil {
		svc.logger.Warn("Failed to update file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
	}

	//
	// STEP 8: Delete file metadata
	//
	err = svc.softDeleteMetadataUseCase.Execute(req.FileID)
	if err != nil {
		svc.logger.Error("Failed to soft-delete file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 9: Release storage quota (only if the file was in active state)
	//
	var releasedBytes int64 = 0
	if file.State == dom_file.FileStateActive && totalFileSize > 0 {
		err = svc.storageQuotaHelperUseCase.OnFileDeleted(ctx, userID, totalFileSize)
		if err != nil {
			svc.logger.Error("🔴 Failed to release storage quota after file deletion",
				zap.String("user_id", userID.String()),
				zap.Int64("file_size", totalFileSize),
				zap.Error(err))
			// Don't fail the entire operation, but log the issue
			svc.logger.Warn("⚠️ File was deleted but storage quota was not properly updated. Manual correction may be needed.")
		} else {
			releasedBytes = totalFileSize
			svc.logger.Info("✅ Storage quota released successfully",
				zap.String("user_id", userID.String()),
				zap.Int64("released_bytes", releasedBytes))
		}
	} else if file.State == dom_file.FileStatePending {
		// For pending files, we should release the reserved quota
		err = svc.storageQuotaHelperUseCase.ReleaseQuota(ctx, userID, totalFileSize)
		if err != nil {
			svc.logger.Error("🔴 Failed to release reserved storage quota for pending file",
				zap.String("user_id", userID.String()),
				zap.Int64("file_size", totalFileSize),
				zap.Error(err))
		} else {
			releasedBytes = totalFileSize
			svc.logger.Info("✅ Reserved storage quota released for pending file",
				zap.String("user_id", userID.String()),
				zap.Int64("released_bytes", releasedBytes))
		}
	}

	svc.logger.Info("File soft-deleted successfully with storage quota updated",
		zap.Any("file_id", req.FileID),
		zap.Any("collection_id", file.CollectionID),
		zap.Any("user_id", userID),
		zap.Int64("released_bytes", releasedBytes))

	return &SoftDeleteFileResponseDTO{
		Success:       true,
		Message:       "File soft-deleted successfully and storage quota updated",
		ReleasedBytes: releasedBytes,
	}, nil
}
