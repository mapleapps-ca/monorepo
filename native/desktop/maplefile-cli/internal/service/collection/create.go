// internal/service/collection/create.go
package collection

import (
	"context"
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CreateInput represents the input for creating a local collection
type CreateInput struct {
	Name           string `json:"name"`
	CollectionType string `json:"collection_type"`
	ParentID       string `json:"parent_id,omitempty"`
}

// CreateOutput represents the result of creating a local collection
type CreateOutput struct {
	Collection *dom_collection.Collection `json:"collection"`
}

// CreateService defines the interface for creating local collections
type CreateService interface {
	Create(ctx context.Context, input CreateInput) (*CreateOutput, error)
}

// createService implements the CreateService interface
type createService struct {
	logger                  *zap.Logger
	configService           config.ConfigService
	getUserByEmailUseCase   uc_user.GetByEmailUseCase
	createCollectionUseCase uc_collection.CreateCollectionUseCase
}

// NewCreateService creates a new local collection creation service
func NewCreateService(
	logger *zap.Logger,
	configService config.ConfigService,
	getUserByEmailUseCase uc_user.GetByEmailUseCase,
	createCollectionUseCase uc_collection.CreateCollectionUseCase,
) CreateService {
	return &createService{
		logger:                  logger,
		configService:           configService,
		getUserByEmailUseCase:   getUserByEmailUseCase,
		createCollectionUseCase: createCollectionUseCase,
	}
}

// Create handles the creation of a local collection
func (s *createService) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	// Validate inputs
	if input.Name == "" {
		s.logger.Error("collection name is required", zap.Any("input", input))
		return nil, errors.NewAppError("collection name is required", nil)
	}

	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = dom_collection.CollectionTypeFolder
	} else if input.CollectionType != dom_collection.CollectionTypeFolder && input.CollectionType != dom_collection.CollectionTypeAlbum {
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
	userData, err := s.getUserByEmailUseCase.Execute(ctx, email)
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
	// Note: In a real implementation, this would use more complex encryption
	nameBytes := []byte(input.Name)
	encryptedName := base64.StdEncoding.EncodeToString(nameBytes)

	// Encrypt the collection key with the user's master key
	// This is a simplified version - real implementation would decrypt master key first
	ciphertext, nonce, err := crypto.EncryptWithSecretBox(collectionKey, collectionKey)
	if err != nil {
		s.logger.Error("failed to encrypt collection key", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	encryptedCollectionKey := keys.EncryptedCollectionKey{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		//TODO: DONT FORGET TO IMPLEMENT THIS:
		// KeyVersion   int                      `json:"key_version" bson:"key_version"`
		// RotatedAt    *time.Time               `json:"rotated_at,omitempty" bson:"rotated_at,omitempty"`
		// PreviousKeys []EncryptedHistoricalKey `json:"previous_keys,omitempty" bson:"previous_keys,omitempty"`
	}

	// Prepare the use case input
	useCaseInput := uc_collection.CreateCollectionInput{
		EncryptedName:          encryptedName,
		DecryptedName:          input.Name, // Store the decrypted name for display
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

	// Call the use case to create the collection
	collection, err := s.createCollectionUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		s.logger.Error("failed to create local collection", zap.String("name", input.Name), zap.Error(err))
		return nil, err
	}

	return &CreateOutput{
		Collection: collection,
	}, nil
}
