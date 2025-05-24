// internal/usecase/file/delete_file.go
package file

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// DeleteFileUseCase defines the interface for deleting a local file
type DeleteFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

// deleteFileUseCase implements the DeleteFileUseCase interface
type deleteFileUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewDeleteFileUseCase creates a new use case for deleting local files
func NewDeleteFileUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) DeleteFileUseCase {
	return &deleteFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute deletes a local file by ID
func (uc *deleteFileUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Validate inputs
	if id.IsZero() {
		return errors.NewAppError("file ID is required", nil)
	}

	// Delete the file
	err := uc.repository.Delete(ctx, id)
	if err != nil {
		return errors.NewAppError("failed to delete local file", err)
	}

	return nil
}
