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
	ID                    primitive.ObjectID
	EncryptedName         *string
	DecryptedName         *string
	Type                  *string
	EncryptedPathSegments *[]string
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

	// Update fields if provided
	if input.EncryptedName != nil {
		collection.EncryptedName = *input.EncryptedName
	}

	if input.DecryptedName != nil {
		collection.DecryptedName = *input.DecryptedName
	}

	if input.Type != nil {
		if *input.Type != dom_collection.CollectionTypeFolder && *input.Type != dom_collection.CollectionTypeAlbum {
			return nil, errors.NewAppError("invalid collection type", nil)
		}
		collection.Type = *input.Type
	}

	if input.EncryptedPathSegments != nil {
		collection.EncryptedPathSegments = *input.EncryptedPathSegments
	}

	// Update timestamps and modification status
	collection.ModifiedAt = time.Now()
	collection.IsModifiedLocally = true

	// Save the updated collection
	err = uc.repository.Save(ctx, collection)
	if err != nil {
		return nil, errors.NewAppError("failed to update local collection", err)
	}

	return collection, nil
}
