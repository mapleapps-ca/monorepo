// internal/usecase/localfile/create.go
package localfile

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
)

// CreateLocalFileInput defines the input for creating a local file
type CreateLocalFileInput struct {
	RemoteID          primitive.ObjectID
	CollectionID      primitive.ObjectID
	EncryptedMetadata string
	DecryptedName     string
	DecryptedMimeType string
	EncryptedFileKey  keys.EncryptedFileKey
	EncryptionVersion string
	EncryptedFileData []byte
	DecryptedFileData []byte
	CreateThumbnail   bool
	ThumbnailData     []byte
}

// CreateLocalFileUseCase defines the interface for creating a local file
type CreateLocalFileUseCase interface {
	Execute(ctx context.Context, input CreateLocalFileInput) (*localfile.LocalFile, error)
}

// createLocalFileUseCase implements the CreateLocalFileUseCase interface
type createLocalFileUseCase struct {
	logger     *zap.Logger
	repository localfile.LocalFileRepository
}

// NewCreateLocalFileUseCase creates a new use case for creating local files
func NewCreateLocalFileUseCase(
	logger *zap.Logger,
	repository localfile.LocalFileRepository,
) CreateLocalFileUseCase {
	return &createLocalFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new local file
func (uc *createLocalFileUseCase) Execute(
	ctx context.Context,
	input CreateLocalFileInput,
) (*localfile.LocalFile, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.EncryptedMetadata == "" {
		return nil, errors.NewAppError("encrypted metadata is required", nil)
	}

	if len(input.EncryptedFileData) == 0 && len(input.EncryptedFileData) == 0 {
		return nil, errors.NewAppError("file data is required", nil)
	}

	// Create a new local file
	file := &localfile.LocalFile{
		ID:                primitive.NewObjectID(),
		RemoteID:          input.RemoteID,
		CollectionID:      input.CollectionID,
		EncryptedMetadata: input.EncryptedMetadata,
		DecryptedName:     input.DecryptedName,
		DecryptedMimeType: input.DecryptedMimeType,
		DecryptedFileSize: int64(len(input.DecryptedFileData)),
		EncryptedFileSize: 0,
		EncryptedFileKey:  input.EncryptedFileKey,
		EncryptionVersion: input.EncryptionVersion,
		CreatedAt:         time.Now(),
		ModifiedAt:        time.Now(),
		IsModifiedLocally: true,
		SyncStatus:        localfile.SyncStatusLocalOnly,
	}

	// Save the file metadata
	if err := uc.repository.Create(ctx, file); err != nil {
		return nil, errors.NewAppError("failed to create file metadata", err)
	}

	if input.DecryptedFileData != nil {

	}

	// // Save the file data
	// if err := uc.repository.SaveFileData(ctx, file, input.FileData); err != nil {
	// 	// If saving file data fails, try to clean up the metadata
	// 	_ = uc.repository.Delete(ctx, file.ID)
	// 	return nil, errors.NewAppError("failed to save file data", err)
	// }

	// Save thumbnail if provided
	if input.CreateThumbnail && input.ThumbnailData != nil && len(input.ThumbnailData) > 0 {
		if err := uc.repository.SaveThumbnail(ctx, file, input.ThumbnailData); err != nil {
			uc.logger.Warn("Failed to save thumbnail",
				zap.String("fileID", file.ID.Hex()),
				zap.Error(err))
			// Continue even if thumbnail saving fails
		}
	}

	return file, nil
}
