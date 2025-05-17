// internal/usecase/filesyncer/sync.go
package filesyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	localfileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	remotefileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// SyncResult contains information about a sync operation
type SyncResult struct {
	LocalFile          *localfile.LocalFile
	RemoteFile         *remotefile.RemoteFile
	UploadedToRemote   bool
	DownloadedToLocal  bool
	SynchronizationLog string
}

// FileSyncerUseCase defines the interface for synchronizing files
type FileSyncerUseCase interface {
	UploadToRemote(ctx context.Context, localID primitive.ObjectID) (*SyncResult, error)
	DownloadToLocal(ctx context.Context, remoteID primitive.ObjectID) (*SyncResult, error)
	SyncByEncryptedFileID(ctx context.Context, encryptedFileID string) (*SyncResult, error)
	SyncCollection(ctx context.Context, collectionID primitive.ObjectID) (int, error)
}

// fileSyncerUseCase implements the FileSyncerUseCase interface
type fileSyncerUseCase struct {
	logger                    *zap.Logger
	localFileGetUseCase       localfileUseCase.GetLocalFileUseCase
	localFileListUseCase      localfileUseCase.ListLocalFilesUseCase
	localFileCreateUseCase    localfileUseCase.CreateLocalFileUseCase
	localFileUpdateUseCase    localfileUseCase.UpdateLocalFileUseCase
	remoteFileCreateUseCase   remotefileUseCase.CreateRemoteFileUseCase
	remoteFileFetchUseCase    remotefileUseCase.FetchRemoteFileUseCase
	remoteFileListUseCase     remotefileUseCase.ListRemoteFilesUseCase
	remoteFileUploadUseCase   remotefileUseCase.UploadRemoteFileUseCase
	remoteFileDownloadUseCase remotefileUseCase.DownloadRemoteFileUseCase
}

// NewFileSyncerUseCase creates a new use case for syncing files
func NewFileSyncerUseCase(
	logger *zap.Logger,
	localFileGetUseCase localfileUseCase.GetLocalFileUseCase,
	localFileListUseCase localfileUseCase.ListLocalFilesUseCase,
	localFileCreateUseCase localfileUseCase.CreateLocalFileUseCase,
	localFileUpdateUseCase localfileUseCase.UpdateLocalFileUseCase,
	remoteFileCreateUseCase remotefileUseCase.CreateRemoteFileUseCase,
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase,
	remoteFileListUseCase remotefileUseCase.ListRemoteFilesUseCase,
	remoteFileUploadUseCase remotefileUseCase.UploadRemoteFileUseCase,
	remoteFileDownloadUseCase remotefileUseCase.DownloadRemoteFileUseCase,
) FileSyncerUseCase {
	return &fileSyncerUseCase{
		logger:                    logger,
		localFileGetUseCase:       localFileGetUseCase,
		localFileListUseCase:      localFileListUseCase,
		localFileCreateUseCase:    localFileCreateUseCase,
		localFileUpdateUseCase:    localFileUpdateUseCase,
		remoteFileCreateUseCase:   remoteFileCreateUseCase,
		remoteFileFetchUseCase:    remoteFileFetchUseCase,
		remoteFileListUseCase:     remoteFileListUseCase,
		remoteFileUploadUseCase:   remoteFileUploadUseCase,
		remoteFileDownloadUseCase: remoteFileDownloadUseCase,
	}
}

// UploadToRemote uploads a local file to the remote server
func (uc *fileSyncerUseCase) UploadToRemote(
	ctx context.Context,
	localID primitive.ObjectID,
) (*SyncResult, error) {
	// Get the local file
	localFile, err := uc.localFileGetUseCase.ByID(ctx, localID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}

	// Get the file data
	fileData, err := uc.localFileGetUseCase.GetFileData(ctx, localFile)
	if err != nil {
		return nil, errors.NewAppError("failed to get file data", err)
	}

	// Prepare result
	result := &SyncResult{
		LocalFile: localFile,
	}

	var remoteFile *remotefile.RemoteFile
	var remoteFileResponse *remotefile.RemoteFileResponse

	// Check if this file already exists on the remote
	if !localFile.RemoteID.IsZero() {
		// File already has a remote ID, update it
		remoteFile, err = uc.remoteFileFetchUseCase.ByID(ctx, localFile.RemoteID)
		if err != nil {
			// Remote file may have been deleted, create a new one
			uc.logger.Warn("Failed to fetch remote file, will create a new one",
				zap.String("localID", localID.Hex()),
				zap.String("remoteID", localFile.RemoteID.Hex()),
				zap.Error(err))
			remoteFile = nil
		}
	}

	if remoteFile == nil {
		// Create a new remote file
		input := remotefileUseCase.CreateRemoteFileInput{
			CollectionID:          localFile.CollectionID,
			EncryptedFileID:       localFile.EncryptedFileID,
			EncryptedSize:         localFile.EncryptedSize,
			EncryptedOriginalSize: "", // Not used in this implementation
			EncryptedMetadata:     localFile.EncryptedMetadata,
			EncryptedFileKey:      localFile.EncryptedFileKey,
			EncryptionVersion:     localFile.EncryptionVersion,
			EncryptedHash:         localFile.EncryptedHash,
			FileData:              fileData, // Upload file data with creation
		}

		remoteFileResponse, err = uc.remoteFileCreateUseCase.Execute(ctx, input)
		if err != nil {
			return nil, errors.NewAppError("failed to create remote file", err)
		}

		// Update local file with remote ID
		localFile, err = uc.localFileUpdateUseCase.UpdateSyncStatus(
			ctx,
			localFile.ID,
			remoteFileResponse.ID,
			localfile.SyncStatusSynced,
		)
		if err != nil {
			return nil, errors.NewAppError("failed to update local file after remote creation", err)
		}

		result.LocalFile = localFile
		result.UploadedToRemote = true
		result.SynchronizationLog = "Created new remote file and uploaded data"

		// Fetch the newly created remote file for the result
		remoteFile, err = uc.remoteFileFetchUseCase.ByID(ctx, remoteFileResponse.ID)
		if err != nil {
			uc.logger.Error("Failed to fetch newly created remote file",
				zap.String("remoteID", remoteFileResponse.ID.Hex()),
				zap.Error(err))
			// Continue without remote file in result
		} else {
			result.RemoteFile = remoteFile
		}
	} else {
		// Update existing remote file
		// For API-based remote files, we typically create a new file version
		// or use a new upload URL rather than updating in place
		err = uc.remoteFileUploadUseCase.Execute(ctx, remoteFile.ID, fileData)
		if err != nil {
			return nil, errors.NewAppError("failed to upload file data to existing remote file", err)
		}

		// Update local file sync status
		localFile, err = uc.localFileUpdateUseCase.UpdateSyncStatus(
			ctx,
			localFile.ID,
			localFile.RemoteID, // Use existing remote ID
			localfile.SyncStatusSynced,
		)
		if err != nil {
			return nil, errors.NewAppError("failed to update local file sync status", err)
		}

		result.LocalFile = localFile
		result.RemoteFile = remoteFile
		result.UploadedToRemote = true
		result.SynchronizationLog = "Updated existing remote file and uploaded data"
	}

	return result, nil
}

// DownloadToLocal downloads a remote file to local storage
func (uc *fileSyncerUseCase) DownloadToLocal(
	ctx context.Context,
	remoteID primitive.ObjectID,
) (*SyncResult, error) {
	// Get the remote file
	remoteFile, err := uc.remoteFileFetchUseCase.ByID(ctx, remoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to get remote file", err)
	}

	// Prepare result
	result := &SyncResult{
		RemoteFile: remoteFile,
	}

	// Check if this file already exists locally
	var localFile *localfile.LocalFile
	localFile, err = uc.localFileGetUseCase.ByRemoteID(ctx, remoteID)
	if err != nil {
		uc.logger.Error("Error checking for existing local file",
			zap.String("remoteID", remoteID.Hex()),
			zap.Error(err))
		// Continue to create a new local file
	}

	// Download the file data
	fileData, err := uc.remoteFileDownloadUseCase.Execute(ctx, remoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to download file data", err)
	}

	if localFile == nil {
		// Create a new local file
		input := localfileUseCase.CreateLocalFileInput{
			EncryptedFileID:   remoteFile.EncryptedFileID,
			CollectionID:      remoteFile.CollectionID,
			EncryptedMetadata: remoteFile.EncryptedMetadata,
			EncryptedFileKey:  remoteFile.EncryptedFileKey,
			EncryptionVersion: remoteFile.EncryptionVersion,
			FileData:          fileData,
			// No decrypted name or mime type - will be set during decryption
		}

		localFile, err = uc.localFileCreateUseCase.Execute(ctx, input)
		if err != nil {
			return nil, errors.NewAppError("failed to create local file", err)
		}

		// Update local file with remote ID
		localFile, err = uc.localFileUpdateUseCase.UpdateSyncStatus(
			ctx,
			localFile.ID,
			remoteID,
			localfile.SyncStatusSynced,
		)
		if err != nil {
			return nil, errors.NewAppError("failed to update local file after creation", err)
		}

		result.LocalFile = localFile
		result.DownloadedToLocal = true
		result.SynchronizationLog = "Created new local file and downloaded data"
	} else {
		// Update existing local file
		// Create an update input
		updateInput := localfileUseCase.UpdateLocalFileInput{
			ID:       localFile.ID,
			FileData: fileData,
		}

		if remoteFile.EncryptedMetadata != localFile.EncryptedMetadata {
			updateInput.EncryptedMetadata = &remoteFile.EncryptedMetadata
		}

		localFile, err = uc.localFileUpdateUseCase.Execute(ctx, updateInput)
		if err != nil {
			return nil, errors.NewAppError("failed to update local file", err)
		}

		// Update sync status
		localFile, err = uc.localFileUpdateUseCase.UpdateSyncStatus(
			ctx,
			localFile.ID,
			remoteID,
			localfile.SyncStatusSynced,
		)
		if err != nil {
			return nil, errors.NewAppError("failed to update local file sync status", err)
		}

		result.LocalFile = localFile
		result.DownloadedToLocal = true
		result.SynchronizationLog = "Updated existing local file and downloaded data"
	}

	return result, nil
}

// SyncByEncryptedFileID synchronizes a file by its encrypted file ID
func (uc *fileSyncerUseCase) SyncByEncryptedFileID(
	ctx context.Context,
	encryptedFileID string,
) (*SyncResult, error) {
	// Check for local file
	localFile, err := uc.localFileGetUseCase.ByEncryptedFileID(ctx, encryptedFileID)
	if err != nil {
		uc.logger.Error("Error checking for local file",
			zap.String("encryptedFileID", encryptedFileID),
			zap.Error(err))
		// Continue to check for remote file
	}

	// Check for remote file
	remoteFile, err := uc.remoteFileFetchUseCase.ByEncryptedFileID(ctx, encryptedFileID)
	if err != nil {
		uc.logger.Error("Error checking for remote file",
			zap.String("encryptedFileID", encryptedFileID),
			zap.Error(err))
		// Continue with sync logic
	}

	// Determine sync direction
	if localFile != nil && remoteFile == nil {
		// Local only, upload to remote
		return uc.UploadToRemote(ctx, localFile.ID)
	} else if localFile == nil && remoteFile != nil {
		// Remote only, download to local
		return uc.DownloadToLocal(ctx, remoteFile.ID)
	} else if localFile != nil && remoteFile != nil {
		// Both exist, determine which is newer
		if localFile.ModifiedAt.After(remoteFile.ModifiedAt) {
			// Local is newer, upload to remote
			return uc.UploadToRemote(ctx, localFile.ID)
		} else {
			// Remote is newer, download to local
			return uc.DownloadToLocal(ctx, remoteFile.ID)
		}
	} else {
		// Neither exists
		return nil, errors.NewAppError("file not found with the specified encrypted file ID", nil)
	}
}

// SyncCollection synchronizes all files in a collection
func (uc *fileSyncerUseCase) SyncCollection(
	ctx context.Context,
	collectionID primitive.ObjectID,
) (int, error) {
	// Get all local files in the collection
	localFiles, err := uc.localFileListUseCase.ByCollection(ctx, collectionID)
	if err != nil {
		return 0, errors.NewAppError("failed to list local files in collection", err)
	}

	// Get all remote files in the collection
	remoteFiles, err := uc.remoteFileListUseCase.ByCollection(ctx, collectionID)
	if err != nil {
		return 0, errors.NewAppError("failed to list remote files in collection", err)
	}

	// Create maps for quick lookup
	localFileMap := make(map[string]*localfile.LocalFile)
	remoteFileMap := make(map[string]*remotefile.RemoteFile)

	for _, file := range localFiles {
		localFileMap[file.EncryptedFileID] = file
	}

	for _, file := range remoteFiles {
		remoteFileMap[file.EncryptedFileID] = file
	}

	// Sync files
	successCount := 0

	// First, process files that exist in both places or only locally
	for encryptedFileID, localFile := range localFileMap {
		remoteFile, exists := remoteFileMap[encryptedFileID]

		if !exists {
			// Local only, upload to remote
			_, err := uc.UploadToRemote(ctx, localFile.ID)
			if err != nil {
				uc.logger.Error("Failed to upload local file to remote",
					zap.String("localID", localFile.ID.Hex()),
					zap.Error(err))
				continue
			}
			successCount++
		} else if localFile.IsModifiedLocally || localFile.SyncStatus != localfile.SyncStatusSynced {
			// Both exist but local is modified, upload to remote
			_, err := uc.UploadToRemote(ctx, localFile.ID)
			if err != nil {
				uc.logger.Error("Failed to upload modified local file to remote",
					zap.String("localID", localFile.ID.Hex()),
					zap.String("remoteID", remoteFile.ID.Hex()),
					zap.Error(err))
				continue
			}
			successCount++
		} else if remoteFile.ModifiedAt.After(localFile.LastSyncedAt) {
			// Remote is newer, download to local
			_, err := uc.DownloadToLocal(ctx, remoteFile.ID)
			if err != nil {
				uc.logger.Error("Failed to download newer remote file to local",
					zap.String("localID", localFile.ID.Hex()),
					zap.String("remoteID", remoteFile.ID.Hex()),
					zap.Error(err))
				continue
			}
			successCount++
		}

		// Remove from remoteFileMap to track processed files
		delete(remoteFileMap, encryptedFileID)
	}

	// Process remaining files in remoteFileMap (remote only)
	for _, remoteFile := range remoteFileMap {
		// Remote only, download to local
		_, err := uc.DownloadToLocal(ctx, remoteFile.ID)
		if err != nil {
			uc.logger.Error("Failed to download remote-only file to local",
				zap.String("remoteID", remoteFile.ID.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
