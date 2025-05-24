// internal/usecase/localfile/delete_files.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// DeleteFilesUseCase defines the interface for deleting multiple local files
type DeleteFilesUseCase interface {
	Execute(ctx context.Context, fileIDs []primitive.ObjectID) error
}

// deleteFilesUseCase implements the DeleteFilesUseCase interface
type deleteFilesUseCase struct {
	logger            *zap.Logger
	fileDeleteUseCase fileUseCase.DeleteFilesUseCase
}

// NewDeleteFilesUseCase creates a new use case for deleting multiple local files
func NewDeleteFilesUseCase(
	logger *zap.Logger,
	fileDeleteUseCase fileUseCase.DeleteFilesUseCase,
) DeleteFilesUseCase {
	return &deleteFilesUseCase{
		logger:            logger,
		fileDeleteUseCase: fileDeleteUseCase,
	}
}

// Execute deletes multiple local files
func (uc *deleteFilesUseCase) Execute(
	ctx context.Context,
	fileIDs []primitive.ObjectID,
) error {
	uc.logger.Debug("Deleting multiple local files", zap.Int("count", len(fileIDs)))

	err := uc.fileDeleteUseCase.Execute(ctx, fileIDs)
	if err != nil {
		return errors.NewAppError("failed to delete local files", err)
	}

	return nil
}
