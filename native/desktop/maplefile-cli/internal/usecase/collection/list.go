// internal/usecase/collection/list.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

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
	logger = logger.Named("ListCollectionsUseCase")
	return &listCollectionsUseCase{
		logger:     logger,
		repository: repository,
	}
}

// ListCollectionsUseCase defines the interface for listing local collections
type ListCollectionsUseCase interface {
	Execute(ctx context.Context, filter collection.CollectionFilter) ([]*collection.Collection, error)
	ListRoots(ctx context.Context) ([]*collection.Collection, error)
	ListRootsByState(ctx context.Context, state string) ([]*collection.Collection, error)
	ListByParent(ctx context.Context, parentID gocql.UUID) ([]*collection.Collection, error)
	ListByParentAndState(ctx context.Context, parentID gocql.UUID, state string) ([]*collection.Collection, error)
	ListModifiedLocally(ctx context.Context) ([]*collection.Collection, error)
	ListByState(ctx context.Context, state string) ([]*collection.Collection, error)
	ListActiveCollections(ctx context.Context) ([]*collection.Collection, error)
	ListDeletedCollections(ctx context.Context) ([]*collection.Collection, error)
	ListArchivedCollections(ctx context.Context) ([]*collection.Collection, error)
}

// Execute lists local collections based on filter criteria
func (uc *listCollectionsUseCase) Execute(
	ctx context.Context,
	filter collection.CollectionFilter,
) ([]*collection.Collection, error) {
	// Validate state filter if provided
	if filter.State != nil {
		if err := collection.ValidateState(*filter.State); err != nil {
			uc.logger.Error("Invalid state filter provided",
				zap.String("state", *filter.State),
				zap.Error(err))
			return nil, errors.NewAppError("invalid state filter", err)
		}
	}

	// Get collections from repository
	collections, err := uc.repository.List(ctx, filter)
	if err != nil {
		return nil, errors.NewAppError("failed to list local collections", err)
	}

	return collections, nil
}

// ListRoots lists root-level local collections (no parent) - only active by default
func (uc *listCollectionsUseCase) ListRoots(
	ctx context.Context,
) ([]*collection.Collection, error) {
	return uc.ListRootsByState(ctx, collection.CollectionStateActive)
}

// ListRootsByState lists root-level local collections with specific state
func (uc *listCollectionsUseCase) ListRootsByState(
	ctx context.Context,
	state string,
) ([]*collection.Collection, error) {
	// Validate state
	if err := collection.ValidateState(state); err != nil {
		uc.logger.Error("Invalid state provided for listing roots",
			zap.String("state", state),
			zap.Error(err))
		return nil, errors.NewAppError("invalid state", err)
	}

	// Create a filter for root collections with specific state
	filter := collection.CollectionFilter{
		State: &state,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}

// ListByParent lists local collections with the specified parent - only active by default
func (uc *listCollectionsUseCase) ListByParent(
	ctx context.Context,
	parentID gocql.UUID,
) ([]*collection.Collection, error) {
	return uc.ListByParentAndState(ctx, parentID, collection.CollectionStateActive)
}

// ListByParentAndState lists local collections with the specified parent and state
func (uc *listCollectionsUseCase) ListByParentAndState(
	ctx context.Context,
	parentID gocql.UUID,
	state string,
) ([]*collection.Collection, error) {
	// Validate inputs
	if parentID.String() == "" {
		return nil, errors.NewAppError("parent ID is required", nil)
	}

	// Validate state
	if err := collection.ValidateState(state); err != nil {
		uc.logger.Error("Invalid state provided for listing by parent",
			zap.String("parentID", parentID.String()),
			zap.String("state", state),
			zap.Error(err))
		return nil, errors.NewAppError("invalid state", err)
	}

	// Create a filter for collections with the specified parent and state
	filter := collection.CollectionFilter{
		ParentID: &parentID,
		State:    &state,
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

// ListByState lists all collections with a specific state
func (uc *listCollectionsUseCase) ListByState(
	ctx context.Context,
	state string,
) ([]*collection.Collection, error) {
	// Validate state
	if err := collection.ValidateState(state); err != nil {
		uc.logger.Error("Invalid state provided for listing by state",
			zap.String("state", state),
			zap.Error(err))
		return nil, errors.NewAppError("invalid state", err)
	}

	// Create a filter for collections with specific state
	filter := collection.CollectionFilter{
		State: &state,
	}

	// Call the main execution method
	return uc.Execute(ctx, filter)
}

// ListActiveCollections lists all active collections
func (uc *listCollectionsUseCase) ListActiveCollections(
	ctx context.Context,
) ([]*collection.Collection, error) {
	return uc.ListByState(ctx, collection.CollectionStateActive)
}

// ListDeletedCollections lists all deleted collections
func (uc *listCollectionsUseCase) ListDeletedCollections(
	ctx context.Context,
) ([]*collection.Collection, error) {
	return uc.ListByState(ctx, collection.CollectionStateDeleted)
}

// ListArchivedCollections lists all archived collections
func (uc *listCollectionsUseCase) ListArchivedCollections(
	ctx context.Context,
) ([]*collection.Collection, error) {
	return uc.ListByState(ctx, collection.CollectionStateArchived)
}
