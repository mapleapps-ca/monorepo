// monorepo/native/desktop/maplefile-cli/internal/service/collection/get_filtered.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// GetFilteredInput represents the input for getting filtered collections
type GetFilteredInput struct {
	IncludeOwned  bool `json:"include_owned"`
	IncludeShared bool `json:"include_shared"`
}

// GetFilteredOutput represents the result of getting filtered collections
type GetFilteredOutput struct {
	OwnedCollections  []*collection.Collection `json:"owned_collections"`
	SharedCollections []*collection.Collection `json:"shared_collections"`
	TotalCount        int                      `json:"total_count"`
}

// GetFilteredService defines the interface for getting filtered collections
type GetFilteredService interface {
	GetFiltered(ctx context.Context, input *GetFilteredInput, userPassword string) (*GetFilteredOutput, error)
}

// getFilteredService implements the GetFilteredService interface
type getFilteredService struct {
	logger                                 *zap.Logger
	getUserByIsLoggedInUseCase             uc_user.GetByIsLoggedInUseCase
	getFilteredCollectionsFromCloudUseCase uc_collectiondto.GetFilteredCollectionsFromCloudUseCase
}

// NewGetFilteredService creates a new service for getting filtered collections
func NewGetFilteredService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getFilteredCollectionsFromCloudUseCase uc_collectiondto.GetFilteredCollectionsFromCloudUseCase,
) GetFilteredService {
	logger = logger.Named("GetFilteredService")
	return &getFilteredService{
		logger:                                 logger,
		getUserByIsLoggedInUseCase:             getUserByIsLoggedInUseCase,
		getFilteredCollectionsFromCloudUseCase: getFilteredCollectionsFromCloudUseCase,
	}
}

// GetFiltered retrieves filtered collections from the cloud and decrypts them
func (s *getFilteredService) GetFiltered(ctx context.Context, input *GetFilteredInput, userPassword string) (*GetFilteredOutput, error) {
	//
	// STEP 1: Validate inputs
	//

	if input == nil {
		s.logger.Error("input is required")
		return nil, errors.NewAppError("input is required", nil)
	}

	if !input.IncludeOwned && !input.IncludeShared {
		s.logger.Error("at least one filter option must be enabled")
		return nil, errors.NewAppError("at least one filter option (include_owned or include_shared) must be enabled", nil)
	}

	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	//
	// STEP 2: Get user data for decryption
	//

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
	// STEP 3: Set up decryption keys
	//

	// Derive keyEncryptionKey from password (E2EE spec)
	keyEncryptionKey, err := s.deriveKeyEncryptionKey(userPassword, userData.PasswordSalt)
	if err != nil {
		return nil, errors.NewAppError("failed to derive key encryption key", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	// Decrypt masterKey with keyEncryptionKey (E2EE spec)
	masterKey, err := s.decryptMasterKey(userData, keyEncryptionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt master key - incorrect password?", err)
	}
	defer crypto.ClearBytes(masterKey)

	//
	// STEP 4: Get filtered collections from cloud
	//

	request := &collectiondto.GetFilteredCollectionsRequest{
		IncludeOwned:  input.IncludeOwned,
		IncludeShared: input.IncludeShared,
	}

	s.logger.Debug("Getting filtered collections from cloud",
		zap.Bool("include_owned", input.IncludeOwned),
		zap.Bool("include_shared", input.IncludeShared))

	cloudResponse, err := s.getFilteredCollectionsFromCloudUseCase.Execute(ctx, request)
	if err != nil {
		s.logger.Error("failed to get filtered collections from cloud", zap.Error(err))
		return nil, errors.NewAppError("failed to get filtered collections from cloud", err)
	}

	//
	// STEP 5: Decrypt collections and convert to local format
	//

	output := &GetFilteredOutput{
		OwnedCollections:  make([]*collection.Collection, 0, len(cloudResponse.OwnedCollections)),
		SharedCollections: make([]*collection.Collection, 0, len(cloudResponse.SharedCollections)),
		TotalCount:        cloudResponse.TotalCount,
	}

	// Decrypt owned collections
	for _, cloudCollection := range cloudResponse.OwnedCollections {
		localCollection, err := s.convertAndDecryptCollection(cloudCollection, masterKey)
		if err != nil {
			s.logger.Warn("failed to decrypt owned collection, skipping",
				zap.String("collection_id", cloudCollection.ID.Hex()),
				zap.Error(err))
			continue
		}
		output.OwnedCollections = append(output.OwnedCollections, localCollection)
	}

	// Decrypt shared collections
	for _, cloudCollection := range cloudResponse.SharedCollections {
		localCollection, err := s.convertAndDecryptCollection(cloudCollection, masterKey)
		if err != nil {
			s.logger.Warn("failed to decrypt shared collection, skipping",
				zap.String("collection_id", cloudCollection.ID.Hex()),
				zap.Error(err))
			continue
		}
		output.SharedCollections = append(output.SharedCollections, localCollection)
	}

	s.logger.Info("Successfully retrieved and decrypted filtered collections",
		zap.Int("owned_count", len(output.OwnedCollections)),
		zap.Int("shared_count", len(output.SharedCollections)),
		zap.Int("total_count", output.TotalCount))

	return output, nil
}

// convertAndDecryptCollection converts a cloud CollectionDTO to local Collection and decrypts it
func (s *getFilteredService) convertAndDecryptCollection(cloudCollection *collectiondto.CollectionDTO, masterKey []byte) (*collection.Collection, error) {
	// Decrypt collection key
	if cloudCollection.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	collectionKey, err := crypto.DecryptWithSecretBox(
		cloudCollection.EncryptedCollectionKey.Ciphertext,
		cloudCollection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Decrypt collection name
	decryptedName, err := s.decryptCollectionName(cloudCollection.EncryptedName, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection name", err)
	}

	// Convert to local collection format
	localCollection := &collection.Collection{
		ID:                     cloudCollection.ID,
		OwnerID:                cloudCollection.OwnerID,
		EncryptedName:          cloudCollection.EncryptedName,
		CollectionType:         cloudCollection.CollectionType,
		EncryptedCollectionKey: cloudCollection.EncryptedCollectionKey,
		ParentID:               cloudCollection.ParentID,
		AncestorIDs:            cloudCollection.AncestorIDs,
		Children:               make([]*collection.Collection, 0),
		CreatedAt:              cloudCollection.CreatedAt,
		CreatedByUserID:        cloudCollection.CreatedByUserID,
		ModifiedAt:             cloudCollection.ModifiedAt,
		ModifiedByUserID:       cloudCollection.ModifiedByUserID,
		Version:                cloudCollection.Version,
		// Decrypted fields
		Name:       decryptedName,
		SyncStatus: collection.SyncStatusSynced, // Assume synced since it came from cloud
	}

	// Convert members if any
	if cloudCollection.Members != nil {
		localCollection.Members = make([]*collection.CollectionMembership, len(cloudCollection.Members))
		for i, cloudMember := range cloudCollection.Members {
			localCollection.Members[i] = &collection.CollectionMembership{
				ID:                     cloudMember.ID,
				CollectionID:           cloudMember.CollectionID,
				RecipientID:            cloudMember.RecipientID,
				RecipientEmail:         cloudMember.RecipientEmail,
				GrantedByID:            cloudMember.GrantedByID,
				EncryptedCollectionKey: cloudMember.EncryptedCollectionKey,
				PermissionLevel:        cloudMember.PermissionLevel,
				CreatedAt:              cloudMember.CreatedAt,
				IsInherited:            cloudMember.IsInherited,
				InheritedFromID:        cloudMember.InheritedFromID,
			}
		}
	} else {
		localCollection.Members = make([]*collection.CollectionMembership, 0)
	}

	return localCollection, nil
}

// Helper: Derive keyEncryptionKey from password (E2EE spec)
func (s *getFilteredService) deriveKeyEncryptionKey(password string, salt []byte) ([]byte, error) {
	return crypto.DeriveKeyFromPassword(password, salt)
}

// Helper: Decrypt masterKey with keyEncryptionKey (E2EE spec)
func (s *getFilteredService) decryptMasterKey(user *dom_user.User, keyEncryptionKey []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
}

// Helper: Decrypt collection name with collectionKey (E2EE spec)
func (s *getFilteredService) decryptCollectionName(encryptedName string, collectionKey []byte) (string, error) {
	// Decode from base64
	combined, err := crypto.DecodeFromBase64(encryptedName)
	if err != nil {
		return "", err
	}

	// Split nonce and ciphertext
	nonce, ciphertext, err := crypto.SplitNonceAndCiphertext(combined, crypto.SecretBoxNonceSize)
	if err != nil {
		return "", err
	}

	// Decrypt
	decryptedBytes, err := crypto.DecryptWithSecretBox(ciphertext, nonce, collectionKey)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}
