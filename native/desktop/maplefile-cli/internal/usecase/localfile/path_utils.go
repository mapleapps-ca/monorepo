// internal/usecase/localfile/path_utils.go
package localfile

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// PathUtilsUseCase defines the interface for cross-platform path operations
type PathUtilsUseCase interface {
	Join(ctx context.Context, elements ...string) string
	Clean(ctx context.Context, path string) string
	IsAbsolute(ctx context.Context, path string) bool
	GetDirectory(ctx context.Context, path string) string
	GetFileName(ctx context.Context, path string) string
	GetFileExtension(ctx context.Context, path string) string
	ToSlash(ctx context.Context, path string) string
	FromSlash(ctx context.Context, path string) string
	GetHomeDirectory(ctx context.Context) (string, error)
	GetWorkingDirectory(ctx context.Context) (string, error)
	GetTempDirectory(ctx context.Context) string
}

// pathUtilsUseCase implements the PathUtilsUseCase interface
type pathUtilsUseCase struct {
	logger *zap.Logger
}

// NewPathUtilsUseCase creates a new use case for path utilities
func NewPathUtilsUseCase(
	logger *zap.Logger,
) PathUtilsUseCase {
	logger = logger.Named("PathUtilsUseCase")
	return &pathUtilsUseCase{
		logger: logger,
	}
}

// Join combines path elements using the OS-specific separator
func (uc *pathUtilsUseCase) Join(ctx context.Context, elements ...string) string {
	result := filepath.Join(elements...)
	uc.logger.Debug("Joined paths",
		zap.Strings("elements", elements),
		zap.String("result", result))
	return result
}

// Clean cleans up the path (removes redundant separators, resolves . and ..)
func (uc *pathUtilsUseCase) Clean(ctx context.Context, path string) string {
	result := filepath.Clean(path)
	uc.logger.Debug("Cleaned path",
		zap.String("original", path),
		zap.String("cleaned", result))
	return result
}

// IsAbsolute checks if the path is absolute
func (uc *pathUtilsUseCase) IsAbsolute(ctx context.Context, path string) bool {
	result := filepath.IsAbs(path)
	uc.logger.Debug("Checked if path is absolute",
		zap.String("path", path),
		zap.Bool("isAbsolute", result))
	return result
}

// GetDirectory returns the directory portion of the path
func (uc *pathUtilsUseCase) GetDirectory(ctx context.Context, path string) string {
	result := filepath.Dir(path)
	uc.logger.Debug("Got directory from path",
		zap.String("path", path),
		zap.String("directory", result))
	return result
}

// GetFileName returns the filename portion of the path
func (uc *pathUtilsUseCase) GetFileName(ctx context.Context, path string) string {
	result := filepath.Base(path)
	uc.logger.Debug("Got filename from path",
		zap.String("path", path),
		zap.String("filename", result))
	return result
}

// GetFileExtension returns the file extension
func (uc *pathUtilsUseCase) GetFileExtension(ctx context.Context, path string) string {
	result := filepath.Ext(path)
	uc.logger.Debug("Got file extension",
		zap.String("path", path),
		zap.String("extension", result))
	return result
}

// ToSlash converts path separators to forward slashes (useful for URLs/cross-platform)
func (uc *pathUtilsUseCase) ToSlash(ctx context.Context, path string) string {
	result := filepath.ToSlash(path)
	uc.logger.Debug("Converted to forward slashes",
		zap.String("original", path),
		zap.String("converted", result))
	return result
}

// FromSlash converts forward slashes to OS-specific separators
func (uc *pathUtilsUseCase) FromSlash(ctx context.Context, path string) string {
	result := filepath.FromSlash(path)
	uc.logger.Debug("Converted from forward slashes",
		zap.String("original", path),
		zap.String("converted", result))
	return result
}

// GetHomeDirectory gets the user's home directory
func (uc *pathUtilsUseCase) GetHomeDirectory(ctx context.Context) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		uc.logger.Error("Failed to get home directory", zap.Error(err))
		return "", errors.NewAppError("failed to get home directory", err)
	}

	uc.logger.Debug("Got home directory", zap.String("home", home))
	return home, nil
}

// GetWorkingDirectory gets the current working directory
func (uc *pathUtilsUseCase) GetWorkingDirectory(ctx context.Context) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		uc.logger.Error("Failed to get working directory", zap.Error(err))
		return "", errors.NewAppError("failed to get working directory", err)
	}

	uc.logger.Debug("Got working directory", zap.String("workingDir", wd))
	return wd, nil
}

// GetTempDirectory gets the system temporary directory
func (uc *pathUtilsUseCase) GetTempDirectory(ctx context.Context) string {
	temp := os.TempDir()
	uc.logger.Debug("Got temp directory", zap.String("tempDir", temp))
	return temp
}
