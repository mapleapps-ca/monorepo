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
	ExecuteForBytes(ctx context.Context, filePath string) ([]byte, error)
	ExecuteForString(ctx context.Context, filePath string) (string, error)
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
func (uc *computeFileHashUseCase) ExecuteForBytes(ctx context.Context, filePath string) ([]byte, error) {
	uc.logger.Debug("Getting file info", zap.String("filePath", filePath))

	if filePath == "" {
		return nil, errors.NewAppError("file path is required", nil)
	}

	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer f.Close()

	// Create a new SHA256 hasher
	h := sha256.New()

	// Developer Note:
	// To efficiently calculate the hash, we read the file in chunks.
	// This allows us to process potentially large files without using too much memory at once.

	buf := make([]byte, 1024*1024) // Reading in 1MB chunks
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read file: %v", err)
		}
		if n == 0 {
			break
		}
		if _, err := h.Write(buf[:n]); err != nil {
			return nil, fmt.Errorf("failed to write to hasher: %v", err)
		}
	}

	// Get the byte representation of the hash
	hashBytes := h.Sum(nil)

	return hashBytes, nil
}

func (uc *computeFileHashUseCase) ExecuteForString(ctx context.Context, filePath string) (string, error) {
	hashBytes, err := uc.ExecuteForBytes(ctx, filePath)
	if err != nil {
		return "", err
	}

	// Convert bytes to hex string
	hashHex := fmt.Sprintf("%x", hashBytes)
	return hashHex, nil
}
