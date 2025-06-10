// internal/usecase/file/list_files_by_collection.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
)

// ListFilesByCollectionUseCase defines the interface for listing files by collection
type ListFilesByCollectionUseCase interface {
	Execute(ctx context.Context, collectionID gocql.UUID) ([]*dom_file.File, error)
}

// listFilesByCollectionUseCase implements the ListFilesByCollectionUseCase interface
type listFilesByCollectionUseCase struct {
	logger     *zap.Logger
	repository dom_file.FileRepository
}

// NewListFilesByCollectionUseCase creates a new use case for listing files by collection
func NewListFilesByCollectionUseCase(
	logger *zap.Logger,
	repository dom_file.FileRepository,
) ListFilesByCollectionUseCase {
	logger = logger.Named("ListFilesByCollectionUseCase")
	return &listFilesByCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists local files within a specific collection
func (uc *listFilesByCollectionUseCase) Execute(
	ctx context.Context,
	collectionID gocql.UUID,
) ([]*dom_file.File, error) {
	// Validate inputs
	if collectionID.String() == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get files from repository
	files, err := uc.repository.GetByCollection(ctx, collectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to list files by collection", err)
	}

	return files, nil
}
