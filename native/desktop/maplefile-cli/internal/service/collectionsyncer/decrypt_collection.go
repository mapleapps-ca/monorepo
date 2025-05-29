package collectionsyncer

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CollectionDecryptionService handles decryption of collection data
type CollectionDecryptionService interface {
	DecryptCollectionName(ctx context.Context, collectionDTO *collectiondto.CollectionDTO, userPassword string) (string, error)
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

// DecryptCollectionName decrypts a collection's encrypted name
func (s *collectionDecryptionService) DecryptCollectionName(ctx context.Context, collectionDTO *collectiondto.CollectionDTO, userPassword string) (string, error) {
	// Validate inputs
	if collectionDTO == nil {
		return "", errors.NewAppError("collection DTO is required", nil)
	}
	if userPassword == "" {
		return "", errors.NewAppError("user password is required for decryption", nil)
	}
	if collectionDTO.EncryptedName == "" {
		return "", errors.NewAppError("encrypted name is required", nil)
	}
	if collectionDTO.EncryptedCollectionKey == nil {
		return "", errors.NewAppError("encrypted collection key is required", nil)
	}

	// Get user data for decryption keys
	userData, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get authenticated user", zap.Error(err))
		return "", errors.NewAppError("failed to get user data", err)
	}
	if userData == nil {
		return "", errors.NewAppError("authenticated user not found; please login first", nil)
	}

	// Step 1: Derive keyEncryptionKey from password + salt (E2EE spec)
	keyEncryptionKey, err := s.deriveKeyEncryptionKey(userPassword, userData.PasswordSalt)
	if err != nil {
		return "", errors.NewAppError("failed to derive key encryption key", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	// Step 2: Decrypt masterKey with keyEncryptionKey (E2EE spec)
	masterKey, err := s.decryptMasterKey(userData, keyEncryptionKey)
	if err != nil {
		return "", errors.NewAppError("failed to decrypt master key - incorrect password?", err)
	}
	defer crypto.ClearBytes(masterKey)

	// Step 3: Decrypt collectionKey with masterKey (E2EE spec)
	collectionKey, err := crypto.DecryptWithSecretBox(
		collectionDTO.EncryptedCollectionKey.Ciphertext,
		collectionDTO.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return "", errors.NewAppError("failed to decrypt collection key", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Step 4: Decrypt collection name with collectionKey (E2EE spec)
	decryptedName, err := s.decryptCollectionName(collectionDTO.EncryptedName, collectionKey)
	if err != nil {
		return "", errors.NewAppError("failed to decrypt collection name", err)
	}

	s.logger.Debug("✅ Successfully decrypted collection name",
		zap.String("collectionID", collectionDTO.ID.Hex()),
		zap.String("decryptedName", decryptedName))

	return decryptedName, nil
}

// Helper: Derive keyEncryptionKey from password (E2EE spec)
func (s *collectionDecryptionService) deriveKeyEncryptionKey(password string, salt []byte) ([]byte, error) {
	return crypto.DeriveKeyFromPassword(password, salt)
}

// Helper: Decrypt masterKey with keyEncryptionKey (E2EE spec)
func (s *collectionDecryptionService) decryptMasterKey(user *dom_user.User, keyEncryptionKey []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
}

// Helper: Decrypt collection name with collectionKey (E2EE spec)
func (s *collectionDecryptionService) decryptCollectionName(encryptedName string, collectionKey []byte) (string, error) {
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
