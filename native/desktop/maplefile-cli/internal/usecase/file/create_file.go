// internal/usecase/file/create_file.go
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

// createFileUseCase implements the CreateFileUseCase interface
type createFileUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewCreateFileUseCase creates a new use case for creating local files
func NewCreateFileUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) CreateFileUseCase {
	logger = logger.Named("CreateFileUseCase")
	return &createFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new local file
func (uc *createFileUseCase) Execute(ctx context.Context, data *file.File) error {
	// Validate inputs
	if data.ID.IsZero() {
		return errors.NewAppError("ID is required", nil)
	}
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
			uc.logger.Error(" encrypted-only storage mode is missing data",
				zap.String("EncryptedFilePath", data.EncryptedFilePath),
			)
			return errors.NewAppError("encrypted file path is required for encrypted-only storage mode", nil)
		}
	case file.StorageModeDecryptedOnly:
		if data.FilePath == "" {
			uc.logger.Error(" decrypted-only storage mode is missing data",
				zap.String("FilePath", data.FilePath),
			)
			return errors.NewAppError("file path is required for decrypted-only storage mode", nil)
		}
	case file.StorageModeHybrid:
		if data.EncryptedFilePath == "" || data.FilePath == "" {
			uc.logger.Error("hybrid storage mode is missing data",
				zap.String("EncryptedFilePath", data.EncryptedFilePath),
				zap.String("FilePath", data.FilePath),
			)
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
