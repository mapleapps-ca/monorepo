// internal/service/collection/create.go
package collection

import (
	"context"
	"encoding/base64"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CreateInput represents the input for creating a local collection
type CreateInput struct {
	Name           string             `json:"name"`
	CollectionType string             `json:"collection_type"`
	ParentID       string             `json:"parent_id,omitempty"`
	OwnerID        primitive.ObjectID `bson:"owner_id" json:"owner_id"`
}

// CreateOutput represents the result of creating a local collection
type CreateOutput struct {
	Collection *dom_collection.Collection `json:"collection"`
}

// CreateService defines the interface for creating local collections
type CreateService interface {
	Create(ctx context.Context, input *CreateInput) (*CreateOutput, error)
}

// createService implements the CreateService interface
type createService struct {
	logger                         *zap.Logger
	configService                  config.ConfigService
	transactionManager             dom_tx.Manager
	getUserByIsLoggedInUseCase     uc_user.GetByIsLoggedInUseCase
	createCollectionInCloudUseCase uc_collectiondto.CreateCollectionInCloudUseCase
	getUserByEmailUseCase          uc_user.GetByEmailUseCase
	createCollectionUseCase        uc_collection.CreateCollectionUseCase
}

// NewCreateService creates a new local collection creation service
func NewCreateService(
	logger *zap.Logger,
	configService config.ConfigService,
	transactionManager dom_tx.Manager,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	createCollectionInCloudUseCase uc_collectiondto.CreateCollectionInCloudUseCase,
	getUserByEmailUseCase uc_user.GetByEmailUseCase,
	createCollectionUseCase uc_collection.CreateCollectionUseCase,
) CreateService {
	return &createService{
		logger:                         logger,
		configService:                  configService,
		transactionManager:             transactionManager,
		getUserByIsLoggedInUseCase:     getUserByIsLoggedInUseCase,
		createCollectionInCloudUseCase: createCollectionInCloudUseCase,
		getUserByEmailUseCase:          getUserByEmailUseCase,
		createCollectionUseCase:        createCollectionUseCase,
	}
}

// Create handles the creation of a local collection
func (s *createService) Create(ctx context.Context, input *CreateInput) (*CreateOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	// Validate inputs
	if input == nil {
		s.logger.Error("input is required", zap.Any("input", input))
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.Name == "" {
		s.logger.Error("collection name is required", zap.Any("input", input))
		return nil, errors.NewAppError("collection name is required", nil)
	}
	if input.OwnerID.IsZero() {
		s.logger.Error("owner ID is required", zap.Any("input", input))
		return nil, errors.NewAppError("owner ID is required", nil)
	}
	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = dom_collection.CollectionTypeFolder
	} else if input.CollectionType != dom_collection.CollectionTypeFolder && input.CollectionType != dom_collection.CollectionTypeAlbum {
		s.logger.Error("invalid collection type", zap.String("type", input.CollectionType))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}

	//
	// STEP 2: Get related records or error.
	//

	// Get user data
	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("failed to get authenticated user", zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("authenticated user not found")
		return nil, errors.NewAppError("authenticated user not found; please login first", nil)
	}

	//
	// STEP 3: Begin transaction
	//

	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("failed to begin transaction", zap.Error(err))
		return nil, errors.NewAppError("failed to begin transaction", err)
	}

	//
	// STEP 4: Create encryption key and encrypted the data with it.
	//

	// Generate a collection key
	collectionKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		s.logger.Error("failed to generate collection key", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to generate collection key", err)
	}

	// Encrypt the collection name (using the collection key)
	// Note: In a real implementation, this would use more complex encryption
	// TODO: Implement real encryption.
	nameBytes := []byte(input.Name)
	encryptedName := base64.StdEncoding.EncodeToString(nameBytes)

	// Encrypt the collection key with the user's master key
	// This is a simplified version - real implementation would decrypt master key first
	ciphertext, nonce, err := crypto.EncryptWithSecretBox(collectionKey, collectionKey)
	if err != nil {
		s.logger.Error("failed to encrypt collection key", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	currentTime := time.Now() // Capture the current time once
	historicalKey := keys.EncryptedHistoricalKey{
		KeyVersion:    1, // Always start at version 1.
		Ciphertext:    ciphertext,
		Nonce:         nonce,
		RotatedAt:     currentTime,
		RotatedReason: "Initial collection creation",
		Algorithm:     "chacha20poly1305", //TODO: Confirm this is the algorithm used.
	}

	encryptedCollectionKey := keys.EncryptedCollectionKey{
		Ciphertext:   ciphertext,
		Nonce:        nonce,
		KeyVersion:   1,            // Always start at version 1.
		RotatedAt:    &currentTime, // Pass the address of the captured time
		PreviousKeys: []keys.EncryptedHistoricalKey{historicalKey},
	}

	//
	// STEP 5: Create our collection data transfer object and submit to the cloud to return the "Cloud ID" of this collection to store locally.
	//

	// DEVELOPER NOTE:
	// (1) In future expand sharing and other features here.

	collectionDTO := &dom_collectiondto.CollectionDTO{
		// ID:          primitive.NewObjectID(), // Do not set ID here, it will be set by the cloud service
		OwnerID:                userData.ID,
		EncryptedName:          encryptedName,
		CollectionType:         input.CollectionType,
		EncryptedCollectionKey: &encryptedCollectionKey,
		Members:                make([]*dom_collectiondto.CollectionMembershipDTO, 0), // (1)
		ParentID:               primitive.NilObjectID,
		Children:               make([]*dom_collectiondto.CollectionDTO, 0),
		CreatedAt:              time.Now(),
		CreatedByUserID:        userData.ID,
		ModifiedAt:             time.Now(),
		ModifiedByUserID:       userData.ID,
		Version:                1, // Always start at version 1, on every edit increment this value by one
	}

	collectionCloudID, err := s.createCollectionInCloudUseCase.Execute(ctx, collectionDTO)
	if err != nil {
		s.logger.Error("failed to create collection in the cloud", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to create collection in the cloud", err)
	}

	//
	// STEP 6: Create collection record in our local database.
	//

	// DEVELOPER NOTE: (1)
	// The cloud has assigned our collection ID and hence we must save it as the unique identifier.
	// It is important that we use what is provided by the cloud and DO NOT CREATE OUR OWN ID USING
	// THE LOCAL DEVICE.
	//
	// DEVELOPER NOTE: (2)
	// In future expand sharing and other features here.

	col := &dom_collection.Collection{
		ID:                     *collectionCloudID, // (1)
		OwnerID:                userData.ID,
		EncryptedName:          encryptedName,
		CollectionType:         input.CollectionType,
		EncryptedCollectionKey: &encryptedCollectionKey,
		Members:                make([]*dom_collection.CollectionMembership, 0), // (2)
		SyncStatus:             dom_collection.SyncStatusSynced,
		ParentID:               collectionDTO.ParentID,
		Children:               make([]*dom_collection.Collection, 0),
		CreatedAt:              collectionDTO.CreatedAt,
		CreatedByUserID:        userData.ID,
		ModifiedAt:             collectionDTO.ModifiedAt,
		ModifiedByUserID:       collectionDTO.ModifiedByUserID,
		Version:                collectionDTO.Version,
	}

	// Call the use case to create the collection
	if err := s.createCollectionUseCase.Execute(ctx, col); err != nil {
		s.logger.Error("failed to create local collection", zap.String("name", input.Name), zap.Error(err))
		s.transactionManager.Rollback()
		return nil, err
	}

	//
	// STEP X: Commit transaction and return method output.
	//
	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	return &CreateOutput{
		Collection: col,
	}, nil
}
