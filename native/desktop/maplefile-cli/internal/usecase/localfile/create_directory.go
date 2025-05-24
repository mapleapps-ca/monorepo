// internal/usecase/localfile/create_directory.go
package localfile

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// CreateDirectoryUseCase defines the interface for creating directories
type CreateDirectoryUseCase interface {
	Execute(ctx context.Context, dirPath string) error
	ExecuteAll(ctx context.Context, dirPath string) error
}

// createDirectoryUseCase implements the CreateDirectoryUseCase interface
type createDirectoryUseCase struct {
	logger *zap.Logger
}

// NewCreateDirectoryUseCase creates a new use case for creating directories
func NewCreateDirectoryUseCase(
	logger *zap.Logger,
) CreateDirectoryUseCase {
	return &createDirectoryUseCase{
		logger: logger,
	}
}

// Execute creates a single directory (parent must exist)
func (uc *createDirectoryUseCase) Execute(
	ctx context.Context,
	dirPath string,
) error {
	uc.logger.Debug("Creating directory", zap.String("dirPath", dirPath))

	if dirPath == "" {
		return errors.NewAppError("directory path is required", nil)
	}

	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		if os.IsExist(err) {
			uc.logger.Debug("Directory already exists", zap.String("dirPath", dirPath))
			return nil
		}
		uc.logger.Error("Failed to create directory", zap.String("dirPath", dirPath), zap.Error(err))
		return errors.NewAppError("failed to create directory", err)
	}

	uc.logger.Debug("Successfully created directory", zap.String("dirPath", dirPath))
	return nil
}

// ExecuteAll creates directory and all parent directories if they don't exist
func (uc *createDirectoryUseCase) ExecuteAll(
	ctx context.Context,
	dirPath string,
) error {
	uc.logger.Debug("Creating directory with parents", zap.String("dirPath", dirPath))

	if dirPath == "" {
		return errors.NewAppError("directory path is required", nil)
	}

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		uc.logger.Error("Failed to create directory with parents", zap.String("dirPath", dirPath), zap.Error(err))
		return errors.NewAppError("failed to create directory with parents", err)
	}

	uc.logger.Debug("Successfully created directory with parents", zap.String("dirPath", dirPath))
	return nil
}
