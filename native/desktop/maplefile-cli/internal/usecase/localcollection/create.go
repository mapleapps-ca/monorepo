// internal/usecase/localcollection/create.go
package localcollection

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

// CreateLocalCollectionInput defines the input for creating a local collection
type CreateLocalCollectionInput struct {
	EncryptedName          string
	DecryptedName          string // For display purposes
	Type                   string
	ParentID               *primitive.ObjectID
	EncryptedPathSegments  []string
	EncryptedCollectionKey keys.EncryptedCollectionKey
}

// CreateLocalCollectionUseCase defines the interface for creating a local collection
type CreateLocalCollectionUseCase interface {
	Execute(ctx context.Context, input CreateLocalCollectionInput) (*localcollection.LocalCollection, error)
}

// createLocalCollectionUseCase implements the CreateLocalCollectionUseCase interface
type createLocalCollectionUseCase struct {
	logger     *zap.Logger
	repository localcollection.LocalCollectionRepository
}

// NewCreateLocalCollectionUseCase creates a new use case for creating local collections
func NewCreateLocalCollectionUseCase(
	logger *zap.Logger,
	repository localcollection.LocalCollectionRepository,
) CreateLocalCollectionUseCase {
	return &createLocalCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new local collection
func (uc *createLocalCollectionUseCase) Execute(
	ctx context.Context,
	input CreateLocalCollectionInput,
) (*localcollection.LocalCollection, error) {
	// Validate inputs
	if input.EncryptedName == "" {
		return nil, errors.NewAppError("encrypted name is required", nil)
	}

	if input.Type != localcollection.CollectionTypeFolder && input.Type != localcollection.CollectionTypeAlbum {
		return nil, errors.NewAppError(fmt.Sprintf("invalid collection type: %s (must be '%s' or '%s')",
			input.Type, localcollection.CollectionTypeFolder, localcollection.CollectionTypeAlbum), nil)
	}

	if input.EncryptedCollectionKey.Ciphertext == nil || len(input.EncryptedCollectionKey.Ciphertext) == 0 ||
		input.EncryptedCollectionKey.Nonce == nil || len(input.EncryptedCollectionKey.Nonce) == 0 {
		return nil, errors.NewAppError("encrypted collection key is required", nil)
	}

	// Create a new local collection
	collection := &localcollection.LocalCollection{
		ID:                     primitive.NewObjectID(),
		EncryptedName:          input.EncryptedName,
		DecryptedName:          input.DecryptedName,
		Type:                   input.Type,
		CreatedAt:              time.Now(),
		ModifiedAt:             time.Now(),
		EncryptedCollectionKey: input.EncryptedCollectionKey,
		IsModifiedLocally:      true,
	}

	// Set parent ID if provided
	if input.ParentID != nil {
		collection.ParentID = *input.ParentID
	}

	// Set encrypted path segments if provided
	if len(input.EncryptedPathSegments) > 0 {
		collection.EncryptedPathSegments = input.EncryptedPathSegments
	}

	// Save the collection
	err := uc.repository.Create(ctx, collection)
	if err != nil {
		return nil, errors.NewAppError("failed to create local collection", err)
	}

	return collection, nil
}
