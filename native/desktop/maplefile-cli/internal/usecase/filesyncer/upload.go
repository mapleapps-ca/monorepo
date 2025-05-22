// internal/usecase/filesyncer/upload.go
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

// UploadToRemoteInput defines the input for uploading to remote
type UploadToRemoteInput struct {
	LocalID primitive.ObjectID
}

// UploadToRemoteResult contains information about an upload operation
type UploadToRemoteResult struct {
	LocalFile          *localfile.LocalFile
	RemoteFile         *remotefile.RemoteFile
	UploadedToRemote   bool
	SynchronizationLog string
}

// UploadToRemoteUseCase defines the interface for uploading local files to remote
type UploadToRemoteUseCase interface {
	Execute(ctx context.Context, input UploadToRemoteInput) (*UploadToRemoteResult, error)
}

// uploadToRemoteUseCase implements the UploadToRemoteUseCase interface
type uploadToRemoteUseCase struct {
	logger                  *zap.Logger
	localFileGetUseCase     localfileUseCase.GetLocalFileUseCase
	localFileUpdateUseCase  localfileUseCase.UpdateLocalFileUseCase
	remoteFileCreateUseCase remotefileUseCase.CreateRemoteFileUseCase
	remoteFileFetchUseCase  remotefileUseCase.FetchRemoteFileUseCase
	remoteFileUploadUseCase remotefileUseCase.UploadRemoteFileUseCase
}

// NewUploadToRemoteUseCase creates a new use case for uploading to remote
func NewUploadToRemoteUseCase(
	logger *zap.Logger,
	localFileGetUseCase localfileUseCase.GetLocalFileUseCase,
	localFileUpdateUseCase localfileUseCase.UpdateLocalFileUseCase,
	remoteFileCreateUseCase remotefileUseCase.CreateRemoteFileUseCase,
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase,
	remoteFileUploadUseCase remotefileUseCase.UploadRemoteFileUseCase,
) UploadToRemoteUseCase {
	return &uploadToRemoteUseCase{
		logger:                  logger,
		localFileGetUseCase:     localFileGetUseCase,
		localFileUpdateUseCase:  localFileUpdateUseCase,
		remoteFileCreateUseCase: remoteFileCreateUseCase,
		remoteFileFetchUseCase:  remoteFileFetchUseCase,
		remoteFileUploadUseCase: remoteFileUploadUseCase,
	}
}

// Execute uploads a local file to the remote server
func (uc *uploadToRemoteUseCase) Execute(
	ctx context.Context,
	input UploadToRemoteInput,
) (*UploadToRemoteResult, error) {
	// Validate inputs
	if input.LocalID.IsZero() {
		return nil, errors.NewAppError("local file ID is required", nil)
	}

	// Get the local file
	localFile, err := uc.localFileGetUseCase.ByID(ctx, input.LocalID)
	if err != nil {
		return nil, errors.NewAppError("failed to get local file", err)
	}

	// Get the file data
	fileData, err := uc.localFileGetUseCase.GetFileData(ctx, localFile)
	if err != nil {
		return nil, errors.NewAppError("failed to get file data", err)
	}

	// Prepare result
	result := &UploadToRemoteResult{
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
				zap.String("localID", input.LocalID.Hex()),
				zap.String("remoteID", localFile.RemoteID.Hex()),
				zap.Error(err))
			remoteFile = nil
		}
	}

	if remoteFile == nil {
		// Create a new remote file
		createInput := remotefileUseCase.CreateRemoteFileInput{
			CollectionID:      localFile.CollectionID,
			EncryptedFileID:   localFile.EncryptedFileID,
			EncryptedFileSize: localFile.EncryptedFileSize,
			EncryptedMetadata: localFile.EncryptedMetadata,
			EncryptedFileKey:  localFile.EncryptedFileKey,
			EncryptionVersion: localFile.EncryptionVersion,
			EncryptedHash:     localFile.EncryptedHash,
			FileData:          fileData, // Upload file data with creation
		}

		remoteFileResponse, err = uc.remoteFileCreateUseCase.Execute(ctx, createInput)
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
