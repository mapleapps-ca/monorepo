// internal/service/collection/collection.go
package collection

import (
	"context"
	"encoding/base64"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	collectionUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CreateCollectionInput represents the input for creating a collection
type CreateCollectionInput struct {
	Name           string `json:"name"`
	CollectionType string `json:"collection_type"`
}

// CreateCollectionOutput represents the result of creating a collection
type CreateCollectionOutput struct {
	Collection *collection.CollectionResponse `json:"collection"`
}

// CollectionService defines the interface for collection operations
type CollectionService interface {
	CreateRootCollection(ctx context.Context, input CreateCollectionInput) (*CreateCollectionOutput, error)
}

// collectionService implements the CollectionService interface
type collectionService struct {
	logger        *zap.Logger
	configService config.ConfigService
	useCase       collectionUseCase.CreateCollectionUseCase
	userRepo      user.Repository
}

// NewCollectionService creates a new collection service
func NewCollectionService(
	logger *zap.Logger,
	configService config.ConfigService,
	useCase collectionUseCase.CreateCollectionUseCase,
	userRepo user.Repository,
) CollectionService {
	return &collectionService{
		logger:        logger,
		configService: configService,
		useCase:       useCase,
		userRepo:      userRepo,
	}
}

// CreateRootCollection handles the creation of a root collection
func (s *collectionService) CreateRootCollection(ctx context.Context, input CreateCollectionInput) (*CreateCollectionOutput, error) {
	// Validate inputs
	if input.Name == "" {
		return nil, errors.NewAppError("collection name is required", nil)
	}

	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = collection.CollectionTypeFolder
	} else if input.CollectionType != collection.CollectionTypeFolder && input.CollectionType != collection.CollectionTypeAlbum {
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	// Get the authenticated user's email
	email, err := s.configService.GetEmail(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get current user", err)
	}

	// Get user data
	userData, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		return nil, errors.NewAppError("user not found; please login first", nil)
	}

	// Generate a collection key
	collectionKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		return nil, errors.NewAppError("failed to generate collection key", err)
	}

	// Encrypt the collection name (using the collection key)
	// Note: In a real implementation, this would use more complex encryption
	nameBytes := []byte(input.Name)
	encryptedName := base64.StdEncoding.EncodeToString(nameBytes)

	// Encrypt the collection key with the user's master key
	// This is a simplified version - real implementation would decrypt master key first
	ciphertext, nonce, err := crypto.EncryptWithSecretBox(collectionKey, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	encryptedCollectionKey := keys.EncryptedCollectionKey{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}

	// Call the use case to create the collection
	response, err := s.useCase.Execute(ctx, encryptedName, input.CollectionType, encryptedCollectionKey)
	if err != nil {
		return nil, err
	}

	return &CreateCollectionOutput{
		Collection: response,
	}, nil
}
