// internal/usecase/localfile/get_file_info.go
package localfile

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// FileInfo represents information about a file
type FileInfo struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ModifiedAt  time.Time `json:"modified_at"`
	IsDirectory bool      `json:"is_directory"`
	Permissions string    `json:"permissions"`
	Path        string    `json:"path"`
}

// GetFileInfoUseCase defines the interface for getting file information
type GetFileInfoUseCase interface {
	Execute(ctx context.Context, filePath string) (*FileInfo, error)
}

// getFileInfoUseCase implements the GetFileInfoUseCase interface
type getFileInfoUseCase struct {
	logger *zap.Logger
}

// NewGetFileInfoUseCase creates a new use case for getting file information
func NewGetFileInfoUseCase(
	logger *zap.Logger,
) GetFileInfoUseCase {
	return &getFileInfoUseCase{
		logger: logger,
	}
}

// Execute gets information about a file
func (uc *getFileInfoUseCase) Execute(
	ctx context.Context,
	filePath string,
) (*FileInfo, error) {
	uc.logger.Debug("Getting file info", zap.String("filePath", filePath))

	if filePath == "" {
		return nil, errors.NewAppError("file path is required", nil)
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewAppError("file does not exist", err)
		}
		uc.logger.Error("Failed to get file info", zap.String("filePath", filePath), zap.Error(err))
		return nil, errors.NewAppError("failed to get file info", err)
	}

	info := &FileInfo{
		Name:        stat.Name(),
		Size:        stat.Size(),
		ModifiedAt:  stat.ModTime(),
		IsDirectory: stat.IsDir(),
		Permissions: stat.Mode().String(),
		Path:        filePath,
	}

	uc.logger.Debug("Successfully got file info",
		zap.String("filePath", filePath),
		zap.String("name", info.Name),
		zap.Int64("size", info.Size))

	return info, nil
}
