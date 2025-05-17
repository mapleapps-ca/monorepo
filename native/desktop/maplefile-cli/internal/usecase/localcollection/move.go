// internal/usecase/localcollection/move.go
package localcollection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// MoveLocalCollectionInput defines the input for moving a local collection
type MoveLocalCollectionInput struct {
	ID          primitive.ObjectID
	NewParentID primitive.ObjectID
	// New encrypted path segments if needed
	EncryptedPathSegments []string
}

// MoveLocalCollectionUseCase defines the interface for moving a local collection
type MoveLocalCollectionUseCase interface {
	Execute(ctx context.Context, input MoveLocalCollectionInput) (*localcollection.LocalCollection, error)
}

// moveLocalCollectionUseCase implements the MoveLocalCollectionUseCase interface
type moveLocalCollectionUseCase struct {
	logger     *zap.Logger
	repository localcollection.LocalCollectionRepository
	getUseCase GetLocalCollectionUseCase
}

// NewMoveLocalCollectionUseCase creates a new use case for moving local collections
func NewMoveLocalCollectionUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
	getUseCase GetLocalCollectionUseCase,
) MoveLocalCollectionUseCase {
	return &moveLocalCollectionUseCase{
		logger:     logger,
		repository: repository,
		getUseCase: getUseCase,
	}
}

// Execute moves a local collection to a new parent
func (uc *moveLocalCollectionUseCase) Execute(
	ctx context.Context,
	input MoveLocalCollectionInput,
) (*localcollection.LocalCollection, error) {
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
