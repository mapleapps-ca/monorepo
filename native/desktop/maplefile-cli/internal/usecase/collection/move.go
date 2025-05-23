// internal/usecase/collection/move.go
package collection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// MoveCollectionInput defines the input for moving a local collection
type MoveCollectionInput struct {
	ID          primitive.ObjectID
	NewParentID primitive.ObjectID
	// New encrypted path segments if needed
	EncryptedPathSegments []string
}

// MoveCollectionUseCase defines the interface for moving a local collection
type MoveCollectionUseCase interface {
	Execute(ctx context.Context, input MoveCollectionInput) (*collection.Collection, error)
}

// moveCollectionUseCase implements the MoveCollectionUseCase interface
type moveCollectionUseCase struct {
	logger     *zap.Logger
	repository collection.CollectionRepository
	getUseCase GetCollectionUseCase
}

// NewMoveCollectionUseCase creates a new use case for moving local collections
func NewMoveCollectionUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
	getUseCase GetCollectionUseCase,
) MoveCollectionUseCase {
	return &moveCollectionUseCase{
		logger:     logger,
		repository: repository,
		getUseCase: getUseCase,
	}
}

// Execute moves a local collection to a new parent
func (uc *moveCollectionUseCase) Execute(
	ctx context.Context,
	input MoveCollectionInput,
) (*collection.Collection, error) {
	// Validate inputs
	if input.ID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.NewParentID.IsZero() {
		return nil, errors.NewAppError("new parent ID is required", nil)
	}

	// Get the collection to move
	collection, err := uc.getUseCase.Execute(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Ensure we're not creating a circular reference
	if input.ID == input.NewParentID {
		return nil, errors.NewAppError("cannot move collection to itself", nil)
	}

	// Verify the new parent exists
	newParent, err := uc.getUseCase.Execute(ctx, input.NewParentID)
	if err != nil {
		return nil, errors.NewAppError("failed to retrieve new parent collection", err)
	}
	if newParent == nil {
		return nil, errors.NewAppError("new parent collection does not exist", nil)
	}

	// Update the collection's parent ID
	collection.ParentID = input.NewParentID

	// Update path segments if provided
	if len(input.EncryptedPathSegments) > 0 {
		collection.EncryptedPathSegments = input.EncryptedPathSegments
	}

	// Update timestamps and modification status
	collection.ModifiedAt = time.Now()
	collection.IsModifiedLocally = true

	// Save the updated collection
	err = uc.repository.Save(ctx, collection)
	if err != nil {
		return nil, errors.NewAppError("failed to save moved collection", err)
	}

	return collection, nil
}
