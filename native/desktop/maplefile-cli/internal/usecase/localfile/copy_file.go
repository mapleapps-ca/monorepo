// internal/usecase/localfile/copy_file.go
package localfile

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// CopyFileUseCase defines the interface for copying files on the local file system
type CopyFileUseCase interface {
	Execute(ctx context.Context, srcPath, destPath string) error
}

// copyFileUseCase implements the CopyFileUseCase interface
type copyFileUseCase struct {
	logger *zap.Logger
}

// NewCopyFileUseCase creates a new use case for copying files
func NewCopyFileUseCase(
	logger *zap.Logger,
) CopyFileUseCase {
	return &copyFileUseCase{
		logger: logger,
	}
}

// Execute copies a file from source to destination
func (uc *copyFileUseCase) Execute(
	ctx context.Context,
	srcPath, destPath string,
) error {
	uc.logger.Debug("Copying file",
		zap.String("srcPath", srcPath),
		zap.String("destPath", destPath))

	if srcPath == "" {
		return errors.NewAppError("source path is required", nil)
	}
	if destPath == "" {
		return errors.NewAppError("destination path is required", nil)
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		uc.logger.Error("Failed to open source file", zap.String("srcPath", srcPath), zap.Error(err))
		return errors.NewAppError("failed to open source file", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		uc.logger.Error("Failed to get source file info", zap.String("srcPath", srcPath), zap.Error(err))
		return errors.NewAppError("failed to get source file info", err)
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		uc.logger.Error("Failed to create destination directories",
			zap.String("destDir", destDir),
			zap.Error(err))
		return errors.NewAppError("failed to create destination directories", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		uc.logger.Error("Failed to create destination file", zap.String("destPath", destPath), zap.Error(err))
		return errors.NewAppError("failed to create destination file", err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		uc.logger.Error("Failed to copy file contents",
			zap.String("srcPath", srcPath),
			zap.String("destPath", destPath),
			zap.Error(err))
		return errors.NewAppError("failed to copy file contents", err)
	}

	uc.logger.Debug("Successfully copied file",
		zap.String("srcPath", srcPath),
		zap.String("destPath", destPath))
	return nil
}
