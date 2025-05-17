// internal/usecase/localcollection/delete.go
package localcollection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// DeleteLocalCollectionUseCase defines the interface for deleting a local collection
type DeleteLocalCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
	DeleteWithChildren(ctx context.Context, id primitive.ObjectID) error
}

// deleteLocalCollectionUseCase implements the DeleteLocalCollectionUseCase interface
type deleteLocalCollectionUseCase struct {
	logger      *zap.Logger
	repository  localcollection.LocalCollectionRepository
	listUseCase ListLocalCollectionsUseCase
}

// NewDeleteLocalCollectionUseCase creates a new use case for deleting local collections
func NewDeleteLocalCollectionUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
	listUseCase ListLocalCollectionsUseCase,
) DeleteLocalCollectionUseCase {
	return &deleteLocalCollectionUseCase{
		logger:      logger,
		repository:  repository,
		listUseCase: listUseCase,
	}
}

// Execute deletes a local collection by ID
func (uc *deleteLocalCollectionUseCase) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Validate inputs
	if id.IsZero() {
		return errors.NewAppError("collection ID is required", nil)
	}

	// Delete the collection
	err := uc.repository.Delete(ctx, id)
	if err != nil {
		return errors.NewAppError("failed to delete local collection", err)
	}

	return nil
}

// DeleteWithChildren deletes a local collection and all its child collections
func (uc *deleteLocalCollectionUseCase) DeleteWithChildren(
	ctx context.Context,
	id primitive.ObjectID,
) error {
	// Validate inputs
	if id.IsZero() {
		return errors.NewAppError("collection ID is required", nil)
	}

	// Get all children of this collection
	children, err := uc.listUseCase.ListByParent(ctx, id)
	if err != nil {
		return errors.NewAppError("failed to list child collections", err)
	}

	// Delete each child recursively
	for _, child := range children {
		err = uc.DeleteWithChildren(ctx, child.ID)
		if err != nil {
			return err
		}
	}

	// Delete the collection itself
	return uc.Execute(ctx, id)
}
