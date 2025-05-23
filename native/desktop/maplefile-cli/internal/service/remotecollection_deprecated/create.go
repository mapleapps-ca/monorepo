// internal/service/remotecollection/create.go
package remotecollection

import (
	"context"
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotecollection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CreateInput represents the input for creating a cloud collection
type CreateInput struct {
	Name           string `json:"name"`
	CollectionType string `json:"collection_type"`
	ParentID       string `json:"parent_id,omitempty"`
}

// CreateOutput represents the result of creating a cloud collection
type CreateOutput struct {
	Collection *remotecollection.RemoteCollectionResponse `json:"collection"`
}

// CreateService defines the interface for creating cloud collections
type CreateService interface {
	Create(ctx context.Context, input CreateInput) (*CreateOutput, error)
}

// createService implements the CreateService interface
type createService struct {
	logger        *zap.Logger
	configService config.ConfigService
	userRepo      user.Repository
	createUseCase uc.CreateRemoteCollectionUseCase
}

// NewCreateService creates a new service for creating cloud collections
func NewCreateService(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	createUseCase uc.CreateRemoteCollectionUseCase,
) CreateService {
	return &createService{
		logger:        logger,
		configService: configService,
		userRepo:      userRepo,
		createUseCase: createUseCase,
	}
}

// Create creates a new cloud collection
func (s *createService) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	// Validate inputs
	if input.Name == "" {
		s.logger.Error("collection name is required", zap.Any("input", input))
		return nil, errors.NewAppError("collection name is required", nil)
	}

	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = remotecollection.CollectionTypeFolder
	} else if input.CollectionType != remotecollection.CollectionTypeFolder && input.CollectionType != remotecollection.CollectionTypeAlbum {
		s.logger.Error("invalid collection type", zap.String("type", input.CollectionType))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	// Get the authenticated user's email
	email, err := s.configService.GetEmail(ctx)
	if err != nil {
		s.logger.Error("failed to get current user email", zap.Error(err))
		return nil, errors.NewAppError("failed to get current user", err)
	}

	// Get user data
	userData, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("failed to get user data", zap.String("email", email), zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("user not found", zap.String("email", email))
		return nil, errors.NewAppError("user not found; please login first", nil)
	}

	// Generate a collection key
	collectionKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		s.logger.Error("failed to generate collection key", zap.Error(err))
		return nil, errors.NewAppError("failed to generate collection key", err)
	}

	// Encrypt the collection name (using the collection key)
	nameBytes := []byte(input.Name)
	encryptedName := base64.StdEncoding.EncodeToString(nameBytes)

	// Encrypt the collection key with the user's master key
	ciphertext, nonce, err := crypto.EncryptWithSecretBox(collectionKey, collectionKey)
	if err != nil {
		s.logger.Error("failed to encrypt collection key", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	encryptedCollectionKey := keys.EncryptedCollectionKey{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}

	// Prepare the use case input
	useCaseInput := uc.CreateRemoteCollectionInput{
		EncryptedName:          encryptedName,
		Type:                   input.CollectionType,
		EncryptedCollectionKey: encryptedCollectionKey,
	}

	// If parent ID is provided, convert it to ObjectID
	if input.ParentID != "" {
		parentObjectID, err := primitive.ObjectIDFromHex(input.ParentID)
		if err != nil {
			s.logger.Error("invalid parent ID format", zap.String("parentID", input.ParentID), zap.Error(err))
			return nil, errors.NewAppError("invalid parent ID format", err)
		}
		useCaseInput.ParentID = &parentObjectID
	}

	// For sub-collections, create a path segment
	if input.ParentID != "" {
		useCaseInput.EncryptedPathSegments = []string{encryptedName}
	}

	// Call the use case to create the collection
	response, err := s.createUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		s.logger.Error("failed to create cloud collection", zap.String("name", input.Name), zap.Error(err))
		return nil, err
	}

	return &CreateOutput{
		Collection: response,
	}, nil
}
