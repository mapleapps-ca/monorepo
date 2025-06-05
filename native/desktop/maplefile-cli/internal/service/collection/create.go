// internal/service/collection/create.go
package collection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
	sprimitive "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage/mongodb"
)

// CreateInput represents the input for creating a local collection
type CreateInput struct {
	Name           string             `json:"name"`
	CollectionType string             `json:"collection_type"`
	ParentID       primitive.ObjectID `json:"parent_id,omitempty"`
	OwnerID        primitive.ObjectID `bson:"owner_id" json:"owner_id"`
}

// CreateOutput represents the result of creating a local collection
type CreateOutput struct {
	Collection *dom_collection.Collection `json:"collection"`
}

// CreateService defines the interface for creating local collections
type CreateService interface {
	Create(ctx context.Context, input *CreateInput, userPassword string) (*CreateOutput, error)
}

// createService implements the CreateService interface
type createService struct {
	logger                         *zap.Logger
	configService                  config.ConfigService
	primitiveIDObjectGenerator     sprimitive.SecurePrimitiveObjectIDGenerator
	transactionManager             dom_tx.Manager
	getUserByIsLoggedInUseCase     uc_user.GetByIsLoggedInUseCase
	collectionEncryptionService    svc_collectioncrypto.CollectionEncryptionService
	createCollectionInCloudUseCase uc_collectiondto.CreateCollectionInCloudUseCase
	getUserByEmailUseCase          uc_user.GetByEmailUseCase
	createCollectionUseCase        uc_collection.CreateCollectionUseCase
}

// NewCreateService creates a new local collection creation service
func NewCreateService(
	logger *zap.Logger,
	configService config.ConfigService,
	primitiveIDObjectGenerator sprimitive.SecurePrimitiveObjectIDGenerator,
	transactionManager dom_tx.Manager,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	collectionEncryptionService svc_collectioncrypto.CollectionEncryptionService,
	createCollectionInCloudUseCase uc_collectiondto.CreateCollectionInCloudUseCase,
	getUserByEmailUseCase uc_user.GetByEmailUseCase,
	createCollectionUseCase uc_collection.CreateCollectionUseCase,
) CreateService {
	logger = logger.Named("CreateService")
	return &createService{
		logger:                         logger,
		configService:                  configService,
		primitiveIDObjectGenerator:     primitiveIDObjectGenerator,
		transactionManager:             transactionManager,
		getUserByIsLoggedInUseCase:     getUserByIsLoggedInUseCase,
		collectionEncryptionService:    collectionEncryptionService,
		createCollectionInCloudUseCase: createCollectionInCloudUseCase,
		getUserByEmailUseCase:          getUserByEmailUseCase,
		createCollectionUseCase:        createCollectionUseCase,
	}
}

// Create handles the creation of a local collection
func (s *createService) Create(ctx context.Context, input *CreateInput, userPassword string) (*CreateOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	// Validate inputs
	if input == nil {
		s.logger.Error("❌ Input is required", zap.Any("input", input))
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.Name == "" {
		s.logger.Error("❌ Collection name is required", zap.Any("input", input))
		return nil, errors.NewAppError("collection name is required", nil)
	}
	if input.OwnerID.IsZero() {
		s.logger.Error("❌ Owner ID is required", zap.Any("input", input))
		return nil, errors.NewAppError("owner ID is required", nil)
	}
	if input.CollectionType == "" {
		// Default to folder if not specified
		input.CollectionType = dom_collection.CollectionTypeFolder
	} else if input.CollectionType != dom_collection.CollectionTypeFolder && input.CollectionType != dom_collection.CollectionTypeAlbum {
		s.logger.Error("❌ Invalid collection type", zap.String("type", input.CollectionType))
		return nil, errors.NewAppError("collection type must be either 'folder' or 'album'", nil)
	}
	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	//
	// STEP 2: Get related records or error.
	//

	// Get user data
	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get authenticated user", zap.Error(err))
		return nil, errors.NewAppError("failed to get user data", err)
	}

	if userData == nil {
		s.logger.Error("❌ Authenticated user not found")
		return nil, errors.NewAppError("authenticated user not found; please login first", nil)
	}

	//
	// STEP 3: Begin transaction
	//

	if err := s.transactionManager.Begin(); err != nil {
		s.logger.Error("❌ Failed to begin transaction", zap.Error(err))
		return nil, errors.NewAppError("failed to begin transaction", err)
	}

	//
	// STEP 4: (a) Create collection encryption key (b) encrypt it with master key and (c) encrypt any data with it.
	//

	encryptedCollectionKey, decryptedCollectionKey, err := s.collectionEncryptionService.ExecuteForCreateCollectionKeyAndEncryptWithMasterKey(ctx, userData, userPassword)
	if err != nil {
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}
	// Developer Note: This needs to be done since it is purposefully not done in the `collectionEncryptionService`.
	defer crypto.ClearBytes(decryptedCollectionKey)

	//
	// STEP 5: Encrypt collection metadata with collectionKey
	//

	encryptedName, err := s.collectionEncryptionService.ExecuteForEncryptData(ctx, input.Name, decryptedCollectionKey)
	if err != nil {
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to encrypt collection name", err)
	}

	currentTime := time.Now()
	historicalKey := keys.EncryptedHistoricalKey{
		Ciphertext:    encryptedCollectionKey.Ciphertext,
		Nonce:         encryptedCollectionKey.Nonce,
		KeyVersion:    1,
		RotatedAt:     currentTime,
		RotatedReason: "Initial collection creation",
		Algorithm:     crypto.ChaCha20Poly1305Algorithm,
	}

	//
	// STEP 5: Create our collection data transfer object and submit to the cloud
	//

	// Generate client-side a ID which is cryptographically secure, cross-platform, and
	// designed for distributed systems.
	collectionID := s.primitiveIDObjectGenerator.GenerateValidObjectID()

	// Create collection with properly encrypted data and default state
	collectionDTO := &dom_collectiondto.CollectionDTO{
		ID:             collectionID,
		OwnerID:        input.OwnerID,
		EncryptedName:  encryptedName,
		CollectionType: input.CollectionType,
		ParentID:       input.ParentID,
		Members:        make([]*dom_collectiondto.CollectionMembershipDTO, 0),
		EncryptedCollectionKey: &keys.EncryptedCollectionKey{
			Ciphertext:   encryptedCollectionKey.Ciphertext,
			Nonce:        encryptedCollectionKey.Nonce,
			KeyVersion:   1,
			RotatedAt:    &currentTime,
			PreviousKeys: []keys.EncryptedHistoricalKey{historicalKey},
		},
		Children:         make([]*dom_collectiondto.CollectionDTO, 0),
		CreatedAt:        time.Now(),
		CreatedByUserID:  input.OwnerID,
		ModifiedAt:       time.Now(),
		ModifiedByUserID: input.OwnerID,
		Version:          1,                                          // Always set `version=1` at creation of a collection
		State:            dom_collectiondto.CollectionDTOStateActive, // SET DEFAULT STATE
	}

	collectionCloudID, err := s.createCollectionInCloudUseCase.Execute(ctx, collectionDTO)
	if err != nil {
		s.logger.Error("❌ Failed to create collection in the cloud", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to create collection in the cloud", err)
	}

	//
	// STEP 6: Create collection record in our local database.
	//

	col := &dom_collection.Collection{
		ID:                     *collectionCloudID,
		OwnerID:                userData.ID,
		EncryptedName:          encryptedName,
		CollectionType:         input.CollectionType,
		Members:                make([]*dom_collection.CollectionMembership, 0),
		EncryptedCollectionKey: collectionDTO.EncryptedCollectionKey,
		ParentID:               collectionDTO.ParentID,
		Children:               make([]*dom_collection.Collection, 0),
		CreatedAt:              collectionDTO.CreatedAt,
		CreatedByUserID:        userData.ID,
		ModifiedAt:             collectionDTO.ModifiedAt,
		ModifiedByUserID:       collectionDTO.ModifiedByUserID,
		Version:                collectionDTO.Version,
		State:                  dom_collection.CollectionStateActive, // SET DEFAULT STATE
		// Decrypted fields saved here:
		Name:       input.Name, // Keep plaintext for local use
		SyncStatus: dom_collection.SyncStatusSynced,
	}

	// Call the use case to create the collection
	if err := s.createCollectionUseCase.Execute(ctx, col); err != nil {
		s.logger.Error("❌ Failed to create local collection", zap.String("name", input.Name), zap.Error(err))
		s.transactionManager.Rollback()
		return nil, err
	}

	//
	// STEP 7: Commit transaction and return method output.
	//
	if err := s.transactionManager.Commit(); err != nil {
		s.logger.Error("❌ Failed to commit transaction", zap.Error(err))
		s.transactionManager.Rollback()
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	s.logger.Info("✅ Successfully created E2EE collection",
		zap.String("collectionID", collectionCloudID.Hex()),
		zap.String("name", input.Name),
		zap.String("state", col.State))

	return &CreateOutput{
		Collection: col,
	}, nil
}
