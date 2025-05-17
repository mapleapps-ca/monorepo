// internal/usecase/localcollection/update.go
package localcollection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// UpdateLocalCollectionInput defines the input for updating a local collection
type UpdateLocalCollectionInput struct {
	ID                    primitive.ObjectID
	EncryptedName         *string
	DecryptedName         *string
	Type                  *string
	EncryptedPathSegments *[]string
}

// UpdateLocalCollectionUseCase defines the interface for updating a local collection
type UpdateLocalCollectionUseCase interface {
	Execute(ctx context.Context, input UpdateLocalCollectionInput) (*localcollection.LocalCollection, error)
}

// updateLocalCollectionUseCase implements the UpdateLocalCollectionUseCase interface
type updateLocalCollectionUseCase struct {
	logger     *zap.Logger
	repository localcollection.LocalCollectionRepository
	getUseCase GetLocalCollectionUseCase
}

// NewUpdateLocalCollectionUseCase creates a new use case for updating local collections
func NewUpdateLocalCollectionUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
	getUseCase GetLocalCollectionUseCase,
) UpdateLocalCollectionUseCase {
	return &updateLocalCollectionUseCase{
		logger:     logger,
		repository: repository,
		getUseCase: getUseCase,
	}
}

// Execute updates a local collection
func (uc *updateLocalCollectionUseCase) Execute(
	ctx context.Context,
	input UpdateLocalCollectionInput,
) (*localcollection.LocalCollection, error) {
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
		if *input.Type != localcollection.CollectionTypeFolder && *input.Type != localcollection.CollectionTypeAlbum {
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
