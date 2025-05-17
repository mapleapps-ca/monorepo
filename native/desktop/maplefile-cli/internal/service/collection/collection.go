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
	CreateSubCollection(ctx context.Context, input CreateSubCollectionInput) (*CreateCollectionOutput, error)
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
		s.logger.Error("collection name is required for root collection", zap.String("name", input.Name))
		return nil, errors.NewAppError("collection name is required", nil)
	}

	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = collection.CollectionTypeFolder
	} else if input.CollectionType != collection.CollectionTypeFolder && input.CollectionType != collection.CollectionTypeAlbum {
		s.logger.Error("invalid collection type for root collection", zap.String("type", input.CollectionType))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	// Get the authenticated user's email
	email, err := s.configService.GetEmail(ctx)
	if err != nil {
		s.logger.Error("failed to get current user email for root collection creation", zap.Error(err))
		return nil, errors.NewAppError("failed to get current user", err)
	}

	// Get user data
	userData, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get user data for root collection creation", zap.String("email", email), zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("user not found for root collection creation", zap.String("email", email))
		return nil, errors.NewAppError("user not found; please login first", nil)
	}

	// Generate a collection key
	collectionKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		s.logger.Error("failed to generate collection key for root collection", zap.Error(err))
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
		s.logger.Error("failed to encrypt collection key for root collection", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	encryptedCollectionKey := keys.EncryptedCollectionKey{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}

	// Call the use case to create the collection
	response, err := s.useCase.Execute(ctx, encryptedName, input.CollectionType, encryptedCollectionKey)
	if err != nil {
		s.logger.Error("use case failed to create root collection", zap.String("name", input.Name), zap.Error(err))
		return nil, err
	}

	return &CreateCollectionOutput{
		Collection: response,
	}, nil
}

// CreateSubCollectionInput represents the input for creating a sub-collection
type CreateSubCollectionInput struct {
	Name           string `json:"name"`
	CollectionType string `json:"collection_type"`
	ParentID       string `json:"parent_id"`
}

// CreateSubCollection handles the creation of a sub-collection under a parent collection
func (s *collectionService) CreateSubCollection(ctx context.Context, input CreateSubCollectionInput) (*CreateCollectionOutput, error) {
	// Validate inputs
	if input.Name == "" {
		s.logger.Error("collection name is required for sub-collection", zap.String("name", input.Name), zap.String("parent_id", input.ParentID))
		return nil, errors.NewAppError("collection name is required", nil)
	}

	if input.ParentID == "" {
		s.logger.Error("parent collection ID is required for sub-collection", zap.String("name", input.Name), zap.String("parent_id", input.ParentID))
		return nil, errors.NewAppError("parent collection ID is required", nil)
	}

	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = collection.CollectionTypeFolder
	} else if input.CollectionType != collection.CollectionTypeFolder && input.CollectionType != collection.CollectionTypeAlbum {
		s.logger.Error("invalid collection type for sub-collection", zap.String("type", input.CollectionType), zap.String("name", input.Name), zap.String("parent_id", input.ParentID))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	// Get the authenticated user's email
	email, err := s.configService.GetEmail(ctx)
	if err != nil {
		s.logger.Error("failed to get current user email for sub-collection creation", zap.String("name", input.Name), zap.String("parent_id", input.ParentID), zap.Error(err))
		return nil, errors.NewAppError("failed to get current user", err)
	}

	// Get user data
	userData, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get user data for sub-collection creation", zap.String("email", email), zap.String("name", input.Name), zap.String("parent_id", input.ParentID), zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("user not found for sub-collection creation", zap.String("email", email), zap.String("name", input.Name), zap.String("parent_id", input.ParentID))
		return nil, errors.NewAppError("user not found; please login first", nil)
	}

	// Generate a collection key
	collectionKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		s.logger.Error("failed to generate collection key for sub-collection", zap.String("name", input.Name), zap.String("parent_id", input.ParentID), zap.Error(err))
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
		s.logger.Error("failed to encrypt collection key for sub-collection", zap.String("name", input.Name), zap.String("parent_id", input.ParentID), zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	encryptedCollectionKey := keys.EncryptedCollectionKey{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}

	// For a sub-collection, we need to create a path segment representing the name
	// In a real implementation, this would be encrypted with the parent collection's key
	encryptedPathSegments := []string{encryptedName}

	// Call the use case to create the sub-collection
	response, err := s.useCase.ExecuteSubCollection(
		ctx,
		encryptedName,
		input.CollectionType,
		input.ParentID,
		encryptedPathSegments,
		encryptedCollectionKey,
	)
	if err != nil {
		s.logger.Error("use case failed to create sub-collection",
			zap.String("name", input.Name),
			zap.String("parent_id", input.ParentID),
			zap.Error(err))
		return nil, err
	}

	return &CreateCollectionOutput{
		Collection: response,
	}, nil
}
