// internal/usecase/collection/delete.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// DeleteCollectionUseCase defines the interface for deleting a local collection
type DeleteCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
	DeleteWithChildren(ctx context.Context, id primitive.ObjectID) error
}

// deleteCollectionUseCase implements the DeleteCollectionUseCase interface
type deleteCollectionUseCase struct {
	logger      *zap.Logger
	repository  collection.CollectionRepository
	listUseCase ListCollectionsUseCase
}

// NewDeleteCollectionUseCase creates a new use case for deleting local collections
func NewDeleteCollectionUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
	listUseCase ListCollectionsUseCase,
) DeleteCollectionUseCase {
	logger = logger.Named("DeleteCollectionUseCase")
	return &deleteCollectionUseCase{
		logger:      logger,
		repository:  repository,
		listUseCase: listUseCase,
	}
}

// Execute deletes a local collection by ID
func (uc *deleteCollectionUseCase) Execute(
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
func (uc *deleteCollectionUseCase) DeleteWithChildren(
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
