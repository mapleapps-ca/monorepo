// internal/usecase/filesyncer/download.go
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

// DownloadToLocalInput defines the input for downloading to local
type DownloadToLocalInput struct {
	RemoteID primitive.ObjectID
}

// DownloadToLocalResult contains information about a download operation
type DownloadToLocalResult struct {
	LocalFile          *localfile.LocalFile
	RemoteFile         *remotefile.RemoteFile
	DownloadedToLocal  bool
	SynchronizationLog string
}

// DownloadToLocalUseCase defines the interface for downloading remote files to local
type DownloadToLocalUseCase interface {
	Execute(ctx context.Context, input DownloadToLocalInput) (*DownloadToLocalResult, error)
}

// downloadToLocalUseCase implements the DownloadToLocalUseCase interface
type downloadToLocalUseCase struct {
	logger                    *zap.Logger
	localFileGetUseCase       localfileUseCase.GetLocalFileUseCase
	localFileCreateUseCase    localfileUseCase.CreateLocalFileUseCase
	localFileUpdateUseCase    localfileUseCase.UpdateLocalFileUseCase
	remoteFileFetchUseCase    remotefileUseCase.FetchRemoteFileUseCase
	remoteFileDownloadUseCase remotefileUseCase.DownloadRemoteFileUseCase
}

// NewDownloadToLocalUseCase creates a new use case for downloading to local
func NewDownloadToLocalUseCase(
	logger *zap.Logger,
	localFileGetUseCase localfileUseCase.GetLocalFileUseCase,
	localFileCreateUseCase localfileUseCase.CreateLocalFileUseCase,
	localFileUpdateUseCase localfileUseCase.UpdateLocalFileUseCase,
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase,
	remoteFileDownloadUseCase remotefileUseCase.DownloadRemoteFileUseCase,
) DownloadToLocalUseCase {
	return &downloadToLocalUseCase{
		logger:                    logger,
		localFileGetUseCase:       localFileGetUseCase,
		localFileCreateUseCase:    localFileCreateUseCase,
		localFileUpdateUseCase:    localFileUpdateUseCase,
		remoteFileFetchUseCase:    remoteFileFetchUseCase,
		remoteFileDownloadUseCase: remoteFileDownloadUseCase,
	}
}

// Execute downloads a remote file to local storage
func (uc *downloadToLocalUseCase) Execute(
	ctx context.Context,
	input DownloadToLocalInput,
) (*DownloadToLocalResult, error) {
	// Validate inputs
	if input.RemoteID.IsZero() {
		return nil, errors.NewAppError("remote file ID is required", nil)
	}

	// Get the remote file
	remoteFile, err := uc.remoteFileFetchUseCase.ByID(ctx, input.RemoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to get remote file", err)
	}

	// Prepare result
	result := &DownloadToLocalResult{
		RemoteFile: remoteFile,
	}

	// Check if this file already exists locally
	var localFile *localfile.LocalFile
	localFile, err = uc.localFileGetUseCase.ByRemoteID(ctx, input.RemoteID)
	if err != nil {
		uc.logger.Error("Error checking for existing local file",
			zap.String("remoteID", input.RemoteID.Hex()),
			zap.Error(err))
		// Continue to create a new local file
	}

	// Download the file data
	fileData, err := uc.remoteFileDownloadUseCase.Execute(ctx, input.RemoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to download file data", err)
	}

	if localFile == nil {
		// Create a new local file
		createInput := localfileUseCase.CreateLocalFileInput{
			EncryptedFileID:   remoteFile.EncryptedFileID,
			CollectionID:      remoteFile.CollectionID,
			EncryptedMetadata: remoteFile.EncryptedMetadata,
			EncryptedFileKey:  remoteFile.EncryptedFileKey,
			EncryptionVersion: remoteFile.EncryptionVersion,
			FileData:          fileData,
			// No decrypted name or mime type - will be set during decryption
		}

		localFile, err = uc.localFileCreateUseCase.Execute(ctx, createInput)
		if err != nil {
			return nil, errors.NewAppError("failed to create local file", err)
		}

		// Update local file with remote ID
		localFile, err = uc.localFileUpdateUseCase.UpdateSyncStatus(
			ctx,
			localFile.ID,
			input.RemoteID,
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
			input.RemoteID,
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
