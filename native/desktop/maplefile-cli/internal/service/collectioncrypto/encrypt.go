package collectioncrypto

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_keys "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CollectionEncryptionService handles Encryption of collection data
type CollectionEncryptionService interface {
	ExecuteForEncryptCollectionKey(ctx context.Context, user *dom_user.User, collection *dom_collection.Collection, password string) (*dom_keys.EncryptedCollectionKey, error)
	ExecuteForEncryptData(ctx context.Context, decryptedData string, fileKey []byte) (string, error)
}

// collectionEncryptionService implements CollectionEncryptionService
type collectionEncryptionService struct {
	logger                     *zap.Logger
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
}

// NewCollectionEncryptionService creates a new collection Encryption service
func NewCollectionEncryptionService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
) CollectionEncryptionService {
	logger = logger.Named("CollectionEncryptionService")
	return &collectionEncryptionService{
		logger:                     logger,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
	}
}

func (s *collectionEncryptionService) ExecuteForEncryptCollectionKey(ctx context.Context, user *dom_user.User, collection *dom_collection.Collection, password string) (*dom_keys.EncryptedCollectionKey, error) {
	s.logger.Debug("🔑 Starting E2EE key chain Encryption",
		zap.String("userID", user.ID.Hex()),
		zap.String("collectionID", collection.ID.Hex()),
		zap.String("collectionOwnerID", collection.OwnerID.Hex()))

	//
	// STEP 1: Derive keyEncryptionKey from password
	//

	s.logger.Debug("🧠 Step 1: Deriving key encryption key from password")
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.logger.Error("❌ Failed to derive key encryption key", zap.Error(err))
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)
	s.logger.Debug("✅ Successfully derived key encryption key")

	//
	// STEP 2: Decrypt masterKey with keyEncryptionKey (E2EE spec)
	//

	s.logger.Debug("🧠 Step 2: Decrypt masterKey with keyEncryptionKey (E2EE spec)")
	masterKey, err := s.decryptMasterKey(user, keyEncryptionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt master key - incorrect password?", err)
	}
	defer crypto.ClearBytes(masterKey)
	s.logger.Debug("✅ Successfully decrypted masterkey")

	//
	// STEP 3: Generate random collectionKey (E2EE spec)
	//

	s.logger.Debug("🧠 Step 3: Generate random collectionKey (E2EE spec)")
	collectionKey, err := crypto.GenerateRandomBytes(crypto.CollectionKeySize)
	if err != nil {
		return nil, errors.NewAppError("failed to generate collection key", err)
	}
	defer crypto.ClearBytes(collectionKey)
	s.logger.Debug("✅ Successfully generated random collectionKey")

	// STEP 4: Encrypt collectionKey with masterKey (E2EE spec)
	encryptedData, err := crypto.EncryptWithSecretBox(collectionKey, masterKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	// // STEP 5: Encrypt collection metadata with collectionKey (E2EE spec)
	// encryptedName, err := s.encryptCollectionName(collection.Name, collectionKey)
	// if err != nil {
	// 	return nil, errors.NewAppError("failed to encrypt collection name", err)
	// }

	currentTime := time.Now()
	historicalKey := keys.EncryptedHistoricalKey{
		Ciphertext:    encryptedData.Ciphertext,
		Nonce:         encryptedData.Nonce,
		KeyVersion:    1,
		RotatedAt:     currentTime,
		RotatedReason: "Initial collection creation",
		Algorithm:     crypto.ChaCha20Poly1305Algorithm,
	}

	encryptedCollectionKey := &keys.EncryptedCollectionKey{
		Ciphertext:   encryptedData.Ciphertext,
		Nonce:        encryptedData.Nonce,
		KeyVersion:   1,
		RotatedAt:    &currentTime,
		PreviousKeys: []keys.EncryptedHistoricalKey{historicalKey},
	}

	return encryptedCollectionKey, nil
}

func (s *collectionEncryptionService) ExecuteForEncryptData(ctx context.Context, decryptedData string, collectionKey []byte) (string, error) {
	encryptedData, err := crypto.EncryptWithSecretBox([]byte(decryptedData), collectionKey)
	if err != nil {
		return "", err
	}

	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)
	return crypto.EncodeToBase64(combined), nil
}

// Helper: Decrypt masterKey with keyEncryptionKey (E2EE spec)
func (s *collectionEncryptionService) decryptMasterKey(user *dom_user.User, keyEncryptionKey []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
}
