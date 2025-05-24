// internal/usecase/localfile/check_file_exists.go
package localfile

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// CheckFileExistsUseCase defines the interface for checking if files exist on the local file system
type CheckFileExistsUseCase interface {
	Execute(ctx context.Context, filePath string) (bool, error)
}

// checkFileExistsUseCase implements the CheckFileExistsUseCase interface
type checkFileExistsUseCase struct {
	logger *zap.Logger
}

// NewCheckFileExistsUseCase creates a new use case for checking file existence
func NewCheckFileExistsUseCase(
	logger *zap.Logger,
) CheckFileExistsUseCase {
	return &checkFileExistsUseCase{
		logger: logger,
	}
}

// Execute checks if a file exists on the file system
func (uc *checkFileExistsUseCase) Execute(
	ctx context.Context,
	filePath string,
) (bool, error) {
	uc.logger.Debug("Checking if file exists", zap.String("filePath", filePath))

	if filePath == "" {
		return false, errors.NewAppError("file path is required", nil)
	}

	_, err := os.Stat(filePath)
	if err == nil {
		uc.logger.Debug("File exists", zap.String("filePath", filePath))
		return true, nil
	}

	if os.IsNotExist(err) {
		uc.logger.Debug("File does not exist", zap.String("filePath", filePath))
		return false, nil
	}

	// Some other error occurred
	uc.logger.Error("Error checking file existence", zap.String("filePath", filePath), zap.Error(err))
	return false, errors.NewAppError("failed to check file existence", err)
}
