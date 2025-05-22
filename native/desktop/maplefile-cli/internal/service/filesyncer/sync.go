// native/desktop/maplefile-cli/internal/service/filesyncer/sync.go
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

// SyncFileInput defines the input for syncing a file
type SyncFileInput struct {
	FileID primitive.ObjectID // Use actual file ID instead of encrypted file ID
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

// SyncFileService defines the interface for syncing individual files
type SyncFileService interface {
	Execute(ctx context.Context, input SyncFileInput) (*SyncFileResult, error)
}

// syncFileService implements the SyncFileService interface
type syncFileService struct {
	logger                 *zap.Logger
	localFileGetUseCase    localfileUseCase.GetLocalFileUseCase
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase
	uploadToRemoteService  UploadToRemoteService
	downloadToLocalService DownloadToLocalService
}

// NewSyncFileService creates a new service for syncing files
func NewSyncFileService(
	logger *zap.Logger,
	localFileGetUseCase localfileUseCase.GetLocalFileUseCase,
	remoteFileFetchUseCase remotefileUseCase.FetchRemoteFileUseCase,
	uploadToRemoteService UploadToRemoteService,
	downloadToLocalService DownloadToLocalService,
) SyncFileService {
	return &syncFileService{
		logger:                 logger,
		localFileGetUseCase:    localFileGetUseCase,
		remoteFileFetchUseCase: remoteFileFetchUseCase,
		uploadToRemoteService:  uploadToRemoteService,
		downloadToLocalService: downloadToLocalService,
	}
}

// Execute synchronizes a file by its ID
func (s *syncFileService) Execute(
	ctx context.Context,
	input SyncFileInput,
) (*SyncFileResult, error) {
	// Validate inputs
	if input.FileID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Try to find the file locally first
	localFile, err := s.localFileGetUseCase.ByID(ctx, input.FileID)
	if err != nil {
		s.logger.Debug("File not found locally, checking if it's a remote file",
			zap.String("fileID", input.FileID.Hex()),
			zap.Error(err))
	}

	// Check if we have a local file with a matching remote ID
	var remoteFile *remotefile.RemoteFile
	if localFile != nil && !localFile.RemoteID.IsZero() {
		// Get the corresponding remote file
		remoteFile, err = s.remoteFileFetchUseCase.ByID(ctx, localFile.RemoteID)
		if err != nil {
			s.logger.Error("Error checking for remote file",
				zap.String("remoteID", localFile.RemoteID.Hex()),
				zap.Error(err))
		}
	} else {
		// Try to find remote file by the provided ID
		remoteFile, err = s.remoteFileFetchUseCase.ByID(ctx, input.FileID)
		if err != nil {
			s.logger.Debug("File not found remotely either",
				zap.String("fileID", input.FileID.Hex()),
				zap.Error(err))
		}

		// If we found a remote file, check if we have a local file with matching remote ID
		if remoteFile != nil && localFile == nil {
			localFile, err = s.localFileGetUseCase.ByRemoteID(ctx, remoteFile.ID)
			if err != nil {
				s.logger.Debug("No local file found with matching remote ID",
					zap.String("remoteID", remoteFile.ID.Hex()),
					zap.Error(err))
			}
		}
	}

	// Determine sync direction and execute
	if localFile != nil && remoteFile == nil {
		// Local only, upload to remote
		uploadResult, err := s.uploadToRemoteService.Execute(ctx, UploadToRemoteInput{
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
		downloadResult, err := s.downloadToLocalService.Execute(ctx, DownloadToLocalInput{
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
			uploadResult, err := s.uploadToRemoteService.Execute(ctx, UploadToRemoteInput{
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
			downloadResult, err := s.downloadToLocalService.Execute(ctx, DownloadToLocalInput{
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
		return nil, errors.NewAppError("file not found with the specified ID", nil)
	}
}
