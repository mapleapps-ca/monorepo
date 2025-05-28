// internal/usecase/collection/update.go
package collection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

// UpdateCollectionInput defines the input for updating a local collection
type UpdateCollectionInput struct {
	ID             primitive.ObjectID
	EncryptedName  *string
	DecryptedName  *string
	CollectionType *string
	State          *string
}

// UpdateCollectionUseCase defines the interface for updating a local collection
type UpdateCollectionUseCase interface {
	Execute(ctx context.Context, input UpdateCollectionInput) (*dom_collection.Collection, error)
}

// updateCollectionUseCase implements the UpdateCollectionUseCase interface
type updateCollectionUseCase struct {
	logger     *zap.Logger
	repository dom_collection.CollectionRepository
	getUseCase GetCollectionUseCase
}

// NewUpdateCollectionUseCase creates a new use case for updating local collections
func NewUpdateCollectionUseCase(
	logger *zap.Logger,
	repository dom_collection.CollectionRepository,
	getUseCase GetCollectionUseCase,
) UpdateCollectionUseCase {
	logger = logger.Named("UpdateCollectionUseCase")
	return &updateCollectionUseCase{
		logger:     logger,
		repository: repository,
		getUseCase: getUseCase,
	}
}

// Execute updates a local collection
func (uc *updateCollectionUseCase) Execute(
	ctx context.Context,
	input UpdateCollectionInput,
) (*dom_collection.Collection, error) {
	// Validate inputs
	if input.ID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get the existing collection
	collection, err := uc.getUseCase.Execute(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Track original state for validation
	originalState := collection.State

	// Update fields if provided
	if input.EncryptedName != nil {
		collection.EncryptedName = *input.EncryptedName
	}

	if input.DecryptedName != nil {
		collection.Name = *input.DecryptedName
	}

	if input.CollectionType != nil {
		if *input.CollectionType != dom_collection.CollectionTypeFolder && *input.CollectionType != dom_collection.CollectionTypeAlbum {
			return nil, errors.NewAppError("invalid collection type", nil)
		}
		collection.CollectionType = *input.CollectionType
	}

	// Handle state updates with validation
	if input.State != nil {
		newState := *input.State

		// Validate the new state
		if err := dom_collection.ValidateState(newState); err != nil {
			uc.logger.Error("Invalid state provided in update",
				zap.String("collectionID", input.ID.Hex()),
				zap.String("newState", newState),
				zap.Error(err))
			return nil, errors.NewAppError("invalid collection state", err)
		}

		// Validate state transition
		if originalState != newState {
			if err := dom_collection.IsValidStateTransition(originalState, newState); err != nil {
				uc.logger.Error("Invalid state transition attempted",
					zap.String("collectionID", input.ID.Hex()),
					zap.String("fromState", originalState),
					zap.String("toState", newState),
					zap.Error(err))
				return nil, errors.NewAppError("invalid state transition", err)
			}

			uc.logger.Info("Collection state transition",
				zap.String("collectionID", input.ID.Hex()),
				zap.String("fromState", originalState),
				zap.String("toState", newState))
		}

		collection.State = newState
	}

	// Update timestamps and modification status
	collection.ModifiedAt = time.Now()
	// collection.IsModifiedLocally = true // Figure out what to do here.

	// Save the updated collection
	err = uc.repository.Save(ctx, collection)
	if err != nil {
		return nil, errors.NewAppError("failed to update local collection", err)
	}

	uc.logger.Info("Collection updated successfully",
		zap.String("collectionID", input.ID.Hex()),
		zap.String("state", collection.State))

	return collection, nil
}
