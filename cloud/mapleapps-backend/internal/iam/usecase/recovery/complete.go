// cloud/mapleapps-backend/internal/iam/usecase/recovery/complete.go
package recovery

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/federateduser"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	dom_recovery "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/recovery"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/crypto"
)

type CompleteRecoveryUseCase interface {
	Execute(ctx context.Context, req *CompleteRecoveryRequest) (*CompleteRecoveryResult, error)
}

type CompleteRecoveryRequest struct {
	RecoveryToken               string `json:"recovery_token"`
	NewSalt                     string `json:"new_salt"`
	NewEncryptedMasterKey       string `json:"new_encrypted_master_key"`
	NewEncryptedPrivateKey      string `json:"new_encrypted_private_key"`
	NewEncryptedRecoveryKey     string `json:"new_encrypted_recovery_key"`
	NewMasterKeyWithRecoveryKey string `json:"new_master_key_encrypted_with_recovery_key"`
}

type CompleteRecoveryResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type completeRecoveryUseCaseImpl struct {
	config       *config.Configuration
	logger       *zap.Logger
	recoveryRepo dom_recovery.RecoveryRepository
	userRepo     dom_user.FederatedUserRepository
}

func NewCompleteRecoveryUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	recoveryRepo dom_recovery.RecoveryRepository,
	userRepo dom_user.FederatedUserRepository,
) CompleteRecoveryUseCase {
	logger = logger.Named("CompleteRecoveryUseCase")
	return &completeRecoveryUseCaseImpl{
		config:       config,
		logger:       logger,
		recoveryRepo: recoveryRepo,
		userRepo:     userRepo,
	}
}

func (uc *completeRecoveryUseCaseImpl) Execute(ctx context.Context, req *CompleteRecoveryRequest) (*CompleteRecoveryResult, error) {
	// Validate input
	if req.RecoveryToken == "" {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Recovery token is required")
	}

	// Decode recovery token
	tokenBytes, err := base64.RawURLEncoding.DecodeString(req.RecoveryToken)
	if err != nil {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Invalid recovery token")
	}

	tokenParts := strings.Split(string(tokenBytes), ":")
	if len(tokenParts) != 2 {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Invalid recovery token format")
	}

	sessionID := tokenParts[0]
	challengeID := tokenParts[1]

	// Get recovery session
	session, err := uc.recoveryRepo.GetRecoverySessionByID(ctx, sessionID)
	if err != nil {
		uc.logger.Error("Failed to get recovery session", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to complete recovery")
	}

	if session == nil {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Invalid or expired recovery session")
	}

	// Verify session
	if !session.IsVerified {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Recovery session not verified")
	}

	if session.ChallengeID != challengeID {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Invalid recovery token")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, httperror.NewForBadRequestWithSingleField("recovery_token", "Recovery session expired")
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		uc.logger.Error("Failed to get user", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to complete recovery")
	}

	if user == nil {
		return nil, httperror.NewForInternalServerError("User not found")
	}

	// Decode new encrypted keys
	newSaltBytes, err := base64.RawURLEncoding.DecodeString(req.NewSalt)
	if err != nil {
		return nil, httperror.NewForBadRequestWithSingleField("new_salt", "Invalid salt format")
	}

	// Process new encrypted master key
	newEncMasterKeyBytes, err := base64.RawURLEncoding.DecodeString(req.NewEncryptedMasterKey)
	if err != nil {
		return nil, httperror.NewForBadRequestWithSingleField("new_encrypted_master_key", "Invalid encrypted master key format")
	}

	if len(newEncMasterKeyBytes) < crypto.NonceSize {
		return nil, httperror.NewForBadRequestWithSingleField("new_encrypted_master_key", "Encrypted master key too short")
	}

	// Create new encrypted master key with history
	currentTime := time.Now()
	newHistoricalKey := keys.EncryptedHistoricalKey{
		KeyVersion:    user.SecurityData.CurrentKeyVersion + 1,
		Nonce:         newEncMasterKeyBytes[:crypto.NonceSize],
		Ciphertext:    newEncMasterKeyBytes[crypto.NonceSize:],
		RotatedAt:     currentTime,
		RotatedReason: "Password recovery",
		Algorithm:     crypto.ChaCha20Poly1305Algorithm,
	}

	// Keep previous keys history
	previousKeys := append(user.SecurityData.EncryptedMasterKey.PreviousKeys, newHistoricalKey)
	if len(previousKeys) > 5 { // Keep only last 5 keys
		previousKeys = previousKeys[len(previousKeys)-5:]
	}

	newEncryptedMasterKey := keys.EncryptedMasterKey{
		Nonce:        newEncMasterKeyBytes[:crypto.NonceSize],
		Ciphertext:   newEncMasterKeyBytes[crypto.NonceSize:],
		KeyVersion:   user.SecurityData.CurrentKeyVersion + 1,
		RotatedAt:    &currentTime,
		PreviousKeys: previousKeys,
	}

	// Process new encrypted private key
	newEncPrivateKeyBytes, err := base64.RawURLEncoding.DecodeString(req.NewEncryptedPrivateKey)
	if err != nil {
		return nil, httperror.NewForBadRequestWithSingleField("new_encrypted_private_key", "Invalid encrypted private key format")
	}

	if len(newEncPrivateKeyBytes) < crypto.NonceSize {
		return nil, httperror.NewForBadRequestWithSingleField("new_encrypted_private_key", "Encrypted private key too short")
	}

	newEncryptedPrivateKey := keys.EncryptedPrivateKey{
		Nonce:      newEncPrivateKeyBytes[:crypto.NonceSize],
		Ciphertext: newEncPrivateKeyBytes[crypto.NonceSize:],
	}

	// Process new encrypted recovery key
	newEncRecoveryKeyBytes, err := base64.RawURLEncoding.DecodeString(req.NewEncryptedRecoveryKey)
	if err != nil {
		return nil, httperror.NewForBadRequestWithSingleField("new_encrypted_recovery_key", "Invalid encrypted recovery key format")
	}

	if len(newEncRecoveryKeyBytes) < crypto.NonceSize {
		return nil, httperror.NewForBadRequestWithSingleField("new_encrypted_recovery_key", "Encrypted recovery key too short")
	}

	newEncryptedRecoveryKey := keys.EncryptedRecoveryKey{
		Nonce:      newEncRecoveryKeyBytes[:crypto.NonceSize],
		Ciphertext: newEncRecoveryKeyBytes[crypto.NonceSize:],
	}

	// Process new master key encrypted with recovery key
	newMasterWithRecoveryBytes, err := base64.RawURLEncoding.DecodeString(req.NewMasterKeyWithRecoveryKey)
	if err != nil {
		return nil, httperror.NewForBadRequestWithSingleField("new_master_key_encrypted_with_recovery_key", "Invalid format")
	}

	if len(newMasterWithRecoveryBytes) < crypto.NonceSize {
		return nil, httperror.NewForBadRequestWithSingleField("new_master_key_encrypted_with_recovery_key", "Data too short")
	}

	newMasterKeyWithRecoveryKey := keys.MasterKeyEncryptedWithRecoveryKey{
		Nonce:      newMasterWithRecoveryBytes[:crypto.NonceSize],
		Ciphertext: newMasterWithRecoveryBytes[crypto.NonceSize:],
	}

	// Update user's security data
	user.SecurityData.PasswordSalt = newSaltBytes
	user.SecurityData.EncryptedMasterKey = newEncryptedMasterKey
	user.SecurityData.EncryptedPrivateKey = newEncryptedPrivateKey
	user.SecurityData.EncryptedRecoveryKey = newEncryptedRecoveryKey
	user.SecurityData.MasterKeyEncryptedWithRecoveryKey = newMasterKeyWithRecoveryKey
	user.SecurityData.LastPasswordChange = currentTime
	user.SecurityData.CurrentKeyVersion = newEncryptedMasterKey.KeyVersion
	user.SecurityData.LastKeyRotation = &currentTime

	// Update modified timestamp
	user.ModifiedAt = currentTime

	// Save updated user
	if err := uc.userRepo.UpdateByID(ctx, user); err != nil {
		uc.logger.Error("Failed to update user", zap.Error(err))
		return nil, httperror.NewForInternalServerError("Failed to complete recovery")
	}

	// Record key rotation
	rotation := &dom_recovery.RecoveryKeyRotation{
		UserID:             user.ID,
		OldRecoveryKeyHash: "***", // Don't store actual hashes
		NewRecoveryKeyHash: "***",
		RotatedAt:          currentTime,
		RotatedBy:          user.ID,
		Reason:             "Password recovery completed",
	}

	if err := uc.recoveryRepo.CreateRecoveryKeyRotation(ctx, rotation); err != nil {
		uc.logger.Warn("Failed to record key rotation", zap.Error(err))
		// Continue anyway
	}

	// Delete recovery session
	if err := uc.recoveryRepo.DeleteRecoverySession(ctx, sessionID); err != nil {
		uc.logger.Warn("Failed to delete recovery session", zap.Error(err))
		// Continue anyway
	}

	uc.logger.Info("Recovery completed successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	return &CompleteRecoveryResult{
		Success: true,
		Message: "Account recovery completed successfully. You can now log in with your new password.",
	}, nil
}
