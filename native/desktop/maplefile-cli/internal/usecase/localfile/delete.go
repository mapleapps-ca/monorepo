// internal/usecase/localfile/delete.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
)

// DeleteLocalFileUseCase defines the interface for deleting a local file
type DeleteLocalFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
	DeleteMultiple(ctx context.Context, ids []primitive.ObjectID) (int, error)
	DeleteByCollection(ctx context.Context, collectionID primitive.ObjectID) (int, error)
}

// deleteLocalFileUseCase implements the DeleteLocalFileUseCase interface
type deleteLocalFileUseCase struct {
	logger      *zap.Logger
	repository  localfile.LocalFileRepository
	listUseCase ListLocalFilesUseCase
}

// NewDeleteLocalFileUseCase creates a new use case for deleting local files
func NewDeleteLocalFileUseCase(
	logger *zap.Logger,
	repository localfile.LocalFileRepository,
	listUseCase ListLocalFilesUseCase,
) DeleteLocalFileUseCase {
	return &deleteLocalFileUseCase{
		logger:      logger,
		repository:  repository,
		listUseCase: listUseCase,
	}
}

// Execute deletes a local file
func (uc *deleteLocalFileUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Validate inputs
	if id.IsZero() {
		return errors.NewAppError("file ID is required", nil)
	}

	// Delete the file
	if err := uc.repository.Delete(ctx, id); err != nil {
		return errors.NewAppError("failed to delete local file", err)
	}

	return nil
}

// DeleteMultiple deletes multiple local files
func (uc *deleteLocalFileUseCase) DeleteMultiple(
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
			uc.logger.Error("Failed to delete file",
				zap.String("fileID", id.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}

// DeleteByCollection deletes all files in a collection
func (uc *deleteLocalFileUseCase) DeleteByCollection(
	ctx context.Context,
	collectionID primitive.ObjectID,
) (int, error) {
	// Validate inputs
	if collectionID.IsZero() {
		return 0, errors.NewAppError("collection ID is required", nil)
	}

	// Get all files in the collection
	files, err := uc.listUseCase.ByCollection(ctx, collectionID)
	if err != nil {
		return 0, errors.NewAppError("failed to list files in collection", err)
	}

	// Delete each file
	successCount := 0
	for _, file := range files {
		if err := uc.repository.Delete(ctx, file.ID); err != nil {
			// Log error but continue with other files
			uc.logger.Error("Failed to delete file in collection",
				zap.String("fileID", file.ID.Hex()),
				zap.String("collectionID", collectionID.Hex()),
				zap.Error(err))
			continue
		}
		successCount++
	}

	return successCount, nil
}
