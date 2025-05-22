// internal/usecase/remotefile/upload.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// UploadRemoteFileUseCase defines the interface for uploading file data
type UploadRemoteFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID, data []byte) error
}

// uploadRemoteFileUseCase implements the UploadRemoteFileUseCase interface
type uploadRemoteFileUseCase struct {
	logger     *zap.Logger
	repository remotefile.RemoteFileRepository
}

// NewUploadRemoteFileUseCase creates a new use case for uploading file data
func NewUploadRemoteFileUseCase(
	logger *zap.Logger,
	repository remotefile.RemoteFileRepository,
) UploadRemoteFileUseCase {
	return &uploadRemoteFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute uploads file data for a remote file
func (uc *uploadRemoteFileUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
	data []byte,
) error {
	// Validate inputs
	if id.IsZero() {
		return errors.NewAppError("remote file ID is required", nil)
	}

	if data == nil || len(data) == 0 {
		return errors.NewAppError("remote file data is required", nil)
	}

	// Upload the file data
	if err := uc.repository.UploadFileByRemoteID(ctx, id, data); err != nil {
		return errors.NewAppError("failed to upload remote file data", err)
	}

	return nil
}
