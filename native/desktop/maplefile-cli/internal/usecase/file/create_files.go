// internal/usecase/file/create_files.go
package file

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// CreateFilesUseCase defines the interface for creating multiple local files
type CreateFilesUseCase interface {
	Execute(ctx context.Context, data []*dom_file.File) error
}

// createFilesUseCase implements the CreateFilesUseCase interface
type createFilesUseCase struct {
	logger     *zap.Logger
	repository dom_file.FileRepository
}

// NewCreateFilesUseCase creates a new use case for creating multiple local files
func NewCreateFilesUseCase(
	logger *zap.Logger,
	repository dom_file.FileRepository,
) CreateFilesUseCase {
	logger = logger.Named("CreateFilesUseCase")
	return &createFilesUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates multiple new local files
func (uc *createFilesUseCase) Execute(ctx context.Context, data []*dom_file.File) error {
	if len(data) == 0 {
		return errors.NewAppError("at least one file is required", nil)
	}

	// Validate each file using the same logic as single file creation
	for i, fileData := range data {
		// Validate without actually creating (we'll use CreateMany for efficiency)
		if fileData.CollectionID.IsZero() {
			return errors.NewAppError(fmt.Sprintf("collection ID is required for file at index %d", i), nil)
		}
		if fileData.OwnerID.IsZero() {
			return errors.NewAppError(fmt.Sprintf("owner ID is required for file at index %d", i), nil)
		}
		if fileData.EncryptedMetadata == "" {
			return errors.NewAppError(fmt.Sprintf("encrypted metadata is required for file at index %d", i), nil)
		}
		if len(fileData.EncryptedFileKey.Ciphertext) == 0 || len(fileData.EncryptedFileKey.Nonce) == 0 {
			return errors.NewAppError(fmt.Sprintf("encrypted file key is required for file at index %d", i), nil)
		}
		if fileData.Name == "" {
			return errors.NewAppError(fmt.Sprintf("file name is required for file at index %d", i), nil)
		}
		if fileData.MimeType == "" {
			return errors.NewAppError(fmt.Sprintf("mime type is required for file at index %d", i), nil)
		}

		// Validate storage mode
		if fileData.StorageMode != dom_file.StorageModeEncryptedOnly &&
			fileData.StorageMode != dom_file.StorageModeDecryptedOnly &&
			fileData.StorageMode != dom_file.StorageModeHybrid {
			return errors.NewAppError(fmt.Sprintf("invalid storage mode for file at index %d: %s", i, fileData.StorageMode), nil)
		}
	}

	// Save all files
	err := uc.repository.CreateMany(ctx, data)
	if err != nil {
		return errors.NewAppError("failed to create local files", err)
	}

	return nil
}
