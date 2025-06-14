// native/desktop/maplefile-cli/internal/usecase/recovery/complete.go
package recovery

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recoverydto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CompleteRecoveryUseCase defines the interface for completing account recovery
type CompleteRecoveryUseCase interface {
	Execute(ctx context.Context, recoveryToken string, newPassword string, masterKeyFromRecovery []byte) (*recoverydto.RecoveryCompleteResponseDTO, error)
}

// completeRecoveryUseCase implements the CompleteRecoveryUseCase interface
type completeRecoveryUseCase struct {
	logger          *zap.Logger
	recoveryDTORepo recoverydto.RecoveryDTORepository
	recoveryRepo    recovery.RecoveryRepository
	userRepo        user.Repository
}

// NewCompleteRecoveryUseCase creates a new complete recovery use case
func NewCompleteRecoveryUseCase(
	logger *zap.Logger,
	recoveryDTORepo recoverydto.RecoveryDTORepository,
	recoveryRepo recovery.RecoveryRepository,
	userRepo user.Repository,
) CompleteRecoveryUseCase {
	logger = logger.Named("CompleteRecoveryUseCase")
	return &completeRecoveryUseCase{
		logger:          logger,
		recoveryDTORepo: recoveryDTORepo,
		recoveryRepo:    recoveryRepo,
		userRepo:        userRepo,
	}
}

// Execute completes the recovery process with new password
func (uc *completeRecoveryUseCase) Execute(ctx context.Context, recoveryToken string, newPassword string, masterKeyFromRecovery []byte) (*recoverydto.RecoveryCompleteResponseDTO, error) {
	//
	// STEP 1: Validate inputs
	//
	if recoveryToken == "" {
		return nil, errors.NewAppError("recovery token is required", nil)
	}
	if newPassword == "" {
		return nil, errors.NewAppError("new password is required", nil)
	}
	if len(masterKeyFromRecovery) != crypto.MasterKeySize {
		return nil, errors.NewAppError("invalid master key size", nil)
	}

	// Sanitize inputs
	recoveryToken = strings.TrimSpace(recoveryToken)

	//
	// STEP 2: Get recovery token from local storage (if exists)
	//
	localToken, _ := uc.recoveryRepo.GetTokenByValue(ctx, recoveryToken)
	if localToken != nil {
		// Check if token is expired
		if localToken.IsExpired() {
			uc.logger.Warn("Recovery token has expired", zap.String("token", recoveryToken[:10]+"..."))
			// Continue anyway - let server validate
		}

		// Check if already used
		if localToken.Used {
			uc.logger.Warn("Recovery token already used", zap.String("token", recoveryToken[:10]+"..."))
			// Continue anyway - server is source of truth
		}
	}

	//
	// STEP 3: Generate new salt for the new password
	//
	newSalt, err := crypto.GenerateRandomBytes(crypto.Argon2SaltSize)
	if err != nil {
		return nil, errors.NewAppError("failed to generate new salt", err)
	}

	//
	// STEP 4: Derive new key encryption key from new password
	//
	newKeyEncryptionKey, err := crypto.DeriveKeyFromPassword(newPassword, newSalt)
	if err != nil {
		return nil, errors.NewAppError("failed to derive key from new password", err)
	}

	//
	// STEP 5: Re-encrypt master key with new key encryption key
	//
	encryptedMasterKey, err := crypto.EncryptWithSecretBox(masterKeyFromRecovery, newKeyEncryptionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt master key with new password", err)
	}

	//
	// STEP 6: Generate new key pair
	//
	publicKey, privateKey, _, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, errors.NewAppError("failed to generate new key pair", err)
	}

	//
	// STEP 7: Encrypt private key with master key
	//
	encryptedPrivateKey, err := crypto.EncryptWithSecretBox(privateKey, masterKeyFromRecovery)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt private key", err)
	}

	//
	// STEP 8: Generate new recovery key
	//
	newRecoveryKey, err := crypto.GenerateRandomBytes(crypto.RecoveryKeySize)
	if err != nil {
		return nil, errors.NewAppError("failed to generate new recovery key", err)
	}

	//
	// STEP 9: Encrypt recovery key with master key
	//
	encryptedRecoveryKey, err := crypto.EncryptWithSecretBox(newRecoveryKey, masterKeyFromRecovery)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt recovery key", err)
	}

	//
	// STEP 10: Encrypt master key with recovery key (for future recovery)
	//
	masterKeyEncryptedWithRecoveryKey, err := crypto.EncryptWithSecretBox(masterKeyFromRecovery, newRecoveryKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt master key with recovery key", err)
	}

	//
	// STEP 11: Prepare complete recovery request
	//
	// Combine nonce and ciphertext for each encrypted item
	encMasterKeyBytes := append(encryptedMasterKey.Nonce, encryptedMasterKey.Ciphertext...)
	encPrivateKeyBytes := append(encryptedPrivateKey.Nonce, encryptedPrivateKey.Ciphertext...)
	encRecoveryKeyBytes := append(encryptedRecoveryKey.Nonce, encryptedRecoveryKey.Ciphertext...)
	encMasterKeyWithRecoveryBytes := append(masterKeyEncryptedWithRecoveryKey.Nonce, masterKeyEncryptedWithRecoveryKey.Ciphertext...)

	request := &recoverydto.RecoveryCompleteRequestDTO{
		RecoveryToken:                        recoveryToken,
		NewSalt:                              base64.RawURLEncoding.EncodeToString(newSalt),
		NewEncryptedMasterKey:                base64.RawURLEncoding.EncodeToString(encMasterKeyBytes),
		NewEncryptedPrivateKey:               base64.RawURLEncoding.EncodeToString(encPrivateKeyBytes),
		NewEncryptedRecoveryKey:              base64.RawURLEncoding.EncodeToString(encRecoveryKeyBytes),
		NewMasterKeyEncryptedWithRecoveryKey: base64.RawURLEncoding.EncodeToString(encMasterKeyWithRecoveryBytes),
	}

	//
	// STEP 12: Call cloud service to complete recovery
	//
	uc.logger.Debug("Completing recovery with cloud")

	response, err := uc.recoveryDTORepo.CompleteRecoveryFromCloud(ctx, request)
	if err != nil {
		uc.logger.Error("Failed to complete recovery with cloud", zap.Error(err))
		return nil, err
	}

	//
	// STEP 13: Update local user record if recovery was successful
	//
	if response.Success && localToken != nil {
		// Get session to find user email
		session, _ := uc.recoveryRepo.GetSessionByID(ctx, localToken.SessionID)
		if session != nil {
			// Get or create user
			existingUser, _ := uc.userRepo.GetByEmail(ctx, session.Email)
			if existingUser == nil {
				// Create new user if doesn't exist
				existingUser = &user.User{
					ID:        session.UserID,
					Email:     session.Email,
					Status:    user.UserStatusActive,
					CreatedAt: time.Now(),
				}
			}

			// Update user with new encryption data
			currentTime := time.Now()
			existingUser.PasswordSalt = newSalt
			existingUser.PublicKey = keys.PublicKey{Key: publicKey}
			existingUser.EncryptedMasterKey = keys.EncryptedMasterKey{
				Ciphertext: encryptedMasterKey.Ciphertext,
				Nonce:      encryptedMasterKey.Nonce,
				KeyVersion: existingUser.EncryptedMasterKey.KeyVersion + 1,
				RotatedAt:  &currentTime,
			}
			existingUser.EncryptedPrivateKey = keys.EncryptedPrivateKey{
				Ciphertext: encryptedPrivateKey.Ciphertext,
				Nonce:      encryptedPrivateKey.Nonce,
			}
			existingUser.EncryptedRecoveryKey = keys.EncryptedRecoveryKey{
				Ciphertext: encryptedRecoveryKey.Ciphertext,
				Nonce:      encryptedRecoveryKey.Nonce,
			}
			existingUser.MasterKeyEncryptedWithRecoveryKey = keys.MasterKeyEncryptedWithRecoveryKey{
				Ciphertext: masterKeyEncryptedWithRecoveryKey.Ciphertext,
				Nonce:      masterKeyEncryptedWithRecoveryKey.Nonce,
			}
			existingUser.LastPasswordChange = currentTime
			existingUser.ModifiedAt = currentTime

			// Save updated user
			if err := uc.userRepo.UpsertByEmail(ctx, existingUser); err != nil {
				uc.logger.Error("Failed to update user after recovery", zap.Error(err))
				// Continue anyway - recovery was successful on server
			}

			// Mark token as used
			now := time.Now()
			localToken.Used = true
			localToken.UsedAt = &now
			if err := uc.recoveryRepo.UpdateToken(ctx, localToken); err != nil {
				uc.logger.Error("Failed to mark recovery token as used", zap.Error(err))
				// Continue anyway
			}

			// Mark session as completed
			if session.CompletedAt == nil {
				session.CompletedAt = &now
				if err := uc.recoveryRepo.UpdateSession(ctx, session); err != nil {
					uc.logger.Error("Failed to mark recovery session as completed", zap.Error(err))
					// Continue anyway
				}
			}
		}
	}

	uc.logger.Info("Successfully completed recovery",
		zap.Bool("success", response.Success),
		zap.String("message", response.Message))

	return response, nil
}
