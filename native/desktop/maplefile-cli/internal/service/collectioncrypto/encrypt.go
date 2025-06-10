// internal/service/collectioncrypto/encrypt.go
package collectioncrypto

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_keys "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// Enhanced CollectionEncryptionService with sharing capabilities
type CollectionEncryptionService interface {
	// Existing methods
	ExecuteForCreateCollectionKeyAndEncryptWithMasterKey(ctx context.Context, user *dom_user.User, password string) (*dom_keys.EncryptedCollectionKey, []byte, error)
	ExecuteForEncryptData(ctx context.Context, decryptedData string, collectionKey []byte) (string, error)
	EncryptCollectionKeyForSharing(ctx context.Context, user *dom_user.User, collection *collection.Collection, recipientPublicKey []byte, userPassword string) (*keys.EncryptedCollectionKey, error)
	EncryptCollectionKeyForMultipleRecipients(ctx context.Context, user *dom_user.User, collection *collection.Collection, recipients []SharingRecipient, userPassword string) (map[string]*keys.EncryptedCollectionKey, error)
	ValidateRecipientPublicKey(publicKey []byte) error
	RotateCollectionKey(
		ctx context.Context,
		user *dom_user.User,
		collection *dom_collection.Collection,
		password string,
		rotationReason string,
	) (*keys.EncryptedCollectionKey, error)
}

// SharingRecipient represents a recipient for collection sharing
type SharingRecipient struct {
	Email     string `json:"email"`
	PublicKey []byte `json:"public_key"`
	UserID    string `json:"user_id"`
}

// SharingEncryptionResult represents the result of encrypting for multiple recipients
type SharingEncryptionResult struct {
	RecipientEmail         string                       `json:"recipient_email"`
	EncryptedCollectionKey *keys.EncryptedCollectionKey `json:"encrypted_collection_key"`
	Success                bool                         `json:"success"`
	Error                  string                       `json:"error,omitempty"`
}

// collectionEncryptionService implements the enhanced CollectionEncryptionService interface
type collectionEncryptionService struct {
	logger                      *zap.Logger
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	collectionDecryptionService CollectionDecryptionService
}

// NewCollectionEncryptionService creates a new enhanced collection encryption service
func NewCollectionEncryptionService(
	logger *zap.Logger,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	collectionDecryptionService CollectionDecryptionService,
) CollectionEncryptionService {
	logger = logger.Named("CollectionEncryptionService")
	return &collectionEncryptionService{
		logger:                      logger,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		collectionDecryptionService: collectionDecryptionService,
	}
}

// Existing methods remain unchanged...
func (s *collectionEncryptionService) ExecuteForCreateCollectionKeyAndEncryptWithMasterKey(ctx context.Context, user *dom_user.User, password string) (*dom_keys.EncryptedCollectionKey, []byte, error) {
	s.logger.Debug("🔑 Starting E2EE key chain Encryption",
		zap.String("userID", user.ID.String()),
	)

	// Derive keyEncryptionKey from password
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.logger.Error("❌ Failed to derive key encryption key", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	// Decrypt masterKey with keyEncryptionKey
	masterKey, err := s.decryptMasterKey(user, keyEncryptionKey)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to decrypt master key - incorrect password?", err)
	}
	defer crypto.ClearBytes(masterKey)

	// Generate random collectionKey
	collectionKey, err := crypto.GenerateRandomBytes(crypto.CollectionKeySize)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to generate collection key", err)
	}

	// Encrypt collectionKey with masterKey
	encryptedData, err := crypto.EncryptWithSecretBox(collectionKey, masterKey)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to encrypt collection key", err)
	}

	// Create structured key
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

// Complete end-to-end collection sharing encryption
func (s *collectionEncryptionService) EncryptCollectionKeyForSharing(
	ctx context.Context,
	user *dom_user.User,
	collection *collection.Collection,
	recipientPublicKey []byte,
	userPassword string,
) (*keys.EncryptedCollectionKey, error) {
	s.logger.Info("🔐 Starting complete E2EE collection sharing encryption",
		zap.String("collectionID", collection.ID.String()),
		zap.Int("recipientPublicKeyLength", len(recipientPublicKey)))

	// STEP 1: Validate recipient public key
	if err := s.ValidateRecipientPublicKey(recipientPublicKey); err != nil {
		return nil, fmt.Errorf("invalid recipient public key: %w", err)
	}

	// STEP 2: Decrypt the collection key using the complete E2EE chain
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, userPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key chain for sharing: %w", err)
	}
	defer crypto.ClearBytes(collectionKey)

	s.logger.Debug("✅ Successfully decrypted collection key for sharing")

	// STEP 3: Encrypt collection key for recipient using BoxSeal
	s.logger.Debug("🔐 Encrypting collection key for recipient using BoxSeal")
	encryptedForRecipient, err := crypto.EncryptWithBoxSeal(collectionKey, recipientPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt collection key for recipient: %w", err)
	}

	// STEP 4: Create properly structured EncryptedCollectionKey for sharing
	encryptedCollectionKey := keys.NewEncryptedCollectionKeyFromBoxSeal(encryptedForRecipient)

	// STEP 5: Validate the encrypted result
	if err := s.validateEncryptedKeyForSharing(encryptedCollectionKey, recipientPublicKey); err != nil {
		return nil, fmt.Errorf("encrypted collection key validation failed: %w", err)
	}

	s.logger.Info("✅ Successfully encrypted collection key for sharing using complete E2EE chain",
		zap.String("collectionID", collection.ID.String()),
		zap.Int("encryptedKeyLength", len(encryptedCollectionKey.ToBoxSealBytes())))

	return encryptedCollectionKey, nil
}

// Encrypt collection key for multiple recipients efficiently
func (s *collectionEncryptionService) EncryptCollectionKeyForMultipleRecipients(
	ctx context.Context,
	user *dom_user.User,
	collection *collection.Collection,
	recipients []SharingRecipient,
	userPassword string,
) (map[string]*keys.EncryptedCollectionKey, error) {
	s.logger.Info("🔐 Starting batch collection sharing encryption",
		zap.String("collectionID", collection.ID.String()),
		zap.Int("recipientCount", len(recipients)))

	if len(recipients) == 0 {
		return nil, errors.NewAppError("no recipients provided", nil)
	}

	// STEP 1: Decrypt collection key once (efficiency optimization)
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, userPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key chain for batch sharing: %w", err)
	}
	defer crypto.ClearBytes(collectionKey)

	s.logger.Debug("✅ Successfully decrypted collection key for batch sharing")

	// STEP 2: Encrypt for each recipient
	results := make(map[string]*keys.EncryptedCollectionKey)
	errors := make([]string, 0)

	for i, recipient := range recipients {
		s.logger.Debug("🔐 Encrypting for recipient",
			zap.Int("recipientIndex", i+1),
			zap.Int("totalRecipients", len(recipients)),
			zap.String("recipientEmail", recipient.Email))

		// Validate recipient public key
		if err := s.ValidateRecipientPublicKey(recipient.PublicKey); err != nil {
			errorMsg := fmt.Sprintf("invalid public key for %s: %v", recipient.Email, err)
			errors = append(errors, errorMsg)
			s.logger.Warn("⚠️ Skipping recipient due to invalid public key",
				zap.String("recipientEmail", recipient.Email),
				zap.Error(err))
			continue
		}

		// Encrypt collection key for this recipient
		encryptedForRecipient, err := crypto.EncryptWithBoxSeal(collectionKey, recipient.PublicKey)
		if err != nil {
			errorMsg := fmt.Sprintf("encryption failed for %s: %v", recipient.Email, err)
			errors = append(errors, errorMsg)
			s.logger.Warn("⚠️ Skipping recipient due to encryption failure",
				zap.String("recipientEmail", recipient.Email),
				zap.Error(err))
			continue
		}

		// Create structured encrypted key
		encryptedCollectionKey := keys.NewEncryptedCollectionKeyFromBoxSeal(encryptedForRecipient)

		// Validate the result
		if err := s.validateEncryptedKeyForSharing(encryptedCollectionKey, recipient.PublicKey); err != nil {
			errorMsg := fmt.Sprintf("validation failed for %s: %v", recipient.Email, err)
			errors = append(errors, errorMsg)
			s.logger.Warn("⚠️ Skipping recipient due to validation failure",
				zap.String("recipientEmail", recipient.Email),
				zap.Error(err))
			continue
		}

		results[recipient.Email] = encryptedCollectionKey
		s.logger.Debug("✅ Successfully encrypted for recipient",
			zap.String("recipientEmail", recipient.Email))
	}

	// STEP 3: Return results and log summary
	successCount := len(results)
	errorCount := len(errors)

	s.logger.Info("✅ Completed batch collection sharing encryption",
		zap.String("collectionID", collection.ID.String()),
		zap.Int("successfulRecipients", successCount),
		zap.Int("failedRecipients", errorCount),
		zap.Int("totalRecipients", len(recipients)))

	if errorCount > 0 {
		s.logger.Warn("⚠️ Some recipients failed encryption",
			zap.Strings("errors", errors))
	}

	if successCount == 0 {
		return nil, fmt.Errorf("failed to encrypt collection key for any recipients. Errors: %v", errors)
	}

	return results, nil
}

// Validate recipient public key format and size
func (s *collectionEncryptionService) ValidateRecipientPublicKey(publicKey []byte) error {
	if len(publicKey) == 0 {
		return fmt.Errorf("public key cannot be empty")
	}

	// BoxSeal expects 32-byte public keys for Curve25519
	expectedLength := crypto.BoxPublicKeySize // Should be 32
	if len(publicKey) != expectedLength {
		return fmt.Errorf("invalid public key length: expected %d bytes, got %d bytes", expectedLength, len(publicKey))
	}

	// Additional validation: check if key is all zeros (invalid)
	allZeros := true
	for _, b := range publicKey {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if allZeros {
		return fmt.Errorf("public key cannot be all zeros")
	}

	return nil
}

// validateEncryptedKeyForSharing validates the encrypted key structure for sharing
func (s *collectionEncryptionService) validateEncryptedKeyForSharing(encryptedKey *keys.EncryptedCollectionKey, recipientPublicKey []byte) error {
	// Get the box_seal bytes
	encryptedBytes := encryptedKey.ToBoxSealBytes()
	if encryptedBytes == nil {
		return fmt.Errorf("encrypted key is nil")
	}

	// Verify it's the right length for box_seal format
	// BoxSeal format: ephemeral_public_key (32) + nonce (24) + ciphertext + auth_tag (16)
	expectedMinLength := crypto.BoxPublicKeySize + crypto.BoxNonceSize + crypto.BoxOverhead
	if len(encryptedBytes) < expectedMinLength {
		return fmt.Errorf("encrypted key too short: got %d, expected at least %d",
			len(encryptedBytes), expectedMinLength)
	}

	s.logger.Debug("✅ Encrypted key validation passed",
		zap.Int("encryptedKeyLength", len(encryptedBytes)),
		zap.Int("expectedMinLength", expectedMinLength),
		zap.Int("recipientPublicKeyLength", len(recipientPublicKey)))

	return nil
}

// Helper: Decrypt masterKey with keyEncryptionKey (E2EE spec)
func (s *collectionEncryptionService) decryptMasterKey(user *dom_user.User, keyEncryptionKey []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
}
