// native/desktop/maplefile-cli/internal/service/recovery/recovery_key.go
package recovery

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/security"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// RecoveryKeyService provides functionality for managing recovery keys
type RecoveryKeyService interface {
	// ShowRecoveryKey displays the user's current recovery key
	ShowRecoveryKey(ctx context.Context, email string, password string) (*RecoveryKeyOutput, error)

	// GenerateNewRecoveryKey creates a new recovery key (requires current password)
	GenerateNewRecoveryKey(ctx context.Context, email string, password string) (*RecoveryKeyOutput, error)

	// ValidateRecoveryKey checks if a recovery key is valid for a user
	ValidateRecoveryKey(ctx context.Context, email string, recoveryKey string) error
}

// RecoveryKeyOutput represents the output of recovery key operations
type RecoveryKeyOutput struct {
	RecoveryKey       string `json:"recovery_key"`
	RecoveryKeyBase64 string `json:"recovery_key_base64"`
	CreatedAt         string `json:"created_at"`
	Instructions      string `json:"instructions"`
}

// recoveryKeyService implements the RecoveryKeyService interface
type recoveryKeyService struct {
	logger                    *zap.Logger
	userRepo                  user.Repository
	getByEmailUseCase         uc_user.GetByEmailUseCase
	upsertByEmailUseCase      uc_user.UpsertByEmailUseCase
	passwordValidationService security.PasswordValidationService
	cryptoAuditService        security.CryptoAuditService
}

// NewRecoveryKeyService creates a new recovery key service
func NewRecoveryKeyService(
	logger *zap.Logger,
	userRepo user.Repository,
	getByEmailUseCase uc_user.GetByEmailUseCase,
	upsertByEmailUseCase uc_user.UpsertByEmailUseCase,
	passwordValidationService security.PasswordValidationService,
	cryptoAuditService security.CryptoAuditService,
) RecoveryKeyService {
	logger = logger.Named("RecoveryKeyService")
	return &recoveryKeyService{
		logger:                    logger,
		userRepo:                  userRepo,
		getByEmailUseCase:         getByEmailUseCase,
		upsertByEmailUseCase:      upsertByEmailUseCase,
		passwordValidationService: passwordValidationService,
		cryptoAuditService:        cryptoAuditService,
	}
}

// ShowRecoveryKey displays the user's current recovery key
func (s *recoveryKeyService) ShowRecoveryKey(ctx context.Context, email string, password string) (*RecoveryKeyOutput, error) {
	s.logger.Info("🔑 Showing recovery key", zap.String("email", email))

	//
	// STEP 1: Validate inputs
	//
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}

	if err := s.passwordValidationService.ValidateForCryptoOperations(password); err != nil {
		return nil, errors.NewAppError("invalid password", err)
	}

	// Sanitize email
	email = strings.ToLower(strings.TrimSpace(email))

	//
	// STEP 2: Get user
	//
	user, err := s.getByEmailUseCase.Execute(ctx, email)
	if err != nil || user == nil {
		return nil, errors.NewAppError("user not found", err)
	}

	//
	// STEP 3: Derive key encryption key from password
	//
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:    "show_recovery_key_derive_kek",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: err.Error(),
		})
		return nil, errors.NewAppError("failed to derive key from password", err)
	}

	//
	// STEP 4: Decrypt master key
	//
	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:    "generate_recovery_key_decrypt_master",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: "incorrect password or corrupted master key",
		})
		return nil, errors.NewAppError("incorrect password", err)
	}

	//
	// STEP 6: Generate new recovery key
	//
	newRecoveryKey, err := crypto.GenerateRandomBytes(crypto.RecoveryKeySize)
	if err != nil {
		crypto.ClearBytes(masterKey)
		crypto.ClearBytes(keyEncryptionKey)
		return nil, errors.NewAppError("failed to generate recovery key", err)
	}

	//
	// STEP 7: Encrypt recovery key with master key
	//
	encryptedRecoveryKey, err := crypto.EncryptWithSecretBox(newRecoveryKey, masterKey)
	if err != nil {
		crypto.ClearBytes(masterKey)
		crypto.ClearBytes(keyEncryptionKey)
		crypto.ClearBytes(newRecoveryKey)
		return nil, errors.NewAppError("failed to encrypt recovery key", err)
	}

	//
	// STEP 8: Encrypt master key with recovery key (for future recovery)
	//
	masterKeyEncryptedWithRecoveryKey, err := crypto.EncryptWithSecretBox(masterKey, newRecoveryKey)
	if err != nil {
		crypto.ClearBytes(masterKey)
		crypto.ClearBytes(keyEncryptionKey)
		crypto.ClearBytes(newRecoveryKey)
		return nil, errors.NewAppError("failed to encrypt master key with recovery key", err)
	}

	//
	// STEP 9: Update user with new recovery key
	//
	user.EncryptedRecoveryKey = keys.EncryptedRecoveryKey{
		Ciphertext: encryptedRecoveryKey.Ciphertext,
		Nonce:      encryptedRecoveryKey.Nonce,
	}
	user.MasterKeyEncryptedWithRecoveryKey = keys.MasterKeyEncryptedWithRecoveryKey{
		Ciphertext: masterKeyEncryptedWithRecoveryKey.Ciphertext,
		Nonce:      masterKeyEncryptedWithRecoveryKey.Nonce,
	}

	// Save updated user
	if err := s.upsertByEmailUseCase.Execute(ctx, user); err != nil {
		crypto.ClearBytes(masterKey)
		crypto.ClearBytes(keyEncryptionKey)
		crypto.ClearBytes(newRecoveryKey)
		return nil, errors.NewAppError("failed to save new recovery key", err)
	}

	//
	// STEP 10: Commit transaction
	//
	if err := s.userRepo.CommitTransaction(); err != nil {
		crypto.ClearBytes(masterKey)
		crypto.ClearBytes(keyEncryptionKey)
		crypto.ClearBytes(newRecoveryKey)
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	// Clear sensitive data
	crypto.ClearBytes(masterKey)
	crypto.ClearBytes(keyEncryptionKey)

	//
	// STEP 11: Format recovery key for display
	//
	recoveryKeyBase64 := base64.StdEncoding.EncodeToString(newRecoveryKey)
	formattedKey := s.formatRecoveryKey(recoveryKeyBase64)

	// Clear the raw recovery key after formatting
	crypto.ClearBytes(newRecoveryKey)

	// Log successful operation
	s.cryptoAuditService.LogKeyOperation(ctx, &security.CryptoAuditEvent{
		Operation: "generate_new_recovery_key",
		UserID:    user.ID.String(),
		Success:   true,
		Metadata: map[string]interface{}{
			"key_rotation": true,
		},
	})

	s.logger.Info("✅ New recovery key generated successfully", zap.String("email", email))

	return &RecoveryKeyOutput{
		RecoveryKey:       formattedKey,
		RecoveryKeyBase64: recoveryKeyBase64,
		CreatedAt:         "Just now",
		Instructions:      "Your new recovery key has been generated. Please save it immediately as it won't be shown again.",
	}, nil
}

// ValidateRecoveryKey checks if a recovery key is valid for a user
func (s *recoveryKeyService) ValidateRecoveryKey(ctx context.Context, email string, recoveryKey string) error {
	s.logger.Debug("🔑 Validating recovery key", zap.String("email", email))

	//
	// STEP 1: Validate inputs
	//
	if email == "" {
		return errors.NewAppError("email is required", nil)
	}
	if recoveryKey == "" {
		return errors.NewAppError("recovery key is required", nil)
	}

	// Sanitize email
	email = strings.ToLower(strings.TrimSpace(email))

	//
	// STEP 2: Get user
	//
	user, err := s.getByEmailUseCase.Execute(ctx, email)
	if err != nil || user == nil {
		return errors.NewAppError("user not found", err)
	}

	//
	// STEP 3: Decode recovery key
	//
	// Remove formatting if present
	cleanKey := strings.ReplaceAll(recoveryKey, "-", "")
	cleanKey = strings.ReplaceAll(cleanKey, " ", "")

	recoveryKeyBytes, err := base64.StdEncoding.DecodeString(cleanKey)
	if err != nil {
		// Try URL-safe encoding
		recoveryKeyBytes, err = base64.RawURLEncoding.DecodeString(cleanKey)
		if err != nil {
			return errors.NewAppError("invalid recovery key format", err)
		}
	}

	// Validate key size
	if len(recoveryKeyBytes) != crypto.RecoveryKeySize {
		return errors.NewAppError("invalid recovery key size", nil)
	}

	//
	// STEP 4: Try to decrypt master key with recovery key
	//
	if len(user.MasterKeyEncryptedWithRecoveryKey.Ciphertext) == 0 {
		return errors.NewAppError("no recovery key configured for this account", nil)
	}

	_, err = crypto.DecryptWithSecretBox(
		user.MasterKeyEncryptedWithRecoveryKey.Ciphertext,
		user.MasterKeyEncryptedWithRecoveryKey.Nonce,
		recoveryKeyBytes,
	)

	// Clear recovery key bytes
	crypto.ClearBytes(recoveryKeyBytes)

	if err != nil {
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:    "validate_recovery_key",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: "invalid recovery key",
		})
		return errors.NewAppError("invalid recovery key", err)
	}

	// Log successful validation
	s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
		Operation: "validate_recovery_key",
		UserID:    user.ID.String(),
		Success:   true,
	})

	s.logger.Debug("✅ Recovery key validated successfully", zap.String("email", email))

	return nil
}

// formatRecoveryKey formats a base64 recovery key into groups for readability
func (s *recoveryKeyService) formatRecoveryKey(base64Key string) string {
	// Remove any existing formatting
	cleanKey := strings.ReplaceAll(base64Key, "-", "")
	cleanKey = strings.ReplaceAll(cleanKey, " ", "")

	// Split into groups of 4 characters
	var groups []string
	for i := 0; i < len(cleanKey); i += 4 {
		end := i + 4
		if end > len(cleanKey) {
			end = len(cleanKey)
		}
		groups = append(groups, cleanKey[i:end])
	}

	// Join with hyphens
	return strings.Join(groups, "-")
}    "show_recovery_key_decrypt_master",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: "incorrect password or corrupted master key",
		})
		return nil, errors.NewAppError("incorrect password", err)
	}

	//
	// STEP 5: Decrypt recovery key
	//
	if len(user.EncryptedRecoveryKey.Ciphertext) == 0 {
		return nil, errors.NewAppError("no recovery key found for this account", nil)
	}

	recoveryKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedRecoveryKey.Ciphertext,
		user.EncryptedRecoveryKey.Nonce,
		masterKey,
	)
	if err != nil {
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:    "show_recovery_key_decrypt_recovery",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: err.Error(),
		})
		return nil, errors.NewAppError("failed to decrypt recovery key", err)
	}

	// Clear sensitive data
	crypto.ClearBytes(masterKey)
	crypto.ClearBytes(keyEncryptionKey)

	//
	// STEP 6: Format recovery key for display
	//
	recoveryKeyBase64 := base64.StdEncoding.EncodeToString(recoveryKey)

	// Create human-readable format (groups of 4 characters)
	formattedKey := s.formatRecoveryKey(recoveryKeyBase64)

	// Log successful operation
	s.cryptoAuditService.LogKeyOperation(ctx, &security.CryptoAuditEvent{
		Operation: "show_recovery_key",
		UserID:    user.ID.String(),
		Success:   true,
	})

	s.logger.Info("✅ Recovery key displayed successfully", zap.String("email", email))

	return &RecoveryKeyOutput{
		RecoveryKey:       formattedKey,
		RecoveryKeyBase64: recoveryKeyBase64,
		CreatedAt:         user.CreatedAt.Format("2006-01-02"),
		Instructions:      "Keep this recovery key in a safe place. You'll need it to recover your account if you forget your password.",
	}, nil
}

// GenerateNewRecoveryKey creates a new recovery key
func (s *recoveryKeyService) GenerateNewRecoveryKey(ctx context.Context, email string, password string) (*RecoveryKeyOutput, error) {
	s.logger.Info("🔑 Generating new recovery key", zap.String("email", email))

	//
	// STEP 1: Validate inputs
	//
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}

	if err := s.passwordValidationService.ValidateForCryptoOperations(password); err != nil {
		return nil, errors.NewAppError("invalid password", err)
	}

	// Sanitize email
	email = strings.ToLower(strings.TrimSpace(email))

	//
	// STEP 2: Start transaction
	//
	if err := s.userRepo.OpenTransaction(); err != nil {
		return nil, errors.NewAppError("failed to open transaction", err)
	}

	// Ensure transaction cleanup
	defer func() {
		if s.userRepo.OpenTransaction() == nil { // Check if still in transaction
			s.userRepo.DiscardTransaction()
		}
	}()

	//
	// STEP 3: Get user
	//
	user, err := s.getByEmailUseCase.Execute(ctx, email)
	if err != nil || user == nil {
		return nil, errors.NewAppError("user not found", err)
	}

	//
	// STEP 4: Derive key encryption key from password
	//
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:    "generate_recovery_key_derive_kek",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: err.Error(),
		})
		return nil, errors.NewAppError("failed to derive key from password", err)
	}

	//
	// STEP 5: Decrypt master key
	//
	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:
