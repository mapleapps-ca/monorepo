// internal/usecase/filesyncer/sync.go
package filesyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	localfileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	remotefileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// SyncFileInput defines the input for syncing a file
type SyncFileInput struct {
	EncryptedFileID string
}

// SyncFileResult contains information about a sync operation
type SyncFileResult struct {
	LocalFile          *localfile.LocalFile
	RemoteFile         *remotefile.RemoteFile
	UploadedToRemote   bool
	DownloadedToLocal  bool
	SynchronizationLog string
	SyncDirection      string // "upload", "download", "none"
}

// SyncFileUseCase defines the interface for syncing individual files
type SyncFileUseCase interface {
	Execute(ctx context.Context, input SyncFileInput) (*SyncFileResult, error)
}

// syncFileUseCase implements the SyncFileUseCase interface
type syncFileUseCase struct {
	logger                 *zap.Logger
	localFileGetUseCase    localfileUseCase.GetLocalFileUseCase
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase
	uploadToRemoteUseCase  UploadToRemoteUseCase
	downloadToLocalUseCase DownloadToLocalUseCase
}

// NewSyncFileUseCase creates a new use case for syncing files
func NewSyncFileUseCase(
	logger *zap.Logger,
	localFileGetUseCase localfileUseCase.GetLocalFileUseCase,
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase,
	uploadToRemoteUseCase UploadToRemoteUseCase,
	downloadToLocalUseCase DownloadToLocalUseCase,
) SyncFileUseCase {
	return &syncFileUseCase{
		logger:                 logger,
		localFileGetUseCase:    localFileGetUseCase,
		remoteFileFetchUseCase: remoteFileFetchUseCase,
		uploadToRemoteUseCase:  uploadToRemoteUseCase,
		downloadToLocalUseCase: downloadToLocalUseCase,
	}
}

// Execute synchronizes a file by its encrypted file ID
func (uc *syncFileUseCase) Execute(
	ctx context.Context,
	input SyncFileInput,
) (*SyncFileResult, error) {
	// Validate inputs
	if input.EncryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	// Check for local file
	localFile, err := uc.localFileGetUseCase.ByEncryptedFileID(ctx, input.EncryptedFileID)
	if err != nil {
		uc.logger.Error("Error checking for local file",
			zap.String("encryptedFileID", input.EncryptedFileID),
			zap.Error(err))
		// Continue to check for remote file
	}

	// Check for remote file
	remoteFile, err := uc.remoteFileFetchUseCase.ByEncryptedFileID(ctx, input.EncryptedFileID)
	if err != nil {
		uc.logger.Error("Error checking for remote file",
			zap.String("encryptedFileID", input.EncryptedFileID),
			zap.Error(err))
		// Continue with sync logic
	}

	// Determine sync direction and execute
	if localFile != nil && remoteFile == nil {
		// Local only, upload to remote
		uploadResult, err := uc.uploadToRemoteUseCase.Execute(ctx, UploadToRemoteInput{
			LocalID: localFile.ID,
		})
		if err != nil {
			return nil, err
		}

		return &SyncFileResult{
			LocalFile:          uploadResult.LocalFile,
			RemoteFile:         uploadResult.RemoteFile,
			UploadedToRemote:   uploadResult.UploadedToRemote,
			SynchronizationLog: uploadResult.SynchronizationLog,
			SyncDirection:      "upload",
		}, nil

	} else if localFile == nil && remoteFile != nil {
		// Remote only, download to local
		downloadResult, err := uc.downloadToLocalUseCase.Execute(ctx, DownloadToLocalInput{
			RemoteID: remoteFile.ID,
		})
		if err != nil {
			return nil, err
		}

		return &SyncFileResult{
			LocalFile:          downloadResult.LocalFile,
			RemoteFile:         downloadResult.RemoteFile,
			DownloadedToLocal:  downloadResult.DownloadedToLocal,
			SynchronizationLog: downloadResult.SynchronizationLog,
			SyncDirection:      "download",
		}, nil

	} else if localFile != nil && remoteFile != nil {
		// Both exist, determine which is newer
		if localFile.ModifiedAt.After(remoteFile.ModifiedAt) {
			// Local is newer, upload to remote
			uploadResult, err := uc.uploadToRemoteUseCase.Execute(ctx, UploadToRemoteInput{
				LocalID: localFile.ID,
			})
			if err != nil {
				return nil, err
			}

			return &SyncFileResult{
				LocalFile:          uploadResult.LocalFile,
				RemoteFile:         uploadResult.RemoteFile,
				UploadedToRemote:   uploadResult.UploadedToRemote,
				SynchronizationLog: uploadResult.SynchronizationLog + " (local was newer)",
				SyncDirection:      "upload",
			}, nil

		} else if remoteFile.ModifiedAt.After(localFile.ModifiedAt) {
			// Remote is newer, download to local
			downloadResult, err := uc.downloadToLocalUseCase.Execute(ctx, DownloadToLocalInput{
				RemoteID: remoteFile.ID,
			})
			if err != nil {
				return nil, err
			}

			return &SyncFileResult{
				LocalFile:          downloadResult.LocalFile,
				RemoteFile:         downloadResult.RemoteFile,
				DownloadedToLocal:  downloadResult.DownloadedToLocal,
				SynchronizationLog: downloadResult.SynchronizationLog + " (remote was newer)",
				SyncDirection:      "download",
			}, nil

		} else {
			// Files are in sync
			return &SyncFileResult{
				LocalFile:          localFile,
				RemoteFile:         remoteFile,
				SynchronizationLog: "Files are already in sync",
				SyncDirection:      "none",
			}, nil
		}

	} else {
		// Neither exists
		return nil, errors.NewAppError("file not found with the specified encrypted file ID", nil)
	}
}
