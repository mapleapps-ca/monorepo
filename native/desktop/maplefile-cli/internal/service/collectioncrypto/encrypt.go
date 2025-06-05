package collectioncrypto

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_keys "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CollectionEncryptionService handles Encryption of collection data
type CollectionEncryptionService interface {
	ExecuteForCreateCollectionKeyAndEncryptWithMasterKey(ctx context.Context, user *dom_user.User, password string) (*dom_keys.EncryptedCollectionKey, []byte, error)
	ExecuteForEncryptData(ctx context.Context, decryptedData string, collectionKey []byte) (string, error)
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

func (s *collectionEncryptionService) ExecuteForCreateCollectionKeyAndEncryptWithMasterKey(ctx context.Context, user *dom_user.User, password string) (*dom_keys.EncryptedCollectionKey, []byte, error) {
	s.logger.Debug("🔑 Starting E2EE key chain Encryption",
		zap.String("userID", user.ID.Hex()),
	)

	//
	// STEP 1: Derive keyEncryptionKey from password
	//

	s.logger.Debug("🧠 Step 1: Deriving key encryption key from password")
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.logger.Error("❌ Failed to derive key encryption key", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)
	s.logger.Debug("✅ Successfully derived key encryption key")

	//
	// STEP 2: Decrypt masterKey with keyEncryptionKey
	//

	s.logger.Debug("🧠 Step 2: Decrypt masterKey with keyEncryptionKey")
	masterKey, err := s.decryptMasterKey(user, keyEncryptionKey)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to decrypt master key - incorrect password?", err)
	}
	defer crypto.ClearBytes(masterKey)
	s.logger.Debug("✅ Successfully decrypted masterkey")

	//
	// STEP 3: Generate random collectionKey
	//

	s.logger.Debug("🧠 Step 3: Generate random collectionKey")
	collectionKey, err := crypto.GenerateRandomBytes(crypto.CollectionKeySize)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to generate collection key", err)
	}
	// Developer Note: Do not run `	defer crypto.ClearBytes(collectionKey)` b/c it'll clear our return! It's now up to the developer to make sure they call this in the app or else risk exposing this.
	s.logger.Debug("✅ Successfully generated random collectionKey")

	//
	// STEP 4: Encrypt collectionKey with masterKey
	//

	s.logger.Debug("🧠 Step 4: Encrypt collectionKey with masterKey")
	encryptedData, err := crypto.EncryptWithSecretBox(collectionKey, masterKey)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to encrypt collection key", err)
	}
	s.logger.Debug("✅ Successfully encrypted collection key with masterKey")

	//
	// STEP 5: Create our structured key
	//

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

	return encryptedCollectionKey, collectionKey, nil
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
