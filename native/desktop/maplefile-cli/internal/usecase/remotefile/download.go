// internal/usecase/remotefile/download.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// DownloadRemoteFileUseCase defines the interface for downloading file data
type DownloadRemoteFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) ([]byte, error)
	GetDownloadURL(ctx context.Context, id primitive.ObjectID) (string, error)
}

// downloadRemoteFileUseCase implements the DownloadRemoteFileUseCase interface
type downloadRemoteFileUseCase struct {
	logger     *zap.Logger
	repository remotefile.RemoteFileRepository
}

// NewDownloadRemoteFileUseCase creates a new use case for downloading file data
func NewDownloadRemoteFileUseCase(
	logger *zap.Logger,
	repository remotefile.RemoteFileRepository,
) DownloadRemoteFileUseCase {
	return &downloadRemoteFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute downloads file data for a remote file
func (uc *downloadRemoteFileUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) ([]byte, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}

	// Download the file data
	data, err := uc.repository.DownloadFile(ctx, id)
	if err != nil {
		return nil, errors.NewAppError("failed to download file data", err)
	}

	return data, nil
}

// GetDownloadURL gets a pre-signed URL for downloading a file
func (uc *downloadRemoteFileUseCase) GetDownloadURL(
	ctx context.Context,
	id primitive.ObjectID,
) (string, error) {
	// Validate inputs
	if id.IsZero() {
		return "", errors.NewAppError("file ID is required", nil)
	}

	// Get the download URL
	url, err := uc.repository.GetDownloadURL(ctx, id)
	if err != nil {
		return "", errors.NewAppError("failed to get download URL", err)
	}

	return url, nil
}
