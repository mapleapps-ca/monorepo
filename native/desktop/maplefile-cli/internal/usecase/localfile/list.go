// internal/usecase/localfile/list.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
)

// ListLocalFilesUseCase defines the interface for listing local files
type ListLocalFilesUseCase interface {
	Execute(ctx context.Context, filter localfile.LocalFileFilter) ([]*localfile.LocalFile, error)
	ByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*localfile.LocalFile, error)
	ModifiedLocally(ctx context.Context) ([]*localfile.LocalFile, error)
	Search(ctx context.Context, nameContains string, mimeType string) ([]*localfile.LocalFile, error)
}

// listLocalFilesUseCase implements the ListLocalFilesUseCase interface
type listLocalFilesUseCase struct {
	logger     *zap.Logger
	repository localfile.LocalFileRepository
}

// NewListLocalFilesUseCase creates a new use case for listing local files
func NewListLocalFilesUseCase(
	logger *zap.Logger,
	repository localfile.LocalFileRepository,
) ListLocalFilesUseCase {
	return &listLocalFilesUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists local files based on filter criteria
func (uc *listLocalFilesUseCase) Execute(
	ctx context.Context,
	filter localfile.LocalFileFilter,
) ([]*localfile.LocalFile, error) {
	// List files using the repository
	files, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list local files", err)
	}

	return files, nil
}

// ByCollection lists local files within a specific collection
func (uc *listLocalFilesUseCase) ByCollection(
	ctx context.Context,
	collectionID primitive.ObjectID,
) ([]*localfile.LocalFile, error) {
	// Validate inputs
	if collectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// List files by collection
	files, err := uc.repository.ListByCollection(ctx, collectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to list files by collection", err)
	}

	return files, nil
}

// ModifiedLocally lists locally modified files
func (uc *listLocalFilesUseCase) ModifiedLocally(
	ctx context.Context,
) ([]*localfile.LocalFile, error) {
	// Create a filter for modified files
	syncStatus := localfile.SyncStatusModifiedLocally
	filter := localfile.LocalFileFilter{
		SyncStatus: &syncStatus,
	}

	// List modified files
	files, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list modified files", err)
	}

	return files, nil
}

// Search searches for files by name and mime type
func (uc *listLocalFilesUseCase) Search(
	ctx context.Context,
	nameContains string,
	mimeType string,
) ([]*localfile.LocalFile, error) {
	// Create a filter for search
	filter := localfile.LocalFileFilter{
		NameContains: &nameContains,
	}

	// Add mime type filter if provided
	if mimeType != "" {
		filter.MimeType = &mimeType
	}

	// List files matching search criteria
	files, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to search for files", err)
	}

	return files, nil
}
