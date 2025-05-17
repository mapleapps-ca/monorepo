// internal/usecase/remotefile/delete.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// DeleteRemoteFileUseCase defines the interface for deleting a remote file
type DeleteRemoteFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
	DeleteMultiple(ctx context.Context, ids []primitive.ObjectID) (int, error)
}

// deleteRemoteFileUseCase implements the DeleteRemoteFileUseCase interface
type deleteRemoteFileUseCase struct {
	logger     *zap.Logger
	repository remotefile.RemoteFileRepository
}

// NewDeleteRemoteFileUseCase creates a new use case for deleting remote files
func NewDeleteRemoteFileUseCase(
	logger *zap.Logger,
	repository remotefile.RemoteFileRepository,
) DeleteRemoteFileUseCase {
	return &deleteRemoteFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute deletes a remote file
func (uc *deleteRemoteFileUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Validate inputs
	if id.IsZero() {
		return errors.NewAppError("file ID is required", nil)
	}

	// Delete the file
	if err := uc.repository.Delete(ctx, id); err != nil {
		return errors.NewAppError("failed to delete remote file", err)
	}

	return nil
}

// DeleteMultiple deletes multiple remote files
func (uc *deleteRemoteFileUseCase) DeleteMultiple(
	ctx context.Context,
	ids []primitive.ObjectID,
) (int, error) {
	if len(ids) == 0 {
		return 0, errors.NewAppError("no file IDs provided", nil)
	}

	// Delete each file
	successCount := 0
	for _, id := range ids {
		if err := uc.repository.Delete(ctx, id); err != nil {
			// Log error but continue with other files
			uc.logger.Error("Failed to delete remote file",
				zap.String("fileID", id.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
