// internal/usecase/remotefile/list.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// ListRemoteFilesUseCase defines the interface for listing remote files
type ListRemoteFilesUseCase interface {
	Execute(ctx context.Context, filter remotefile.RemoteFileFilter) ([]*remotefile.RemoteFile, error)
	ByCollection(ctx context.Context, collectionID primitive.ObjectID) ([]*remotefile.RemoteFile, error)
}

// listRemoteFilesUseCase implements the ListRemoteFilesUseCase interface
type listRemoteFilesUseCase struct {
	logger     *zap.Logger
	repository remotefile.RemoteFileRepository
}

// NewListRemoteFilesUseCase creates a new use case for listing remote files
func NewListRemoteFilesUseCase(
	logger *zap.Logger,
	repository remotefile.RemoteFileRepository,
) ListRemoteFilesUseCase {
	return &listRemoteFilesUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists remote files based on filter criteria
func (uc *listRemoteFilesUseCase) Execute(
	ctx context.Context,
	filter remotefile.RemoteFileFilter,
) ([]*remotefile.RemoteFile, error) {
	// List files using the repository
	files, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list remote files", err)
	}

	return files, nil
}

// ByCollection lists remote files within a specific collection
func (uc *listRemoteFilesUseCase) ByCollection(
	ctx context.Context,
	collectionID primitive.ObjectID,
) ([]*remotefile.RemoteFile, error) {
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
