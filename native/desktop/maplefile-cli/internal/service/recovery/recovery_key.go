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
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/security"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
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
		if err == nil { // Handle user being nil but no error returned by use case
			err = errors.NewAppError("user not found", nil)
		}
		return nil, err // Return the AppError from use case
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

	// Ensure keyEncryptionKey is cleared
	defer crypto.ClearBytes(keyEncryptionKey)

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
			Operation:    "show_recovery_key_decrypt_master",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: "incorrect password or corrupted master key",
		})
		return nil, errors.NewAppError("incorrect password", nil) // Don't include underlying crypto error
	}

	// Ensure masterKey is cleared
	defer crypto.ClearBytes(masterKey)

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

	// Ensure recoveryKey is cleared after use/formatting
	defer crypto.ClearBytes(recoveryKey)

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

	// User struct does not have RecoveryKeyUpdatedAt. Use CreatedAt.
	createdAt := user.CreatedAt

	return &RecoveryKeyOutput{
		RecoveryKey:       formattedKey,
		RecoveryKeyBase64: recoveryKeyBase64,
		CreatedAt:         createdAt.Format("2006-01-02"),
		Instructions:      "Keep this recovery key in a safe place. You'll need it to recover your account if you forget your password.",
	}, nil
}

// GenerateNewRecoveryKey creates a new recovery key
func (s *recoveryKeyService) GenerateNewRecoveryKey(ctx context.Context, email string, password string) (output *RecoveryKeyOutput, err error) {
	s.logger.Info("🔑 Generating new recovery key", zap.String("email", email))

	//
	// STEP 1: Validate inputs
	//
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}

	if err = s.passwordValidationService.ValidateForCryptoOperations(password); err != nil {
		return nil, errors.NewAppError("invalid password", err)
	}

	// Sanitize email
	email = strings.ToLower(strings.TrimSpace(email))

	//
	// STEP 2: Start transaction
	//
	// Note: The original code's defer logic for transaction discard was problematic.
	// A common pattern is to discard on error return or panic.
	// We will add a manual commit call on success.
	if err = s.userRepo.OpenTransaction(); err != nil {
		return nil, errors.NewAppError("failed to open transaction", err)
	}
	// Defer discard in case of error or panic before commit
	var commitErr error // Track potential commit error
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("recovered from panic, discarding transaction", zap.Any("panic", r))
			s.userRepo.DiscardTransaction()
			panic(r) // Re-panic after cleanup
		}
		// If commitErr is set, the commit failed, transaction is likely invalid/rolled back by repo impl,
		// or we explicitly discard. If err != nil, the function returned an error before commit.
		// Note: This defer runs *after* the function returns, so `err` captured here is the return value error.
		if err != nil || commitErr != nil {
			s.logger.Debug("GenerateNewRecoveryKey returning with error or commit failed, ensuring transaction discarded.")
			// Assuming DiscardTransaction is safe to call multiple times or checks internally.
			s.userRepo.DiscardTransaction()
		}
	}()

	//
	// STEP 3: Get user
	//
	user, err := s.getByEmailUseCase.Execute(ctx, email)
	if err != nil || user == nil {
		// Execute should return AppError if user not found etc.
		if err == nil { // Handle user being nil but no error returned by use case
			err = errors.NewAppError("user not found", nil)
		}
		return nil, err // Return the AppError from use case
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
	defer crypto.ClearBytes(keyEncryptionKey) // Clear KEK

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
			Operation:    "generate_recovery_key_decrypt_master",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: "incorrect password or corrupted master key",
		})
		// Return a generic "incorrect password" error to the user
		return nil, errors.NewAppError("incorrect password", nil) // Don't expose decryption failure detail
	}
	defer crypto.ClearBytes(masterKey) // Clear master key

	//
	// STEP 6: Generate new recovery key
	//
	newRecoveryKey, err := crypto.GenerateRandomBytes(crypto.RecoveryKeySize)
	if err != nil {
		return nil, errors.NewAppError("failed to generate recovery key", err)
	}
	defer crypto.ClearBytes(newRecoveryKey) // Clear raw new recovery key after base64 encoding

	//
	// STEP 7: Encrypt new recovery key with master key
	//
	encryptedRecoveryKey, err := crypto.EncryptWithSecretBox(newRecoveryKey, masterKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt new recovery key", err)
	}

	//
	// STEP 8: Encrypt master key with new recovery key (for future recovery)
	//
	masterKeyEncryptedWithRecoveryKey, err := crypto.EncryptWithSecretBox(masterKey, newRecoveryKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt master key with new recovery key", err)
	}

	//
	// STEP 9: Update user with new recovery key data
	//
	user.EncryptedRecoveryKey = keys.EncryptedRecoveryKey{
		Ciphertext: encryptedRecoveryKey.Ciphertext,
		Nonce:      encryptedRecoveryKey.Nonce,
	}
	user.MasterKeyEncryptedWithRecoveryKey = keys.MasterKeyEncryptedWithRecoveryKey{
		Ciphertext: masterKeyEncryptedWithRecoveryKey.Ciphertext,
		Nonce:      masterKeyEncryptedWithRecoveryKey.Nonce,
	}
	// User struct does not have RecoveryKeyUpdatedAt. Cannot update timestamp specifically for recovery key.

	//
	// STEP 10: Save updated user
	//
	if err = s.upsertByEmailUseCase.Execute(ctx, user); err != nil {
		// upsertByEmailUseCase should return AppError
		return nil, err
	}

	//
	// STEP 11: Commit transaction
	//
	commitErr = s.userRepo.CommitTransaction() // Assign to commitErr for defer
	if commitErr != nil {
		err = errors.NewAppError("failed to commit transaction", commitErr) // Assign to named return variable
		return nil, err
	}

	// Sensitive data clear already handled by defers

	//
	// STEP 12: Format recovery key for display
	//
	recoveryKeyBase64 := base64.StdEncoding.EncodeToString(newRecoveryKey)
	formattedKey := s.formatRecoveryKey(recoveryKeyBase64)

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

	// Use CreatedAt as RecoveryKeyUpdatedAt field does not exist
	return &RecoveryKeyOutput{
		RecoveryKey:       formattedKey,
		RecoveryKeyBase64: recoveryKeyBase64,
		CreatedAt:         user.CreatedAt.Format("2006-01-02"), // Use the original creation timestamp
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
		if err == nil { // Handle user being nil but no error returned by use case
			err = errors.NewAppError("user not found", nil)
		}
		return err // Return the AppError from use case
	}

	//
	// STEP 3: Decode recovery key
	//
	// Remove formatting if present
	cleanKey := strings.ReplaceAll(recoveryKey, "-", "")
	cleanKey = strings.ReplaceAll(cleanKey, " ", "")

	recoveryKeyBytes, err := base64.StdEncoding.DecodeString(cleanKey)
	if err != nil {
		// Try URL-safe encoding (though standard is more likely for display)
		recoveryKeyBytes, err = base64.RawURLEncoding.DecodeString(cleanKey)
		if err != nil {
			return errors.NewAppError("invalid recovery key format", err)
		}
	}
	defer crypto.ClearBytes(recoveryKeyBytes) // Clear decoded key after use

	// Validate key size
	if len(recoveryKeyBytes) != crypto.RecoveryKeySize {
		return errors.NewAppError(fmt.Sprintf("invalid recovery key size: expected %d bytes, got %d", crypto.RecoveryKeySize, len(recoveryKeyBytes)), nil)
	}

	//
	// STEP 4: Try to decrypt master key with recovery key
	//
	if len(user.MasterKeyEncryptedWithRecoveryKey.Ciphertext) == 0 {
		// Log warning if trying to validate when none is configured
		s.logger.Warn("Attempted to validate recovery key when none is configured", zap.String("email", email))
		return errors.NewAppError("no recovery key configured for this account", nil)
	}

	// Attempt decryption. We only need to know if it succeeds or fails.
	_, err = crypto.DecryptWithSecretBox(
		user.MasterKeyEncryptedWithRecoveryKey.Ciphertext,
		user.MasterKeyEncryptedWithRecoveryKey.Nonce,
		recoveryKeyBytes,
	)

	if err != nil {
		// Log failure
		s.cryptoAuditService.LogCryptoOperation(ctx, &security.CryptoAuditEvent{
			Operation:    "validate_recovery_key",
			UserID:       user.ID.String(),
			Success:      false,
			ErrorMessage: "invalid recovery key (decryption failed)", // More specific error message for audit
		})
		// Return a generic "invalid" message to the user to prevent timing attacks
		return errors.NewAppError("invalid recovery key", nil) // Do not include crypto error
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
}
