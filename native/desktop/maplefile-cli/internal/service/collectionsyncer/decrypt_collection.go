package collectionsyncer

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CollectionDecryptionService handles decryption of collection data
type CollectionDecryptionService interface {
	ExecuteDecryptCollectionKeyChain(ctx context.Context, user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error)
	ExecuteDecryptData(ctx context.Context, encryptedData string, fileKey []byte) (string, error)
}

// collectionDecryptionService implements CollectionDecryptionService
type collectionDecryptionService struct {
	logger                     *zap.Logger
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
}

// NewCollectionDecryptionService creates a new collection decryption service
func NewCollectionDecryptionService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
) CollectionDecryptionService {
	logger = logger.Named("CollectionDecryptionService")
	return &collectionDecryptionService{
		logger:                     logger,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
	}
}

func (s *collectionDecryptionService) ExecuteDecryptCollectionKeyChain(ctx context.Context, user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	s.logger.Debug("🔑 Starting E2EE key chain decryption",
		zap.String("userID", user.ID.Hex()),
		zap.String("collectionID", collection.ID.Hex()))

	// STEP 1: Derive keyEncryptionKey from password
	s.logger.Debug("🧠 Step 1: Deriving key encryption key from password")
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.logger.Error("❌ Failed to derive key encryption key", zap.Error(err))
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)
	s.logger.Debug("✅ Successfully derived key encryption key")

	// STEP 2: Decrypt masterKey with keyEncryptionKey (ChaCha20-Poly1305)
	s.logger.Debug("🧠 Step 2: Decrypting master key with key encryption key")
	if len(user.EncryptedMasterKey.Ciphertext) == 0 || len(user.EncryptedMasterKey.Nonce) == 0 {
		s.logger.Error("❌ User encrypted master key is empty or invalid",
			zap.Int("ciphertextLen", len(user.EncryptedMasterKey.Ciphertext)),
			zap.Int("nonceLen", len(user.EncryptedMasterKey.Nonce)))
		return nil, fmt.Errorf("user encrypted master key is invalid")
	}

	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt master key - this usually means incorrect password",
			zap.Error(err),
			zap.String("userID", user.ID.Hex()))
		return nil, fmt.Errorf("failed to decrypt master key - incorrect password?: %w", err)
	}
	defer crypto.ClearBytes(masterKey)
	s.logger.Debug("✅ Successfully decrypted master key")

	// STEP 3: Decrypt collectionKey with masterKey (ChaCha20-Poly1305)
	s.logger.Debug("🧠 Step 3: Decrypting collection key with master key")
	if collection.EncryptedCollectionKey == nil {
		s.logger.Error("❌ Collection has no encrypted key", zap.String("collectionID", collection.ID.Hex()))
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	if len(collection.EncryptedCollectionKey.Ciphertext) == 0 || len(collection.EncryptedCollectionKey.Nonce) == 0 {
		s.logger.Error("❌ Collection encrypted key is empty or invalid",
			zap.Int("ciphertextLen", len(collection.EncryptedCollectionKey.Ciphertext)),
			zap.Int("nonceLen", len(collection.EncryptedCollectionKey.Nonce)))
		return nil, fmt.Errorf("collection encrypted key is invalid")
	}

	collectionKey, err := crypto.DecryptWithSecretBox(
		collection.EncryptedCollectionKey.Ciphertext,
		collection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt collection key", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt collection key: %w", err)
	}
	s.logger.Debug("✅ Successfully decrypted collection key")

	return collectionKey, nil
}

func (s *collectionDecryptionService) ExecuteDecryptData(ctx context.Context, encryptedDatga string, fileKey []byte) (string, error) {
	s.logger.Debug("🔑 Decrypting collection data")

	// The encrypted metadata is stored as base64 encoded (nonce + ciphertext)
	// Format: base64(12-byte-nonce + ciphertext) for ChaCha20-Poly1305
	combined, err := base64.StdEncoding.DecodeString(encryptedDatga)
	if err != nil {
		s.logger.Error("❌ Failed to decode encrypted data from base64", zap.Error(err))
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// Split nonce and ciphertext for ChaCha20-Poly1305 (12-byte nonce)
	if len(combined) < crypto.ChaCha20Poly1305NonceSize {
		s.logger.Error("❌ Combined data too short",
			zap.Int("expectedMinSize", crypto.ChaCha20Poly1305NonceSize),
			zap.Int("actualSize", len(combined)))
		return "", fmt.Errorf("combined data too short: expected at least %d bytes for ChaCha20-Poly1305, got %d", crypto.ChaCha20Poly1305NonceSize, len(combined))
	}

	nonce := make([]byte, crypto.ChaCha20Poly1305NonceSize)
	copy(nonce, combined[:crypto.ChaCha20Poly1305NonceSize])

	ciphertext := make([]byte, len(combined)-crypto.ChaCha20Poly1305NonceSize)
	copy(ciphertext, combined[crypto.ChaCha20Poly1305NonceSize:])

	// Decrypt metadata using ChaCha20-Poly1305
	decryptedBytes, err := crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt collection data", zap.Error(err))
		return "nil", fmt.Errorf("failed to decrypt collection data: %w", err)
	}

	s.logger.Debug("✅ Successfully decrypted collection data",
		zap.String("decrypted_data", string(decryptedBytes)),
	)
	return string(decryptedBytes), nil
}
