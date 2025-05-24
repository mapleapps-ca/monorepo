// internal/usecase/localfile/delete_file.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// DeleteFileUseCase defines the interface for deleting a local file
type DeleteFileUseCase interface {
	Execute(ctx context.Context, fileID primitive.ObjectID) error
}

// deleteFileUseCase implements the DeleteFileUseCase interface
type deleteFileUseCase struct {
	logger            *zap.Logger
	fileDeleteUseCase fileUseCase.DeleteFileUseCase
}

// NewDeleteFileUseCase creates a new use case for deleting a local file
func NewDeleteFileUseCase(
	logger *zap.Logger,
	fileDeleteUseCase fileUseCase.DeleteFileUseCase,
) DeleteFileUseCase {
	return &deleteFileUseCase{
		logger:            logger,
		fileDeleteUseCase: fileDeleteUseCase,
	}
}

// Execute deletes a single local file
func (uc *deleteFileUseCase) Execute(
	ctx context.Context,
	fileID primitive.ObjectID,
) error {
	uc.logger.Debug("Deleting local file", zap.String("fileID", fileID.Hex()))

	err := uc.fileDeleteUseCase.Execute(ctx, fileID)
	if err != nil {
		return errors.NewAppError("failed to delete local file", err)
	}

	return nil
}
