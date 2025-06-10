// internal/service/filesyncer/update_local_from_cloud.go
package filesyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/httperror"
)

// UpdateLocalFileFromCloudFileService defines the interface for updating a local file from a cloud file
type UpdateLocalFileFromCloudFileService interface {
	Execute(ctx context.Context, cloudID gocql.UUID, password string) (*dom_file.File, error)
}

// updateLocalFileFromCloudFileService implements the UpdateLocalFileFromCloudFileService interface
type updateLocalFileFromCloudFileService struct {
	logger            *zap.Logger
	cloudRepository   filedto.FileDTORepository
	getFileUseCase    uc_file.GetFileUseCase
	updateFileUseCase uc_file.UpdateFileUseCase
	deleteFileUseCase uc_file.DeleteFileUseCase
}

// NewUpdateLocalFileFromCloudFileService creates a new use case for updating local files from cloud
func NewUpdateLocalFileFromCloudFileService(
	logger *zap.Logger,
	cloudRepository filedto.FileDTORepository,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	deleteFileUseCase uc_file.DeleteFileUseCase,
) UpdateLocalFileFromCloudFileService {
	logger = logger.Named("UpdateLocalFileFromCloudFileService")
	return &updateLocalFileFromCloudFileService{
		logger:            logger,
		cloudRepository:   cloudRepository,
		getFileUseCase:    getFileUseCase,
		updateFileUseCase: updateFileUseCase,
		deleteFileUseCase: deleteFileUseCase,
	}
}

// Execute updates a local file from cloud file data
func (s *updateLocalFileFromCloudFileService) Execute(ctx context.Context, cloudFileID gocql.UUID, password string) (*dom_file.File, error) {
	//
	// STEP 1: Validate the input
	//
	e := make(map[string]string)

	if cloudFileID.IsZero() {
		e["cloudFileID"] = "Cloud file ID is required"
	}
	if password == "" {
		e["password"] = "Password is required"
	}

	// If any errors were found, return bad request error
	if len(e) != 0 {
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get the existing local file
	//
	localFile, err := s.getFileUseCase.Execute(ctx, cloudFileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}
	if localFile == nil {
		err := errors.NewAppError("no local file found", nil)
		s.logger.Error("❌ Failed to fetch local file",
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 3: Check if user has local file content (not just metadata)
	//
	hasLocalContent := s.hasLocalFileContent(localFile)

	if !hasLocalContent {
		s.logger.Debug("⏭️ Skipping file update - user only has metadata (SyncStatusCloudOnly), no local file content",
			zap.String("file_id", cloudFileID.Hex()),
			zap.Any("sync_status", localFile.SyncStatus),
			zap.String("file_path", localFile.FilePath),
			zap.String("encrypted_file_path", localFile.EncryptedFilePath))

		// Still update the metadata if cloud version is newer, but don't download content
		return s.updateMetadataOnly(ctx, cloudFileID, localFile)
	}

	s.logger.Debug("✅ User has local file content, proceeding with full update",
		zap.String("file_id", cloudFileID.Hex()),
		zap.Any("sync_status", localFile.SyncStatus))

	//
	// STEP 4: Get the file from cloud for full update
	//
	cloudFileDTO, err := s.cloudRepository.DownloadByIDFromCloud(ctx, cloudFileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get file from the cloud", err)
	}
	if cloudFileDTO == nil {
		err := errors.NewAppError("cloud file not found", nil)
		s.logger.Error("❌ Failed to fetch file from cloud",
			zap.Error(err))
		return nil, err
	}

	//
	// STEP 5: Check if update is needed
	//
	if localFile.Version >= cloudFileDTO.Version {
		s.logger.Debug("✅ Local file is already same or newest version compared with the cloud file",
			zap.String("file_id", cloudFileID.Hex()),
			zap.Uint64("local_version", localFile.Version),
			zap.Uint64("cloud_version", cloudFileDTO.Version))
		return nil, nil
	}

	// Handle deletion case
	if cloudFileDTO.State == "deleted" {
		if err := s.deleteFileUseCase.Execute(ctx, localFile.ID); err != nil {
			s.logger.Error("❌ Failed to delete local file",
				zap.String("file_id", cloudFileID.Hex()),
				zap.Uint64("local_version", localFile.Version),
				zap.Uint64("cloud_version", cloudFileDTO.Version),
				zap.Error(err))
			return nil, err
		}
		s.logger.Debug("🗑️ Local file is marked as deleted",
			zap.String("file_id", cloudFileID.Hex()),
			zap.Uint64("local_version", localFile.Version),
			zap.Uint64("cloud_version", cloudFileDTO.Version))
		return nil, nil
	}

	//
	// STEP 6: Update the local file from cloud data
	//
	cloudFile := mapFileDTOToDomain(cloudFileDTO)

	// Preserve local file paths and sync status if they exist
	cloudFile.FilePath = localFile.FilePath
	cloudFile.EncryptedFilePath = localFile.EncryptedFilePath
	cloudFile.ThumbnailPath = localFile.ThumbnailPath
	cloudFile.EncryptedThumbnailPath = localFile.EncryptedThumbnailPath
	cloudFile.StorageMode = localFile.StorageMode

	// Update sync status based on local file state
	if localFile.SyncStatus == dom_file.SyncStatusLocalOnly {
		cloudFile.SyncStatus = dom_file.SyncStatusSynced
	} else {
		cloudFile.SyncStatus = localFile.SyncStatus
	}

	// Preserve decrypted metadata if available
	if localFile.Name != "[Encrypted]" && localFile.Name != "" {
		cloudFile.Name = localFile.Name
	}
	if localFile.MimeType != "application/octet-stream" && localFile.MimeType != "" {
		cloudFile.MimeType = localFile.MimeType
	}

	// Update the file
	updateInput := uc_file.UpdateFileInput{
		ID:                     cloudFile.ID,
		CollectionID:           &cloudFile.CollectionID,
		OwnerID:                &cloudFile.OwnerID,
		EncryptedMetadata:      &cloudFile.EncryptedMetadata,
		EncryptionVersion:      &cloudFile.EncryptionVersion,
		EncryptedHash:          &cloudFile.EncryptedHash,
		EncryptedFileSize:      &cloudFile.EncryptedFileSize,
		EncryptedThumbnailSize: &cloudFile.EncryptedThumbnailSize,
		ModifiedAt:             &cloudFile.ModifiedAt,
		ModifiedByUserID:       &cloudFile.ModifiedByUserID,
		Version:                &cloudFile.Version,
		State:                  &cloudFile.State,
		SyncStatus:             &cloudFile.SyncStatus,
	}

	updatedFile, err := s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ Failed to update local file from cloud",
			zap.String("id", cloudFileDTO.ID.Hex()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("✅ Local file is updated with full content",
		zap.String("id", cloudFileID.Hex()),
		zap.Uint64("old_version", localFile.Version),
		zap.Uint64("new_version", cloudFileDTO.Version))

	return updatedFile, nil
}

// hasLocalFileContent checks if the local file has actual content downloaded
// (not just metadata)
func (s *updateLocalFileFromCloudFileService) hasLocalFileContent(localFile *dom_file.File) bool {
	// If sync status is CloudOnly, user definitely doesn't have local content
	if localFile.SyncStatus == dom_file.SyncStatusCloudOnly {
		return false
	}

	// For other sync statuses, check if there are actual file paths
	// indicating downloaded content
	hasDecryptedContent := localFile.FilePath != "" && localFile.SyncStatus != dom_file.SyncStatusCloudOnly
	hasEncryptedContent := localFile.EncryptedFilePath != "" && localFile.SyncStatus != dom_file.SyncStatusCloudOnly

	// User has local content if they have either decrypted or encrypted files
	return hasDecryptedContent || hasEncryptedContent
}

// updateMetadataOnly updates only the metadata without downloading file content
func (s *updateLocalFileFromCloudFileService) updateMetadataOnly(ctx context.Context, cloudFileID gocql.UUID, localFile *dom_file.File) (*dom_file.File, error) {
	// Get the file metadata from cloud (lightweight operation)
	cloudFileDTO, err := s.cloudRepository.DownloadByIDFromCloud(ctx, cloudFileID)
	if err != nil {
		// If we can't get metadata, just return the local file as-is
		s.logger.Warn("⚠️ Failed to get cloud metadata for metadata-only update, keeping local file as-is",
			zap.String("file_id", cloudFileID.Hex()),
			zap.Error(err))
		return localFile, nil
	}
	if cloudFileDTO == nil {
		s.logger.Debug("ℹ️ Cloud file not found for metadata-only update",
			zap.String("file_id", cloudFileID.Hex()))
		return localFile, nil
	}

	// Check if metadata update is needed
	if localFile.Version >= cloudFileDTO.Version {
		s.logger.Debug("✅ Local file metadata is already up to date",
			zap.String("file_id", cloudFileID.Hex()),
			zap.Uint64("local_version", localFile.Version),
			zap.Uint64("cloud_version", cloudFileDTO.Version))
		return localFile, nil
	}

	// Handle deletion case
	if cloudFileDTO.State == "deleted" {
		if err := s.deleteFileUseCase.Execute(ctx, localFile.ID); err != nil {
			s.logger.Error("❌ Failed to delete local file during metadata-only update",
				zap.String("file_id", cloudFileID.Hex()),
				zap.Error(err))
			return nil, err
		}
		s.logger.Debug("🗑️ Local file metadata marked as deleted",
			zap.String("file_id", cloudFileID.Hex()))
		return nil, nil
	}

	// Update only metadata fields, preserve local content references
	updateInput := uc_file.UpdateFileInput{
		ID:                     localFile.ID,
		EncryptedMetadata:      &cloudFileDTO.EncryptedMetadata,
		EncryptionVersion:      &cloudFileDTO.EncryptionVersion,
		EncryptedHash:          &cloudFileDTO.EncryptedHash,
		EncryptedFileSize:      &cloudFileDTO.EncryptedFileSizeInBytes,
		EncryptedThumbnailSize: &cloudFileDTO.EncryptedThumbnailSizeInBytes,
		ModifiedAt:             &cloudFileDTO.ModifiedAt,
		ModifiedByUserID:       &cloudFileDTO.ModifiedByUserID,
		Version:                &cloudFileDTO.Version,
		State:                  &cloudFileDTO.State,
		// Preserve local sync status (should remain CloudOnly)
		SyncStatus: &localFile.SyncStatus,
	}

	updatedFile, err := s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ Failed to update local file metadata",
			zap.String("id", cloudFileDTO.ID.Hex()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("✅ Local file metadata updated (content not downloaded)",
		zap.String("id", cloudFileID.Hex()),
		zap.Uint64("old_version", localFile.Version),
		zap.Uint64("new_version", cloudFileDTO.Version),
		zap.Any("sync_status", localFile.SyncStatus))

	return updatedFile, nil
}
