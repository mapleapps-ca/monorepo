// internal/usecase/localfile/read_file.go
package localfile

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// ReadFileUseCase defines the interface for reading files from the local file system
type ReadFileUseCase interface {
	Execute(ctx context.Context, filePath string) ([]byte, error)
	ExecuteAsString(ctx context.Context, filePath string) (string, error)
}

// readFileUseCase implements the ReadFileUseCase interface
type readFileUseCase struct {
	logger *zap.Logger
}

// NewReadFileUseCase creates a new use case for reading files
func NewReadFileUseCase(
	logger *zap.Logger,
) ReadFileUseCase {
	logger = logger.Named("ReadFileUseCase")
	return &readFileUseCase{
		logger: logger,
	}
}

// Execute reads a file and returns its contents as bytes
func (uc *readFileUseCase) Execute(
	ctx context.Context,
	filePath string,
) ([]byte, error) {
	uc.logger.Debug("Reading file", zap.String("filePath", filePath))

	if filePath == "" {
		return nil, errors.NewAppError("file path is required", nil)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		uc.logger.Error("Failed to read file", zap.String("filePath", filePath), zap.Error(err))
		return nil, errors.NewAppError("failed to read file", err)
	}

	uc.logger.Debug("Successfully read file",
		zap.String("filePath", filePath),
		zap.Int("size", len(data)))
	return data, nil
}

// ExecuteAsString reads a file and returns its contents as a string
func (uc *readFileUseCase) ExecuteAsString(
	ctx context.Context,
	filePath string,
) (string, error) {
	data, err := uc.Execute(ctx, filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
