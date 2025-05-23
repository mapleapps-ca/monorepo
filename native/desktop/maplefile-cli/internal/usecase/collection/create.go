// internal/usecase/collection/create.go
package collection

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// CreateCollectionInput defines the input for creating a local collection
type CreateCollectionInput struct {
	EncryptedName          string
	DecryptedName          string // For display purposes
	Type                   string
	ParentID               *primitive.ObjectID
	EncryptedPathSegments  []string
	EncryptedCollectionKey keys.EncryptedCollectionKey
}

// CreateCollectionUseCase defines the interface for creating a local collection
type CreateCollectionUseCase interface {
	Execute(ctx context.Context, input CreateCollectionInput) (*collection.Collection, error)
}

// createCollectionUseCase implements the CreateCollectionUseCase interface
type createCollectionUseCase struct {
	logger     *zap.Logger
	repository collection.CollectionRepository
}

// NewCreateCollectionUseCase creates a new use case for creating local collections
func NewCreateCollectionUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
) CreateCollectionUseCase {
	return &createCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new local collection
func (uc *createCollectionUseCase) Execute(
	ctx context.Context,
	input CreateCollectionInput,
) (*collection.Collection, error) {
	// Validate inputs
	if input.EncryptedName == "" {
		return nil, errors.NewAppError("encrypted name is required", nil)
	}

	if input.Type != collection.CollectionTypeFolder && input.Type != collection.CollectionTypeAlbum {
		return nil, errors.NewAppError(fmt.Sprintf("invalid collection type: %s (must be '%s' or '%s')",
			input.Type, collection.CollectionTypeFolder, collection.CollectionTypeAlbum), nil)
	}

	if input.EncryptedCollectionKey.Ciphertext == nil || len(input.EncryptedCollectionKey.Ciphertext) == 0 ||
		input.EncryptedCollectionKey.Nonce == nil || len(input.EncryptedCollectionKey.Nonce) == 0 {
		return nil, errors.NewAppError("encrypted collection key is required", nil)
	}

	// Create a new local collection
	collection := &collection.Collection{
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
