// internal/usecase/collection/create.go
package collection

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// CreateCollectionUseCase defines the interface for creating a collection
type CreateCollectionUseCase interface {
	Execute(ctx context.Context, encryptedName, collectionType string, collectionKey keys.EncryptedCollectionKey) (*collection.CollectionResponse, error)
	ExecuteSubCollection(
		ctx context.Context,
		encryptedName string,
		collectionType string,
		parentID string,
		encryptedPathSegments []string,
		collectionKey keys.EncryptedCollectionKey,
	) (*collection.CollectionResponse, error)
}

// createCollectionUseCase implements the CreateCollectionUseCase interface
type createCollectionUseCase struct {
	logger     *zap.Logger
	repository collection.CollectionRepository
}

// NewCreateCollectionUseCase creates a new use case for creating collections
func NewCreateCollectionUseCase(
	logger *zap.Logger,
	repository collection.CollectionRepository,
) CreateCollectionUseCase {
	return &createCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new root collection
func (uc *createCollectionUseCase) Execute(
	ctx context.Context,
	encryptedName string,
	collectionType string,
	collectionKey keys.EncryptedCollectionKey,
) (*collection.CollectionResponse, error) {
	// Validate inputs
	if encryptedName == "" {
		return nil, errors.NewAppError("encrypted name is required", nil)
	}

	if collectionType != collection.CollectionTypeFolder && collectionType != collection.CollectionTypeAlbum {
		return nil, errors.NewAppError(fmt.Sprintf("invalid collection type: %s (must be '%s' or '%s')",
			collectionType, collection.CollectionTypeFolder, collection.CollectionTypeAlbum), nil)
	}

	if collectionKey.Ciphertext == nil || len(collectionKey.Ciphertext) == 0 ||
		collectionKey.Nonce == nil || len(collectionKey.Nonce) == 0 {
		return nil, errors.NewAppError("encrypted collection key is required", nil)
	}

	// Create the collection request
	request := &collection.CreateCollectionRequest{
		EncryptedName:          encryptedName,
		Type:                   collectionType,
		EncryptedCollectionKey: collectionKey,
	}

	// Call the repository to create the collection
	return uc.repository.CreateCollection(ctx, request)
}

// ExecuteSubCollection creates a new sub-collection under a parent collection
func (uc *createCollectionUseCase) ExecuteSubCollection(
	ctx context.Context,
	encryptedName string,
	collectionType string,
	parentID string,
	encryptedPathSegments []string,
	collectionKey keys.EncryptedCollectionKey,
) (*collection.CollectionResponse, error) {
	// Validate inputs
	if encryptedName == "" {
		return nil, errors.NewAppError("encrypted name is required", nil)
	}

	if collectionType != collection.CollectionTypeFolder && collectionType != collection.CollectionTypeAlbum {
		return nil, errors.NewAppError(fmt.Sprintf("invalid collection type: %s (must be '%s' or '%s')",
			collectionType, collection.CollectionTypeFolder, collection.CollectionTypeAlbum), nil)
	}

	if parentID == "" {
		return nil, errors.NewAppError("parent collection ID is required for sub-collections", nil)
	}

	if collectionKey.Ciphertext == nil || len(collectionKey.Ciphertext) == 0 ||
		collectionKey.Nonce == nil || len(collectionKey.Nonce) == 0 {
		return nil, errors.NewAppError("encrypted collection key is required", nil)
	}

	// Convert parent ID string to ObjectID
	parentObjectID, err := primitive.ObjectIDFromHex(parentID)
	if err != nil {
		return nil, errors.NewAppError(fmt.Sprintf("invalid parent ID format: %s", parentID), err)
	}

	// Create the collection request
	request := &collection.CreateCollectionRequest{
		EncryptedName:          encryptedName,
		Type:                   collectionType,
		ParentID:               parentObjectID,
		EncryptedPathSegments:  encryptedPathSegments,
		EncryptedCollectionKey: collectionKey,
	}

	// Call the repository to create the collection
	return uc.repository.CreateCollection(ctx, request)
}
