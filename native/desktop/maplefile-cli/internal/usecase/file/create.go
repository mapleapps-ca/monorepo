// internal/usecase/file/create.go
package file

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// CreateFileUseCase defines the interface for creating a local file
type CreateFileUseCase interface {
	Execute(ctx context.Context, data *file.File) error
}

// CreateFilesUseCase defines the interface for creating multiple local files
type CreateFilesUseCase interface {
	Execute(ctx context.Context, data []*file.File) error
}

// createFileUseCase implements the CreateFileUseCase interface
type createFileUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// createFilesUseCase implements the CreateFilesUseCase interface
type createFilesUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewCreateFileUseCase creates a new use case for creating local files
func NewCreateFileUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) CreateFileUseCase {
	return &createFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// NewCreateFilesUseCase creates a new use case for creating multiple local files
func NewCreateFilesUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) CreateFilesUseCase {
	return &createFilesUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new local file
func (uc *createFileUseCase) Execute(ctx context.Context, data *file.File) error {
	// Validate inputs
	if data.CollectionID.IsZero() {
		return errors.NewAppError("collection ID is required", nil)
	}

	if data.OwnerID.IsZero() {
		return errors.NewAppError("owner ID is required", nil)
	}

	if data.EncryptedMetadata == "" {
		return errors.NewAppError("encrypted metadata is required", nil)
	}

	if len(data.EncryptedFileKey.Ciphertext) == 0 || len(data.EncryptedFileKey.Nonce) == 0 {
		return errors.NewAppError("encrypted file key is required", nil)
	}

	if data.Name == "" {
		return errors.NewAppError("file name is required", nil)
	}

	if data.MimeType == "" {
		return errors.NewAppError("mime type is required", nil)
	}

	// Validate storage mode
	if data.StorageMode != file.StorageModeEncryptedOnly &&
		data.StorageMode != file.StorageModeDecryptedOnly &&
		data.StorageMode != file.StorageModeHybrid {
		return errors.NewAppError(fmt.Sprintf("invalid storage mode: %s (must be '%s', '%s', or '%s')",
			data.StorageMode, file.StorageModeEncryptedOnly, file.StorageModeDecryptedOnly, file.StorageModeHybrid), nil)
	}

	// Validate file paths based on storage mode
	switch data.StorageMode {
	case file.StorageModeEncryptedOnly:
		if data.EncryptedFilePath == "" {
			return errors.NewAppError("encrypted file path is required for encrypted-only storage mode", nil)
		}
	case file.StorageModeDecryptedOnly:
		if data.FilePath == "" {
			return errors.NewAppError("file path is required for decrypted-only storage mode", nil)
		}
	case file.StorageModeHybrid:
		if data.EncryptedFilePath == "" || data.FilePath == "" {
			return errors.NewAppError("both encrypted and decrypted file paths are required for hybrid storage mode", nil)
		}
	}

	// Save the file
	err := uc.repository.Create(ctx, data)
	if err != nil {
		return errors.NewAppError("failed to create local file", err)
	}

	return nil
}

// Execute creates multiple new local files
func (uc *createFilesUseCase) Execute(ctx context.Context, data []*file.File) error {
	if len(data) == 0 {
		return errors.NewAppError("at least one file is required", nil)
	}

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
		if fileData.StorageMode != file.StorageModeEncryptedOnly &&
			fileData.StorageMode != file.StorageModeDecryptedOnly &&
			fileData.StorageMode != file.StorageModeHybrid {
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
