// internal/usecase/file/delete_files.go
package file

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// DeleteFilesUseCase defines the interface for deleting multiple local files
type DeleteFilesUseCase interface {
	Execute(ctx context.Context, ids []gocql.UUID) error
	DeleteByCollection(ctx context.Context, collectionID gocql.UUID) error
}

// deleteFilesUseCase implements the DeleteFilesUseCase interface
type deleteFilesUseCase struct {
	logger      *zap.Logger
	repository  dom_file.FileRepository
	listUseCase ListFilesByCollectionUseCase
}

// NewDeleteFilesUseCase creates a new use case for deleting multiple local files
func NewDeleteFilesUseCase(
	logger *zap.Logger,
	repository dom_file.FileRepository,
	listUseCase ListFilesByCollectionUseCase,
) DeleteFilesUseCase {
	logger = logger.Named("DeleteFilesUseCase")
	return &deleteFilesUseCase{
		logger:      logger,
		repository:  repository,
		listUseCase: listUseCase,
	}
}

// Execute deletes multiple local files by IDs
func (uc *deleteFilesUseCase) Execute(
	ctx context.Context,
	ids []gocql.UUID,
) error {
	if len(ids) == 0 {
		return errors.NewAppError("at least one file ID is required", nil)
	}

	// Validate all IDs
	for i, id := range ids {
		if id.String() == "" {
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
	collectionID gocql.UUID,
) error {
	// Validate inputs
	if collectionID.String() == "" {
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get all files in the collection
	files, err := uc.listUseCase.Execute(ctx, collectionID)
	if err != nil {
		return errors.NewAppError("failed to list files in collection", err)
	}

	if len(files) == 0 {
		// No files to delete
		return nil
	}

	// Extract file IDs
	fileIDs := make([]gocql.UUID, len(files))
	for i, file := range files {
		fileIDs[i] = file.ID
	}

	// Delete all files
	return uc.Execute(ctx, fileIDs)
}
