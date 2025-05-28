// internal/usecase/localfile/delete_file.go
package localfile

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// DeleteFileUseCase defines the interface for deleting files from the local file system
type DeleteFileUseCase interface {
	Execute(ctx context.Context, filePath string) error
}

// deleteFileUseCase implements the DeleteFileUseCase interface
type deleteFileUseCase struct {
	logger *zap.Logger
}

// NewDeleteFileUseCase creates a new use case for deleting files
func NewDeleteFileUseCase(
	logger *zap.Logger,
) DeleteFileUseCase {
	logger = logger.Named("DeleteFileUseCase")
	return &deleteFileUseCase{
		logger: logger,
	}
}

// Execute deletes a file from the file system
func (uc *deleteFileUseCase) Execute(
	ctx context.Context,
	filePath string,
) error {
	uc.logger.Debug("Deleting file", zap.String("filePath", filePath))

	if filePath == "" {
		return errors.NewAppError("file path is required", nil)
	}

	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			uc.logger.Warn("File does not exist", zap.String("filePath", filePath))
			return errors.NewAppError("file does not exist", err)
		}
		uc.logger.Error("Failed to delete file", zap.String("filePath", filePath), zap.Error(err))
		return errors.NewAppError("failed to delete file", err)
	}

	uc.logger.Debug("Successfully deleted file", zap.String("filePath", filePath))
	return nil
}
