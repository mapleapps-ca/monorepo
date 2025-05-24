// internal/usecase/localfile/move_file.go
package localfile

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// MoveFileUseCase defines the interface for moving files on the local file system
type MoveFileUseCase interface {
	Execute(ctx context.Context, srcPath, destPath string) error
}

// moveFileUseCase implements the MoveFileUseCase interface
type moveFileUseCase struct {
	logger        *zap.Logger
	copyUseCase   CopyFileUseCase
	deleteUseCase DeleteFileUseCase
}

// NewMoveFileUseCase creates a new use case for moving files
func NewMoveFileUseCase(
	logger *zap.Logger,
	copyUseCase CopyFileUseCase,
	deleteUseCase DeleteFileUseCase,
) MoveFileUseCase {
	return &moveFileUseCase{
		logger:        logger,
		copyUseCase:   copyUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// Execute moves a file from source to destination
func (uc *moveFileUseCase) Execute(
	ctx context.Context,
	srcPath, destPath string,
) error {
	uc.logger.Debug("Moving file",
		zap.String("srcPath", srcPath),
		zap.String("destPath", destPath))

	if srcPath == "" {
		return errors.NewAppError("source path is required", nil)
	}
	if destPath == "" {
		return errors.NewAppError("destination path is required", nil)
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		uc.logger.Error("Failed to create destination directories",
			zap.String("destDir", destDir),
			zap.Error(err))
		return errors.NewAppError("failed to create destination directories", err)
	}

	// Try to rename first (works if on same filesystem)
	err := os.Rename(srcPath, destPath)
	if err == nil {
		uc.logger.Debug("Successfully moved file using rename",
			zap.String("srcPath", srcPath),
			zap.String("destPath", destPath))
		return nil
	}

	uc.logger.Debug("Rename failed, using copy and delete", zap.Error(err))

	// If rename fails, copy then delete
	err = uc.copyUseCase.Execute(ctx, srcPath, destPath)
	if err != nil {
		return errors.NewAppError("failed to copy file during move", err)
	}

	// Delete source file after successful copy
	err = uc.deleteUseCase.Execute(ctx, srcPath)
	if err != nil {
		uc.logger.Error("Failed to delete source file after copy",
			zap.String("srcPath", srcPath),
			zap.Error(err))
		return errors.NewAppError("failed to delete source file after copy", err)
	}

	uc.logger.Debug("Successfully moved file using copy and delete",
		zap.String("srcPath", srcPath),
		zap.String("destPath", destPath))
	return nil
}
