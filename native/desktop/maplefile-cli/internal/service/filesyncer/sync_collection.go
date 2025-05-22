// internal/service/filesyncer/sync_collection.go
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

// SyncCollectionInput defines the input for syncing a collection
type SyncCollectionInput struct {
	CollectionID primitive.ObjectID
}

// SyncCollectionResult contains information about a collection sync operation
type SyncCollectionResult struct {
	TotalFiles      int
	SuccessfulSyncs int
	FailedSyncs     int
	UploadedFiles   int
	DownloadedFiles int
	Details         []SyncFileResult
}

// SyncCollectionService defines the interface for syncing collections
type SyncCollectionService interface {
	Execute(ctx context.Context, input SyncCollectionInput) (*SyncCollectionResult, error)
}

// syncCollectionService implements the SyncCollectionService interface
type syncCollectionService struct {
	logger                 *zap.Logger
	localFileListUseCase   localfileUseCase.ListLocalFilesUseCase
	remoteFileListUseCase  remotefileUseCase.ListRemoteFilesUseCase
	uploadToRemoteService  UploadToRemoteService
	downloadToLocalService DownloadToLocalService
}

// NewSyncCollectionService creates a new service for syncing collections
func NewSyncCollectionService(
	logger *zap.Logger,
	localFileListUseCase localfileUseCase.ListLocalFilesUseCase,
	remoteFileListUseCase remotefileUseCase.ListRemoteFilesUseCase,
	uploadToRemoteService UploadToRemoteService,
	downloadToLocalService DownloadToLocalService,
) SyncCollectionService {
	return &syncCollectionService{
		logger:                 logger,
		localFileListUseCase:   localFileListUseCase,
		remoteFileListUseCase:  remoteFileListUseCase,
		uploadToRemoteService:  uploadToRemoteService,
		downloadToLocalService: downloadToLocalService,
	}
}

// Execute synchronizes all files in a collection
func (s *syncCollectionService) Execute(
	ctx context.Context,
	input SyncCollectionInput,
) (*SyncCollectionResult, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get all local files in the collection
	localFiles, err := s.localFileListUseCase.ByCollection(ctx, input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to list local files in collection", err)
	}

	// Get all remote files in the collection
	remoteFiles, err := s.remoteFileListUseCase.ByCollection(ctx, input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to list remote files in collection", err)
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

	// Initialize result
	result := &SyncCollectionResult{
		TotalFiles: len(localFileMap) + len(remoteFileMap),
		Details:    make([]SyncFileResult, 0),
	}

	// Adjust total files count to avoid double counting
	for encryptedFileID := range localFileMap {
		if _, exists := remoteFileMap[encryptedFileID]; exists {
			result.TotalFiles--
		}
	}

	// First, process files that exist in both places or only locally
	for encryptedFileID, localFile := range localFileMap {
		remoteFile, exists := remoteFileMap[encryptedFileID]

		if !exists {
			// Local only, upload to remote
			uploadResult, err := s.uploadToRemoteService.Execute(ctx, UploadToRemoteInput{
				LocalID: localFile.ID,
			})
			if err != nil {
				s.logger.Error("Failed to upload local file to remote",
					zap.String("localID", localFile.ID.Hex()),
					zap.String("encryptedFileID", encryptedFileID),
					zap.Error(err))
				result.FailedSyncs++
				continue
			}

			result.SuccessfulSyncs++
			result.UploadedFiles++
			result.Details = append(result.Details, SyncFileResult{
				LocalFile:          uploadResult.LocalFile,
				RemoteFile:         uploadResult.RemoteFile,
				UploadedToRemote:   uploadResult.UploadedToRemote,
				SynchronizationLog: uploadResult.SynchronizationLog,
				SyncDirection:      "upload",
			})

		} else if localFile.IsModifiedLocally || localFile.SyncStatus != localfile.SyncStatusSynced {
			// Both exist but local is modified, upload to remote
			uploadResult, err := s.uploadToRemoteService.Execute(ctx, UploadToRemoteInput{
				LocalID: localFile.ID,
			})
			if err != nil {
				s.logger.Error("Failed to upload modified local file to remote",
					zap.String("localID", localFile.ID.Hex()),
					zap.String("remoteID", remoteFile.ID.Hex()),
					zap.String("encryptedFileID", encryptedFileID),
					zap.Error(err))
				result.FailedSyncs++
				continue
			}

			result.SuccessfulSyncs++
			result.UploadedFiles++
			result.Details = append(result.Details, SyncFileResult{
				LocalFile:          uploadResult.LocalFile,
				RemoteFile:         uploadResult.RemoteFile,
				UploadedToRemote:   uploadResult.UploadedToRemote,
				SynchronizationLog: uploadResult.SynchronizationLog + " (local was modified)",
				SyncDirection:      "upload",
			})

		} else if remoteFile.ModifiedAt.After(localFile.LastSyncedAt) {
			// Remote is newer, download to local
			downloadResult, err := s.downloadToLocalService.Execute(ctx, DownloadToLocalInput{
				RemoteID: remoteFile.ID,
			})
			if err != nil {
				s.logger.Error("Failed to download newer remote file to local",
					zap.String("localID", localFile.ID.Hex()),
					zap.String("remoteID", remoteFile.ID.Hex()),
					zap.String("encryptedFileID", encryptedFileID),
					zap.Error(err))
				result.FailedSyncs++
				continue
			}

			result.SuccessfulSyncs++
			result.DownloadedFiles++
			result.Details = append(result.Details, SyncFileResult{
				LocalFile:          downloadResult.LocalFile,
				RemoteFile:         downloadResult.RemoteFile,
				DownloadedToLocal:  downloadResult.DownloadedToLocal,
				SynchronizationLog: downloadResult.SynchronizationLog + " (remote was newer)",
				SyncDirection:      "download",
			})

		} else {
			// Files are in sync
			result.Details = append(result.Details, SyncFileResult{
				LocalFile:          localFile,
				RemoteFile:         remoteFile,
				SynchronizationLog: "Files are already in sync",
				SyncDirection:      "none",
			})
		}

		// Remove from remoteFileMap to track processed files
		delete(remoteFileMap, encryptedFileID)
	}

	// Process remaining files in remoteFileMap (remote only)
	for encryptedFileID, remoteFile := range remoteFileMap {
		// Remote only, download to local
		downloadResult, err := s.downloadToLocalService.Execute(ctx, DownloadToLocalInput{
			RemoteID: remoteFile.ID,
		})
		if err != nil {
			s.logger.Error("Failed to download remote-only file to local",
				zap.String("remoteID", remoteFile.ID.Hex()),
				zap.String("encryptedFileID", encryptedFileID),
				zap.Error(err))
			result.FailedSyncs++
			continue
		}

		result.SuccessfulSyncs++
		result.DownloadedFiles++
		result.Details = append(result.Details, SyncFileResult{
			LocalFile:          downloadResult.LocalFile,
			RemoteFile:         downloadResult.RemoteFile,
			DownloadedToLocal:  downloadResult.DownloadedToLocal,
			SynchronizationLog: downloadResult.SynchronizationLog,
			SyncDirection:      "download",
		})
	}

	return result, nil
}
