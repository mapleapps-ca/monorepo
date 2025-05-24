// internal/usecase/remotecollection/create.go
package remotecollection

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
)

// CreateRemoteCollectionInput defines the input for creating a cloud collection
type CreateRemoteCollectionInput struct {
	EncryptedName          string
	Type                   string
	ParentID               *primitive.ObjectID
	EncryptedCollectionKey keys.EncryptedCollectionKey
}

// CreateRemoteCollectionUseCase defines the interface for creating a cloud collection
type CreateRemoteCollectionUseCase interface {
	Execute(ctx context.Context, input CreateRemoteCollectionInput) (*remotecollection.RemoteCollectionResponse, error)
}

// createRemoteCollectionUseCase implements the CreateRemoteCollectionUseCase interface
type createRemoteCollectionUseCase struct {
	logger     *zap.Logger
	repository remotecollection.RemoteCollectionRepository
}

// NewCreateRemoteCollectionUseCase creates a new use case for creating cloud collections
func NewCreateRemoteCollectionUseCase(
	logger *zap.Logger,
	repository remotecollection.RemoteCollectionRepository,
) CreateRemoteCollectionUseCase {
	return &createRemoteCollectionUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new cloud collection
func (uc *createRemoteCollectionUseCase) Execute(
	ctx context.Context,
	input CreateRemoteCollectionInput,
) (*remotecollection.RemoteCollectionResponse, error) {
	// Validate inputs
	if input.EncryptedName == "" {
		return nil, errors.NewAppError("encrypted name is required", nil)
	}

	if input.Type != remotecollection.CollectionTypeFolder && input.Type != remotecollection.CollectionTypeAlbum {
		return nil, errors.NewAppError(fmt.Sprintf("invalid collection type: %s (must be '%s' or '%s')",
			input.Type, remotecollection.CollectionTypeFolder, remotecollection.CollectionTypeAlbum), nil)
	}

	if input.EncryptedCollectionKey.Ciphertext == nil || len(input.EncryptedCollectionKey.Ciphertext) == 0 ||
		input.EncryptedCollectionKey.Nonce == nil || len(input.EncryptedCollectionKey.Nonce) == 0 {
		return nil, errors.NewAppError("encrypted collection key is required", nil)
	}

	// Create the request
	request := &remotecollection.RemoteCreateCollectionRequest{
		EncryptedName:          input.EncryptedName,
		Type:                   input.Type,
		EncryptedCollectionKey: input.EncryptedCollectionKey,
	}

	// Add parent ID if specified
	if input.ParentID != nil && !input.ParentID.IsZero() {
		request.ParentID = *input.ParentID
	}

	// Add path segments if specified
	if len(input.EncryptedPathSegments) > 0 {
		request.EncryptedPathSegments = input.EncryptedPathSegments
	}

	// Call the repository to create the collection
	response, err := uc.repository.Create(ctx, request)
	if err != nil {
		return nil, errors.NewAppError("failed to create cloud collection", err)
	}

	return response, nil
}
