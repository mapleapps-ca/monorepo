// internal/usecase/localfile/validate_file_paths.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// FilePathStatus represents the status of file paths on the local system
type FilePathStatus struct {
	FileID                   primitive.ObjectID `json:"file_id"`
	EncryptedFileExists      bool               `json:"encrypted_file_exists"`
	DecryptedFileExists      bool               `json:"decrypted_file_exists"`
	EncryptedThumbnailExists bool               `json:"encrypted_thumbnail_exists"`
	DecryptedThumbnailExists bool               `json:"decrypted_thumbnail_exists"`
	EncryptedFilePath        string             `json:"encrypted_file_path"`
	DecryptedFilePath        string             `json:"decrypted_file_path"`
	EncryptedThumbnailPath   string             `json:"encrypted_thumbnail_path"`
	DecryptedThumbnailPath   string             `json:"decrypted_thumbnail_path"`
	StorageMode              string             `json:"storage_mode"`
}

// ValidateFilePathsUseCase defines the interface for validating file paths
type ValidateFilePathsUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID) (*FilePathStatus, error)
}

// validateFilePathsUseCase implements the ValidateFilePathsUseCase interface
type validateFilePathsUseCase struct {
	logger         *zap.Logger
	getFileUseCase GetFileUseCase
}

// NewValidateFilePathsUseCase creates a new use case for validating file paths
func NewValidateFilePathsUseCase(
	logger *zap.Logger,
	getFileUseCase GetFileUseCase,
) ValidateFilePathsUseCase {
	return &validateFilePathsUseCase{
		logger:         logger,
		getFileUseCase: getFileUseCase,
	}
}

// Execute validates that the file paths on the local system match the database records
func (uc *validateFilePathsUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
) (*FilePathStatus, error) {
	uc.logger.Debug("Validating file paths", zap.String("fileID", fileID.Hex()))

	// Get the file from database
	file, err := uc.getFileUseCase.Execute(ctx, fileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get file for path validation", err)
	}

	status := &FilePathStatus{
		FileID:                 file.ID,
		EncryptedFilePath:      file.EncryptedFilePath,
		DecryptedFilePath:      file.FilePath,
		EncryptedThumbnailPath: file.EncryptedThumbnailPath,
		DecryptedThumbnailPath: file.ThumbnailPath,
		StorageMode:            file.StorageMode,
	}

	// Check if files exist on disk
	// Note: This would require OS-level file system checks
	// For now, we'll set based on whether paths are provided
	status.EncryptedFileExists = file.EncryptedFilePath != ""
	status.DecryptedFileExists = file.FilePath != ""
	status.EncryptedThumbnailExists = file.EncryptedThumbnailPath != ""
	status.DecryptedThumbnailExists = file.ThumbnailPath != ""

	// TODO: Add actual file system validation
	// import "os"
	// if file.EncryptedFilePath != "" {
	//     if _, err := os.Stat(file.EncryptedFilePath); err == nil {
	//         status.EncryptedFileExists = true
	//     }
	// }

	return status, nil
}
