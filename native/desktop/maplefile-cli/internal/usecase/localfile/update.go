// internal/usecase/localfile/update.go
package localfile

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
)

// UpdateLocalFileInput defines the input for updating a local file
type UpdateLocalFileInput struct {
	ID                    primitive.ObjectID
	EncryptedMetadata     *string
	DecryptedName         *string
	DecryptedMimeType     *string
	FileData              []byte
	ThumbnailData         []byte
	MarkAsModifiedLocally bool
}

// UpdateLocalFileUseCase defines the interface for updating a local file
type UpdateLocalFileUseCase interface {
	Execute(ctx context.Context, input UpdateLocalFileInput) (*localfile.LocalFile, error)
	UpdateSyncStatus(ctx context.Context, id primitive.ObjectID, remoteID primitive.ObjectID, status localfile.SyncStatus) (*localfile.LocalFile, error)
}

// updateLocalFileUseCase implements the UpdateLocalFileUseCase interface
type updateLocalFileUseCase struct {
	logger     *zap.Logger
	repository localfile.LocalFileRepository
	getUseCase GetLocalFileUseCase
}

// NewUpdateLocalFileUseCase creates a new use case for updating local files
func NewUpdateLocalFileUseCase(
	logger *zap.Logger,
	repository localfile.LocalFileRepository,
	getUseCase GetLocalFileUseCase,
) UpdateLocalFileUseCase {
	return &updateLocalFileUseCase{
		logger:     logger,
		repository: repository,
		getUseCase: getUseCase,
	}
}

// Execute updates a local file
func (uc *updateLocalFileUseCase) Execute(
	ctx context.Context,
	input UpdateLocalFileInput,
) (*localfile.LocalFile, error) {
	// Validate inputs
	if input.ID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Get the existing file
	file, err := uc.getUseCase.ByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.EncryptedMetadata != nil {
		file.EncryptedMetadata = *input.EncryptedMetadata
	}

	if input.DecryptedName != nil {
		file.DecryptedName = *input.DecryptedName
	}

	if input.DecryptedMimeType != nil {
		file.DecryptedMimeType = *input.DecryptedMimeType
	}

	// Update modification status
	if input.MarkAsModifiedLocally {
		file.IsModifiedLocally = true
		file.SyncStatus = localfile.SyncStatusModifiedLocally
	}

	// Update timestamp
	file.ModifiedAt = time.Now()

	// Save file data if provided
	if input.FileData != nil && len(input.FileData) > 0 {
		if err := uc.repository.SaveFileData(ctx, file, input.FileData); err != nil {
			return nil, errors.NewAppError("failed to save updated file data", err)
		}
		file.EncryptedSize = int64(len(input.FileData))
		file.IsModifiedLocally = true
		file.SyncStatus = localfile.SyncStatusModifiedLocally
	}

	// Save thumbnail data if provided
	if input.ThumbnailData != nil && len(input.ThumbnailData) > 0 {
		if err := uc.repository.SaveThumbnail(ctx, file, input.ThumbnailData); err != nil {
			// Log but continue if thumbnail save fails
			uc.logger.Warn("Failed to save updated thumbnail data",
				zap.String("fileID", file.ID.Hex()),
				zap.Error(err))
		}
	}

	// Save the updated file metadata
	if err := uc.repository.Save(ctx, file); err != nil {
		return nil, errors.NewAppError("failed to save updated file metadata", err)
	}

	return file, nil
}

// UpdateSyncStatus updates the sync status of a local file
func (uc *updateLocalFileUseCase) UpdateSyncStatus(
	ctx context.Context,
	id primitive.ObjectID,
	remoteID primitive.ObjectID,
	status localfile.SyncStatus,
) (*localfile.LocalFile, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Get the existing file
	file, err := uc.getUseCase.ByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update sync status and related fields
	file.SyncStatus = status
	file.ModifiedAt = time.Now()

	if !remoteID.IsZero() {
		file.RemoteID = remoteID
	}

	// If synced, update last synced time and modified locally status
	if status == localfile.SyncStatusSynced {
		file.LastSyncedAt = time.Now()
		file.IsModifiedLocally = false
	}

	// Save the updated file metadata
	if err := uc.repository.Save(ctx, file); err != nil {
		return nil, errors.NewAppError("failed to update file sync status", err)
	}

	return file, nil
}
