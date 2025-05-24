// internal/usecase/localfile/list_directory.go
package localfile

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// ListDirectoryUseCase defines the interface for listing directory contents
type ListDirectoryUseCase interface {
	Execute(ctx context.Context, dirPath string) ([]*FileInfo, error)
	ExecuteFilesOnly(ctx context.Context, dirPath string) ([]*FileInfo, error)
	ExecuteDirectoriesOnly(ctx context.Context, dirPath string) ([]*FileInfo, error)
}

// listDirectoryUseCase implements the ListDirectoryUseCase interface
type listDirectoryUseCase struct {
	logger             *zap.Logger
	getFileInfoUseCase GetFileInfoUseCase
}

// NewListDirectoryUseCase creates a new use case for listing directories
func NewListDirectoryUseCase(
	logger *zap.Logger,
	getFileInfoUseCase GetFileInfoUseCase,
) ListDirectoryUseCase {
	return &listDirectoryUseCase{
		logger:             logger,
		getFileInfoUseCase: getFileInfoUseCase,
	}
}

// Execute lists all items in a directory
func (uc *listDirectoryUseCase) Execute(
	ctx context.Context,
	dirPath string,
) ([]*FileInfo, error) {
	uc.logger.Debug("Listing directory", zap.String("dirPath", dirPath))

	if dirPath == "" {
		return nil, errors.NewAppError("directory path is required", nil)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		uc.logger.Error("Failed to read directory", zap.String("dirPath", dirPath), zap.Error(err))
		return nil, errors.NewAppError("failed to read directory", err)
	}

	var fileInfos []*FileInfo
	for _, entry := range entries {
		entryPath := dirPath + string(os.PathSeparator) + entry.Name()
		info, err := uc.getFileInfoUseCase.Execute(ctx, entryPath)
		if err != nil {
			uc.logger.Warn("Failed to get info for directory entry",
				zap.String("entryPath", entryPath),
				zap.Error(err))
			continue
		}
		fileInfos = append(fileInfos, info)
	}

	uc.logger.Debug("Successfully listed directory",
		zap.String("dirPath", dirPath),
		zap.Int("count", len(fileInfos)))
	return fileInfos, nil
}

// ExecuteFilesOnly lists only files in a directory
func (uc *listDirectoryUseCase) ExecuteFilesOnly(
	ctx context.Context,
	dirPath string,
) ([]*FileInfo, error) {
	allItems, err := uc.Execute(ctx, dirPath)
	if err != nil {
		return nil, err
	}

	var files []*FileInfo
	for _, item := range allItems {
		if !item.IsDirectory {
			files = append(files, item)
		}
	}

	return files, nil
}

// ExecuteDirectoriesOnly lists only directories in a directory
func (uc *listDirectoryUseCase) ExecuteDirectoriesOnly(
	ctx context.Context,
	dirPath string,
) ([]*FileInfo, error) {
	allItems, err := uc.Execute(ctx, dirPath)
	if err != nil {
		return nil, err
	}

	var directories []*FileInfo
	for _, item := range allItems {
		if item.IsDirectory {
			directories = append(directories, item)
		}
	}

	return directories, nil
}
