// internal/usecase/file/delete.go
package file

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// DeleteFileUseCase defines the interface for deleting a local file
type DeleteFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

// DeleteFilesUseCase defines the interface for deleting multiple local files
type DeleteFilesUseCase interface {
	Execute(ctx context.Context, ids []primitive.ObjectID) error
	DeleteByCollection(ctx context.Context, collectionID primitive.ObjectID) error
}

// deleteFileUseCase implements the DeleteFileUseCase interface
type deleteFileUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// deleteFilesUseCase implements the DeleteFilesUseCase interface
type deleteFilesUseCase struct {
	logger      *zap.Logger
	repository  file.FileRepository
	listUseCase ListFilesUseCase
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

// NewDeleteFilesUseCase creates a new use case for deleting multiple local files
func NewDeleteFilesUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
	listUseCase ListFilesUseCase,
) DeleteFilesUseCase {
	return &deleteFilesUseCase{
		logger:      logger,
		repository:  repository,
		listUseCase: listUseCase,
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

// Execute deletes multiple local files by IDs
func (uc *deleteFilesUseCase) Execute(
	ctx context.Context,
	ids []primitive.ObjectID,
) error {
	if len(ids) == 0 {
		return errors.NewAppError("at least one file ID is required", nil)
	}

	// Validate all IDs
	for i, id := range ids {
		if id.IsZero() {
			return errors.NewAppError(fmt.Sprintf("file ID at index %d is invalid", i), nil)
		}
	}

	// Delete the files
	err := uc.repository.DeleteMany(ctx, ids)
	if err != nil {
		return errors.NewAppError("failed to delete local files", err)
	}

	return nil
}

// DeleteByCollection deletes all files in a collection
func (uc *deleteFilesUseCase) DeleteByCollection(
	ctx context.Context,
	collectionID primitive.ObjectID,
) error {
	// Validate inputs
	if collectionID.IsZero() {
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get all files in the collection
	files, err := uc.listUseCase.ListByCollection(ctx, collectionID)
	if err != nil {
		return errors.NewAppError("failed to list files in collection", err)
	}

	if len(files) == 0 {
		// No files to delete
		return nil
	}

	// Extract file IDs
	fileIDs := make([]primitive.ObjectID, len(files))
	for i, file := range files {
		fileIDs[i] = file.ID
	}

	// Delete all files
	return uc.Execute(ctx, fileIDs)
}
