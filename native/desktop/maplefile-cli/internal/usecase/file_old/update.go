// internal/usecase/file/update.go
package file

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// UpdateFileInput defines the input for updating a local file
type UpdateFileInput struct {
	ID                     primitive.ObjectID
	EncryptedMetadata      *string
	DecryptedName          *string
	DecryptedMimeType      *string
	EncryptedFilePath      *string
	FilePath               *string
	EncryptedThumbnailPath *string
	ThumbnailPath          *string
	StorageMode            *string
}

// UpdateFileUseCase defines the interface for updating a local file
type UpdateFileUseCase interface {
	Execute(ctx context.Context, input UpdateFileInput) (*dom_file.File, error)
}

// updateFileUseCase implements the UpdateFileUseCase interface
type updateFileUseCase struct {
	logger     *zap.Logger
	repository dom_file.FileRepository
	getUseCase GetFileUseCase
}

// NewUpdateFileUseCase creates a new use case for updating local files
func NewUpdateFileUseCase(
	logger *zap.Logger,
	repository dom_file.FileRepository,
	getUseCase GetFileUseCase,
) UpdateFileUseCase {
	return &updateFileUseCase{
		logger:     logger,
		repository: repository,
		getUseCase: getUseCase,
	}
}

// Execute updates a local file
func (uc *updateFileUseCase) Execute(
	ctx context.Context,
	input UpdateFileInput,
) (*dom_file.File, error) {
	// Validate inputs
	if input.ID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Get the existing file
	file, err := uc.getUseCase.Execute(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.EncryptedMetadata != nil {
		file.EncryptedMetadata = *input.EncryptedMetadata
	}

	if input.DecryptedName != nil {
		file.Name = *input.DecryptedName
	}

	if input.DecryptedMimeType != nil {
		file.MimeType = *input.DecryptedMimeType
	}

	if input.EncryptedFilePath != nil {
		file.EncryptedFilePath = *input.EncryptedFilePath
	}

	if input.FilePath != nil {
		file.FilePath = *input.FilePath
	}

	if input.EncryptedThumbnailPath != nil {
		file.EncryptedThumbnailPath = *input.EncryptedThumbnailPath
	}

	if input.ThumbnailPath != nil {
		file.ThumbnailPath = *input.ThumbnailPath
	}

	if input.StorageMode != nil {
		// Validate storage mode
		if *input.StorageMode != dom_file.StorageModeEncryptedOnly &&
			*input.StorageMode != dom_file.StorageModeDecryptedOnly &&
			*input.StorageMode != dom_file.StorageModeHybrid {
			return nil, errors.NewAppError(fmt.Sprintf("invalid storage mode: %s", *input.StorageMode), nil)
		}
		file.StorageMode = *input.StorageMode
	}

	// Update timestamps and modification status
	file.ModifiedAt = time.Now()
	file.IsModifiedLocally = true
	file.SyncStatus = dom_file.SyncStatusModifiedLocally

	// Save the updated file
	err = uc.repository.Update(ctx, file)
	if err != nil {
		return nil, errors.NewAppError("failed to update local file", err)
	}

	return file, nil
}
