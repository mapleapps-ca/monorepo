// internal/usecase/file/list.go
package file

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// ListFilesUseCase defines the interface for listing local files
type ListFilesUseCase interface {
	ListByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*file.File, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*file.File, error)
}

// listFilesUseCase implements the ListFilesUseCase interface
type listFilesUseCase struct {
	logger     *zap.Logger
	repository file.FileRepository
}

// NewListFilesUseCase creates a new use case for listing local files
func NewListFilesUseCase(
	logger *zap.Logger,
	repository file.FileRepository,
) ListFilesUseCase {
	return &listFilesUseCase{
		logger:     logger,
		repository: repository,
	}
}

// ListByCollection lists local files within a specific collection
func (uc *listFilesUseCase) ListByCollection(
	ctx context.Context,
	collectionID primitive.ObjectID,
) ([]*file.File, error) {
	// Validate inputs
	if collectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get files from repository
	files, err := uc.repository.GetByCollection(ctx, collectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to list files by collection", err)
	}

	return files, nil
}

// GetByIDs retrieves multiple local files by their IDs
func (uc *listFilesUseCase) GetByIDs(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]*file.File, error) {
	if len(ids) == 0 {
		return []*file.File{}, nil
	}

	// Validate all IDs
	for i, id := range ids {
		if id.IsZero() {
			return nil, errors.NewAppError(fmt.Sprintf("file ID at index %d is invalid", i), nil)
		}
	}

	// Get files from repository
	files, err := uc.repository.GetByIDs(ctx, ids)
	if err != nil {
		return nil, errors.NewAppError("failed to get files by IDs", err)
	}

	return files, nil
}
