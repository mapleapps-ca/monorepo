// internal/usecase/localfile/get_file_info.go
package localfile

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

// ComputeFileHashUseCase defines the interface for getting file information
type ComputeFileHashUseCase interface {
	Execute(ctx context.Context, filePath string) (string, error)
}

// computeFileHashUseCase implements the ComputeFileHashUseCase interface
type computeFileHashUseCase struct {
	logger *zap.Logger
}

// NewComputeFileHashUseCase creates a new use case for getting file information
func NewComputeFileHashUseCase(
	logger *zap.Logger,
) ComputeFileHashUseCase {
	return &computeFileHashUseCase{
		logger: logger,
	}
}

// Execute gets information about a file
func (uc *computeFileHashUseCase) Execute(ctx context.Context, filePath string) (string, error) {
	uc.logger.Debug("Getting file info", zap.String("filePath", filePath))

	if filePath == "" {
		return "", errors.NewAppError("file path is required", nil)
	}

	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	// Create a new SHA256 hasher
	h := sha256.New()

	// Read the file and write its contents into the hasher
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Get the byte representation of the hash
	hashBytes := h.Sum(nil)

	// Convert bytes to hex string
	hashHex := fmt.Sprintf("%x", hashBytes)

	return hashHex, nil
}
