// internal/usecase/localfile/write_file.go
package localfile

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// WriteFileUseCase defines the interface for writing files to the local file system
type WriteFileUseCase interface {
	Execute(ctx context.Context, filePath string, data []byte) error
	ExecuteString(ctx context.Context, filePath string, content string) error
	ExecuteWithPermissions(ctx context.Context, filePath string, data []byte, perm os.FileMode) error
}

// writeFileUseCase implements the WriteFileUseCase interface
type writeFileUseCase struct {
	logger *zap.Logger
}

// NewWriteFileUseCase creates a new use case for writing files
func NewWriteFileUseCase(
	logger *zap.Logger,
) WriteFileUseCase {
	return &writeFileUseCase{
		logger: logger,
	}
}

// Execute writes data to a file with default permissions (0644)
func (uc *writeFileUseCase) Execute(
	ctx context.Context,
	filePath string,
	data []byte,
) error {
	return uc.ExecuteWithPermissions(ctx, filePath, data, 0644)
}

// ExecuteString writes string content to a file with default permissions
func (uc *writeFileUseCase) ExecuteString(
	ctx context.Context,
	filePath string,
	content string,
) error {
	return uc.Execute(ctx, filePath, []byte(content))
}

// ExecuteWithPermissions writes data to a file with specified permissions
func (uc *writeFileUseCase) ExecuteWithPermissions(
	ctx context.Context,
	filePath string,
	data []byte,
	perm os.FileMode,
) error {
	uc.logger.Debug("Writing file",
		zap.String("filePath", filePath),
		zap.Int("size", len(data)),
		zap.String("permissions", perm.String()))

	if filePath == "" {
		return errors.NewAppError("file path is required", nil)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		uc.logger.Error("Failed to create parent directories",
			zap.String("dir", dir),
			zap.Error(err))
		return errors.NewAppError("failed to create parent directories", err)
	}

	// Write the file
	err := os.WriteFile(filePath, data, perm)
	if err != nil {
		uc.logger.Error("Failed to write file", zap.String("filePath", filePath), zap.Error(err))
		return errors.NewAppError("failed to write file", err)
	}

	uc.logger.Debug("Successfully wrote file", zap.String("filePath", filePath))
	return nil
}
