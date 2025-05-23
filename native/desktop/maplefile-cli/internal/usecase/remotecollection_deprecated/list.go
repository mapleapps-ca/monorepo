// internal/usecase/remotecollection/list.go
package remotecollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

// ListRemoteCollectionsUseCase defines the interface for listing cloud collections
type ListRemoteCollectionsUseCase interface {
	Execute(ctx context.Context, filter remotecollection.CollectionFilter) ([]*remotecollection.RemoteCollection, error)
	ListRoots(ctx context.Context) ([]*remotecollection.RemoteCollection, error)
	ListByParent(ctx context.Context, parentID primitive.ObjectID) ([]*remotecollection.RemoteCollection, error)
}

// listRemoteCollectionsUseCase implements the ListRemoteCollectionsUseCase interface
type listRemoteCollectionsUseCase struct {
	logger     *zap.Logger
	repository remotecollection.RemoteCollectionRepository
}

// NewListRemoteCollectionsUseCase creates a new use case for listing cloud collections
func NewListRemoteCollectionsUseCase(
	logger *zap.Logger,
	repository remotecollection.RemoteCollectionRepository,
) ListRemoteCollectionsUseCase {
	return &listRemoteCollectionsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists cloud collections based on filter criteria
func (uc *listRemoteCollectionsUseCase) Execute(
	ctx context.Context,
	filter remotecollection.CollectionFilter,
) ([]*remotecollection.RemoteCollection, error) {
	// List collections from the repository
	collections, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list cloud collections", err)
	}

	return collections, nil
}

// ListRoots lists root-level cloud collections
func (uc *listRemoteCollectionsUseCase) ListRoots(
	ctx context.Context,
) ([]*remotecollection.RemoteCollection, error) {
	// Create a filter for root collections (parent ID is nil)
	emptyID := primitive.NilObjectID
	filter := remotecollection.CollectionFilter{
		ParentID: &emptyID,
	}

	// Execute the main method
	return uc.Execute(ctx, filter)
}

// ListByParent lists cloud collections with the specified parent
func (uc *listRemoteCollectionsUseCase) ListByParent(
	ctx context.Context,
	parentID primitive.ObjectID,
) ([]*remotecollection.RemoteCollection, error) {
	// Validate inputs
	if parentID.IsZero() {
		return nil, errors.NewAppError("parent ID is required", nil)
	}

	// Create a filter for collections with the specified parent
	filter := remotecollection.CollectionFilter{
		ParentID: &parentID,
	}

	// Execute the main method
	return uc.Execute(ctx, filter)
}
