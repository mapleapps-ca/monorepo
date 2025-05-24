// internal/usecase/localfile/delete_files_by_collection.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	fileUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// DeleteFilesByCollectionUseCase defines the interface for deleting files by collection
type DeleteFilesByCollectionUseCase interface {
	Execute(ctx context.Context, collectionID primitive.ObjectID) error
}

// deleteFilesByCollectionUseCase implements the DeleteFilesByCollectionUseCase interface
type deleteFilesByCollectionUseCase struct {
	logger            *zap.Logger
	fileDeleteUseCase fileUseCase.DeleteFilesUseCase
}

// NewDeleteFilesByCollectionUseCase creates a new use case for deleting files by collection
func NewDeleteFilesByCollectionUseCase(
	logger *zap.Logger,
	fileDeleteUseCase fileUseCase.DeleteFilesUseCase,
) DeleteFilesByCollectionUseCase {
	return &deleteFilesByCollectionUseCase{
		logger:            logger,
		fileDeleteUseCase: fileDeleteUseCase,
	}
}

// Execute deletes all files in a collection
func (uc *deleteFilesByCollectionUseCase) Execute(
	ctx context.Context,
	collectionID primitive.ObjectID,
) error {
	uc.logger.Debug("Deleting files by collection", zap.String("collectionID", collectionID.Hex()))

	err := uc.fileDeleteUseCase.DeleteByCollection(ctx, collectionID)
	if err != nil {
		return errors.NewAppError("failed to delete files by collection", err)
	}

	return nil
}
