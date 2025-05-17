// internal/service/filesyncer/sync.go
package filesyncer

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/filesyncer"
)

// SyncOutput represents the result of a file synchronization operation
type SyncOutput struct {
	Success            bool                   `json:"success"`
	Message            string                 `json:"message"`
	LocalFile          *localfile.LocalFile   `json:"local_file,omitempty"`
	RemoteFile         *remotefile.RemoteFile `json:"remote_file,omitempty"`
	UploadedToRemote   bool                   `json:"uploaded_to_remote,omitempty"`
	DownloadedToLocal  bool                   `json:"downloaded_to_local,omitempty"`
	SynchronizationLog string                 `json:"synchronization_log,omitempty"`
	SyncCount          int                    `json:"sync_count,omitempty"`
}

// SyncService defines the interface for synchronizing files
type SyncService interface {
	UploadToRemote(ctx context.Context, localID string) (*SyncOutput, error)
	DownloadToLocal(ctx context.Context, remoteID string) (*SyncOutput, error)
	SyncByEncryptedFileID(ctx context.Context, encryptedFileID string) (*SyncOutput, error)
	SyncCollection(ctx context.Context, collectionID string) (*SyncOutput, error)
}

// syncService implements the SyncService interface
type syncService struct {
	logger            *zap.Logger
	fileSyncerUseCase filesyncer.FileSyncerUseCase
}

// NewSyncService creates a new service for synchronizing files
func NewSyncService(
	logger *zap.Logger,
	fileSyncerUseCase filesyncer.FileSyncerUseCase,
) SyncService {
	return &syncService{
		logger:            logger,
		fileSyncerUseCase: fileSyncerUseCase,
	}
}

// UploadToRemote uploads a local file to the remote server
func (s *syncService) UploadToRemote(ctx context.Context, localID string) (*SyncOutput, error) {
	// Validate inputs
	if localID == "" {
		return nil, errors.NewAppError("local file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(localID)
	if err != nil {
		return nil, errors.NewAppError("invalid local file ID format", err)
	}

	// Sync the file
	result, err := s.fileSyncerUseCase.UploadToRemote(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &SyncOutput{
		Success:            true,
		Message:            "File synced to remote successfully",
		LocalFile:          result.LocalFile,
		RemoteFile:         result.RemoteFile,
		UploadedToRemote:   result.UploadedToRemote,
		DownloadedToLocal:  result.DownloadedToLocal,
		SynchronizationLog: result.SynchronizationLog,
	}, nil
}

// DownloadToLocal downloads a remote file to local storage
func (s *syncService) DownloadToLocal(ctx context.Context, remoteID string) (*SyncOutput, error) {
	// Validate inputs
	if remoteID == "" {
		return nil, errors.NewAppError("remote file ID is required", nil)
	}

	// Convert ID to ObjectID
	fileID, err := primitive.ObjectIDFromHex(remoteID)
	if err != nil {
		return nil, errors.NewAppError("invalid remote file ID format", err)
	}

	// Sync the file
	result, err := s.fileSyncerUseCase.DownloadToLocal(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return &SyncOutput{
		Success:            true,
		Message:            "File synced to local successfully",
		LocalFile:          result.LocalFile,
		RemoteFile:         result.RemoteFile,
		UploadedToRemote:   result.UploadedToRemote,
		DownloadedToLocal:  result.DownloadedToLocal,
		SynchronizationLog: result.SynchronizationLog,
	}, nil
}

// SyncByEncryptedFileID synchronizes a file by its encrypted file ID
func (s *syncService) SyncByEncryptedFileID(ctx context.Context, encryptedFileID string) (*SyncOutput, error) {
	// Validate inputs
	if encryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	// Sync the file
	result, err := s.fileSyncerUseCase.SyncByEncryptedFileID(ctx, encryptedFileID)
	if err != nil {
		return nil, err
	}

	return &SyncOutput{
		Success:            true,
		Message:            "File synced successfully",
		LocalFile:          result.LocalFile,
		RemoteFile:         result.RemoteFile,
		UploadedToRemote:   result.UploadedToRemote,
		DownloadedToLocal:  result.DownloadedToLocal,
		SynchronizationLog: result.SynchronizationLog,
	}, nil
}

// SyncCollection synchronizes all files in a collection
func (s *syncService) SyncCollection(ctx context.Context, collectionID string) (*SyncOutput, error) {
	// Validate inputs
	if collectionID == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Convert ID to ObjectID
	colID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Sync the collection
	count, err := s.fileSyncerUseCase.SyncCollection(ctx, colID)
	if err != nil {
		return nil, err
	}

	return &SyncOutput{
		Success:   true,
		Message:   "Collection synced successfully",
		SyncCount: count,
	}, nil
}
