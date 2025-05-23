// internal/usecase/collection/list.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// ListCollectionsUseCase defines the interface for listing local collections
type ListCollectionsUseCase interface {
	Execute(ctx context.Context, filter collection.CollectionFilter) ([]*collection.Collection, error)
	ListRoots(ctx context.Context) ([]*collection.Collection, error)
	ListByParent(ctx context.Context, parentID primitive.ObjectID) ([]*collection.Collection, error)
	ListModifiedLocally(ctx context.Context) ([]*collection.Collection, error)
}

// listCollectionsUseCase implements the ListCollectionsUseCase interface
type listCollectionsUseCase struct {
	logger     *zap.Logger
	repository collection.CollectionRepository
}

// NewListCollectionsUseCase creates a new use case for listing local collections
func NewListCollectionsUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
) ListCollectionsUseCase {
	return &listCollectionsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute lists local collections based on filter criteria
func (uc *listCollectionsUseCase) Execute(
	ctx context.Context,
	filter collection.CollectionFilter,
) ([]*collection.Collection, error) {
	// Get collections from repository
	collections, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list local collections", err)
	}

	return collections, nil
}

// ListRoots lists root-level local collections (no parent)
func (uc *listCollectionsUseCase) ListRoots(
	ctx context.Context,
) ([]*collection.Collection, error) {
	// Create a filter for root collections (parent ID is nil)
	emptyID := primitive.NilObjectID
	filter := collection.CollectionFilter{
		ParentID: &emptyID,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}

// ListByParent lists local collections with the specified parent
func (uc *listCollectionsUseCase) ListByParent(
	ctx context.Context,
	parentID primitive.ObjectID,
) ([]*collection.Collection, error) {
	// Validate inputs
	if parentID.IsZero() {
		return nil, errors.NewAppError("parent ID is required", nil)
	}

	// Create a filter for collections with the specified parent
	filter := collection.CollectionFilter{
		ParentID: &parentID,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}

// ListModifiedLocally lists local collections that have been modified locally
func (uc *listCollectionsUseCase) ListModifiedLocally(
	ctx context.Context,
) ([]*collection.Collection, error) {
	// Create a filter for locally modified collections
	status := collection.SyncStatusModifiedLocally
	filter := collection.CollectionFilter{
		SyncStatus: &status,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}
