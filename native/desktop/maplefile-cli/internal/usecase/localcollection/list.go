// internal/usecase/localcollection/list.go
package localcollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// ListLocalCollectionsUseCase defines the interface for listing local collections
type ListLocalCollectionsUseCase interface {
	Execute(ctx context.Context, filter localcollection.LocalCollectionFilter) ([]*localcollection.LocalCollection, error)
	ListRoots(ctx context.Context) ([]*localcollection.LocalCollection, error)
	ListByParent(ctx context.Context, parentID primitive.ObjectID) ([]*localcollection.LocalCollection, error)
	ListModifiedLocally(ctx context.Context) ([]*localcollection.LocalCollection, error)
}

// listLocalCollectionsUseCase implements the ListLocalCollectionsUseCase interface
type listLocalCollectionsUseCase struct {
	logger     *zap.Logger
	repository localcollection.LocalCollectionRepository
}

// NewListLocalCollectionsUseCase creates a new use case for listing local collections
func NewListLocalCollectionsUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
) ListLocalCollectionsUseCase {
	return &listLocalCollectionsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists local collections based on filter criteria
func (uc *listLocalCollectionsUseCase) Execute(
	ctx context.Context,
	filter localcollection.LocalCollectionFilter,
) ([]*localcollection.LocalCollection, error) {
	// Get collections from repository
	collections, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list local collections", err)
	}

	return collections, nil
}

// ListRoots lists root-level local collections (no parent)
func (uc *listLocalCollectionsUseCase) ListRoots(
	ctx context.Context,
) ([]*localcollection.LocalCollection, error) {
	// Create a filter for root collections (parent ID is nil)
	emptyID := primitive.NilObjectID
	filter := localcollection.LocalCollectionFilter{
		ParentID: &emptyID,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}

// ListByParent lists local collections with the specified parent
func (uc *listLocalCollectionsUseCase) ListByParent(
	ctx context.Context,
	parentID primitive.ObjectID,
) ([]*localcollection.LocalCollection, error) {
	// Validate inputs
	if parentID.IsZero() {
		return nil, errors.NewAppError("parent ID is required", nil)
	}

	// Create a filter for collections with the specified parent
	filter := localcollection.LocalCollectionFilter{
		ParentID: &parentID,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}

// ListModifiedLocally lists local collections that have been modified locally
func (uc *listLocalCollectionsUseCase) ListModifiedLocally(
	ctx context.Context,
) ([]*localcollection.LocalCollection, error) {
	// Create a filter for locally modified collections
	status := localcollection.SyncStatusModifiedLocally
	filter := localcollection.LocalCollectionFilter{
		SyncStatus: &status,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}
