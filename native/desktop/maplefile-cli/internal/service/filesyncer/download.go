// internal/service/filesyncer/download.go
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

// DownloadToLocalService defines the interface for downloading remote files to local
type DownloadToLocalService interface {
	Execute(ctx context.Context, input DownloadToLocalInput) (*DownloadToLocalResult, error)
}

// downloadToLocalService implements the DownloadToLocalService interface
type downloadToLocalService struct {
	logger                    *zap.Logger
	localFileGetUseCase       localfileUseCase.GetLocalFileUseCase
	localFileCreateUseCase    localfileUseCase.CreateLocalFileUseCase
	localFileUpdateUseCase    localfileUseCase.UpdateLocalFileUseCase
	remoteFileFetchUseCase    remotefileUseCase.FetchRemoteFileUseCase
	remoteFileDownloadUseCase remotefileUseCase.DownloadRemoteFileUseCase
}

// NewDownloadToLocalService creates a new service for downloading to local
func NewDownloadToLocalService(
	logger *zap.Logger,
	localFileGetUseCase localfileUseCase.GetLocalFileUseCase,
	localFileCreateUseCase localfileUseCase.CreateLocalFileUseCase,
	localFileUpdateUseCase localfileUseCase.UpdateLocalFileUseCase,
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase,
	remoteFileDownloadUseCase remotefileUseCase.DownloadRemoteFileUseCase,
) DownloadToLocalService {
	return &downloadToLocalService{
		logger:                    logger,
		localFileGetUseCase:       localFileGetUseCase,
		localFileCreateUseCase:    localFileCreateUseCase,
		localFileUpdateUseCase:    localFileUpdateUseCase,
		remoteFileFetchUseCase:    remoteFileFetchUseCase,
		remoteFileDownloadUseCase: remoteFileDownloadUseCase,
	}
}

// Execute downloads a remote file to local storage
func (s *downloadToLocalService) Execute(
	ctx context.Context,
	input DownloadToLocalInput,
) (*DownloadToLocalResult, error) {
	// Validate inputs
	if input.RemoteID.IsZero() {
		return nil, errors.NewAppError("remote file ID is required", nil)
	}

	// Get the remote file
	remoteFile, err := s.remoteFileFetchUseCase.ByID(ctx, input.RemoteID)
	if err != nil {
		return nil, errors.NewAppError("failed to get remote file", err)
	}

	// Prepare result
	result := &DownloadToLocalResult{
		RemoteFile: remoteFile,
	}

	// Check if this file already exists locally
	var localFile *localfile.LocalFile
	localFile, err = s.localFileGetUseCase.ByRemoteID(ctx, input.RemoteID)
	if err != nil {
		s.logger.Error("Error checking for existing local file",
			zap.String("remoteID", input.RemoteID.Hex()),
			zap.Error(err))
		// Continue to create a new local file
	}

	// Download the file data
	fileData, err := s.remoteFileDownloadUseCase.Execute(ctx, input.RemoteID)
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

		localFile, err = s.localFileCreateUseCase.Execute(ctx, createInput)
		if err != nil {
			return nil, errors.NewAppError("failed to create local file", err)
		}

		// Update local file with remote ID
		localFile, err = s.localFileUpdateUseCase.UpdateSyncStatus(
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

		localFile, err = s.localFileUpdateUseCase.Execute(ctx, updateInput)
		if err != nil {
			return nil, errors.NewAppError("failed to update local file", err)
		}

		// Update sync status
		localFile, err = s.localFileUpdateUseCase.UpdateSyncStatus(
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
