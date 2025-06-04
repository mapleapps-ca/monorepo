// monorepo/native/desktop/maplefile-cli/internal/usecase/authdto/recovery.go
package authdto

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// RecoveryUseCase defines the interface for recovery use cases
type RecoveryUseCase interface {
	// InitiateRecovery starts the recovery process and returns decrypted keys
	InitiateRecovery(ctx context.Context, email, recoveryKey string) (*RecoveryData, error)

	// CompleteRecovery sets new password and completes recovery
	CompleteRecovery(ctx context.Context, recoveryData *RecoveryData, newPassword string) (*dom_authdto.RecoveryCompleteResponse, *user.User, error)
}

// RecoveryData holds decrypted data during recovery process
type RecoveryData struct {
	SessionID  string
	Email      string
	MasterKey  []byte
	PrivateKey []byte
	PublicKey  []byte
	Salt       []byte // Original salt
	KDFParams  keys.KDFParams
	ExpiresAt  time.Time
}

// recoveryUseCase implements the RecoveryUseCase interface
type recoveryUseCase struct {
	logger          *zap.Logger
	recoveryRepo    dom_authdto.RecoveryRepository
	userRepo        user.Repository
	tokenRepository dom_authdto.TokenRepository
}

// NewRecoveryUseCase creates a new recovery use case
func NewRecoveryUseCase(
	logger *zap.Logger,
	recoveryRepo dom_authdto.RecoveryRepository,
	userRepo user.Repository,
	tokenRepository dom_authdto.TokenRepository,
) RecoveryUseCase {
	logger = logger.Named("RecoveryUseCase")
	return &recoveryUseCase{
		logger:          logger,
		recoveryRepo:    recoveryRepo,
		userRepo:        userRepo,
		tokenRepository: tokenRepository,
	}
}

// InitiateRecovery starts the recovery process
func (uc *recoveryUseCase) InitiateRecovery(ctx context.Context, email, recoveryKeyStr string) (*RecoveryData, error) {
	// Validate inputs
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}
	if recoveryKeyStr == "" {
		return nil, errors.NewAppError("recovery key is required", nil)
	}

	// Sanitize inputs
	email = strings.ToLower(strings.TrimSpace(email))
	recoveryKeyStr = strings.TrimSpace(recoveryKeyStr)

	// Decode recovery key from base64
	recoveryKey, err := base64.StdEncoding.DecodeString(recoveryKeyStr)
	if err != nil {
		// Try URL-safe base64 as fallback
		recoveryKey, err = base64.RawURLEncoding.DecodeString(recoveryKeyStr)
		if err != nil {
			return nil, errors.NewAppError("invalid recovery key format", err)
		}
	}

	// Validate recovery key length
	if len(recoveryKey) != crypto.RecoveryKeySize {
		return nil, errors.NewAppError(fmt.Sprintf("invalid recovery key size: expected %d bytes, got %d", crypto.RecoveryKeySize, len(recoveryKey)), nil)
	}

	uc.logger.Debug("🔐 Initiating recovery process", zap.String("email", email))

	// Create recovery request
	request := &dom_authdto.RecoveryRequest{
		Email:       email,
		RecoveryKey: base64.RawURLEncoding.EncodeToString(recoveryKey),
	}

	// Call repository to initiate recovery
	response, err := uc.recoveryRepo.InitiateRecovery(ctx, request)
	if err != nil {
		return nil, err
	}

	// Decode encrypted master key (encrypted with recovery key)
	encMasterKeyBytes, err := base64.StdEncoding.DecodeString(response.MasterKeyEncryptedWithRecoveryKey)
	if err != nil {
		encMasterKeyBytes, err = base64.RawURLEncoding.DecodeString(response.MasterKeyEncryptedWithRecoveryKey)
		if err != nil {
			return nil, errors.NewAppError("failed to decode encrypted master key", err)
		}
	}

	// Split nonce and ciphertext for master key
	if len(encMasterKeyBytes) < crypto.ChaCha20Poly1305NonceSize {
		return nil, errors.NewAppError("encrypted master key data too short", nil)
	}

	nonce := encMasterKeyBytes[:crypto.ChaCha20Poly1305NonceSize]
	ciphertext := encMasterKeyBytes[crypto.ChaCha20Poly1305NonceSize:]

	// Decrypt master key using recovery key
	masterKey, err := crypto.DecryptWithSecretBox(ciphertext, nonce, recoveryKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt master key with recovery key", err)
	}

	// Decode encrypted private key
	encPrivateKeyBytes, err := base64.StdEncoding.DecodeString(response.EncryptedPrivateKey)
	if err != nil {
		encPrivateKeyBytes, err = base64.RawURLEncoding.DecodeString(response.EncryptedPrivateKey)
		if err != nil {
			return nil, errors.NewAppError("failed to decode encrypted private key", err)
		}
	}

	// Split nonce and ciphertext for private key
	if len(encPrivateKeyBytes) < crypto.ChaCha20Poly1305NonceSize {
		return nil, errors.NewAppError("encrypted private key data too short", nil)
	}

	privKeyNonce := encPrivateKeyBytes[:crypto.ChaCha20Poly1305NonceSize]
	privKeyCiphertext := encPrivateKeyBytes[crypto.ChaCha20Poly1305NonceSize:]

	// Decrypt private key using master key
	privateKey, err := crypto.DecryptWithSecretBox(privKeyCiphertext, privKeyNonce, masterKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt private key", err)
	}

	// Decode public key
	publicKey, err := base64.StdEncoding.DecodeString(response.PublicKey)
	if err != nil {
		publicKey, err = base64.RawURLEncoding.DecodeString(response.PublicKey)
		if err != nil {
			return nil, errors.NewAppError("failed to decode public key", err)
		}
	}

	// Decode salt
	salt, err := base64.StdEncoding.DecodeString(response.Salt)
	if err != nil {
		salt, err = base64.RawURLEncoding.DecodeString(response.Salt)
		if err != nil {
			return nil, errors.NewAppError("failed to decode salt", err)
		}
	}

	// Parse KDF params
	var kdfParams keys.KDFParams
	if err := json.Unmarshal([]byte(response.KDFParams), &kdfParams); err != nil {
		// Use default if parsing fails
		kdfParams = keys.DefaultKDFParams()
	}

	uc.logger.Info("✅ Recovery keys successfully decrypted", zap.String("sessionID", response.SessionID))

	return &RecoveryData{
		SessionID:  response.SessionID,
		Email:      email,
		MasterKey:  masterKey,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Salt:       salt,
		KDFParams:  kdfParams,
		ExpiresAt:  response.ExpiresAt,
	}, nil
}

// CompleteRecovery sets new password and completes recovery
func (uc *recoveryUseCase) CompleteRecovery(ctx context.Context, recoveryData *RecoveryData, newPassword string) (*dom_authdto.RecoveryCompleteResponse, *user.User, error) {
	// Validate inputs
	if recoveryData == nil {
		return nil, nil, errors.NewAppError("recovery data is required", nil)
	}
	if newPassword == "" {
		return nil, nil, errors.NewAppError("new password is required", nil)
	}

	// Check if recovery session has expired
	if time.Now().After(recoveryData.ExpiresAt) {
		return nil, nil, errors.NewAppError("recovery session has expired", nil)
	}

	uc.logger.Debug("🔐 Completing recovery with new password", zap.String("email", recoveryData.Email))

	// Generate new salt for password
	newSalt, err := crypto.GenerateRandomBytes(crypto.Argon2SaltSize)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to generate new salt", err)
	}

	// Derive new key encryption key from new password
	newKeyEncryptionKey, err := crypto.DeriveKeyFromPassword(newPassword, newSalt)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to derive key from new password", err)
	}

	// Re-encrypt master key with new key encryption key
	encryptedMasterKey, err := crypto.EncryptWithSecretBox(recoveryData.MasterKey, newKeyEncryptionKey)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to encrypt master key with new password", err)
	}

	// Re-encrypt private key with master key (this doesn't change, but we need to send it)
	encryptedPrivateKey, err := crypto.EncryptWithSecretBox(recoveryData.PrivateKey, recoveryData.MasterKey)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to encrypt private key", err)
	}

	// Combine nonce and ciphertext for master key
	encMasterKeyBytes := append(encryptedMasterKey.Nonce, encryptedMasterKey.Ciphertext...)
	encPrivateKeyBytes := append(encryptedPrivateKey.Nonce, encryptedPrivateKey.Ciphertext...)

	// Create complete recovery request
	completeRequest := &dom_authdto.RecoveryCompleteRequest{
		Email:               recoveryData.Email,
		SessionID:           recoveryData.SessionID,
		NewPassword:         newPassword,
		EncryptedMasterKey:  base64.RawURLEncoding.EncodeToString(encMasterKeyBytes),
		EncryptedPrivateKey: base64.RawURLEncoding.EncodeToString(encPrivateKeyBytes),
		Salt:                base64.RawURLEncoding.EncodeToString(newSalt),
	}

	// Call repository to complete recovery
	response, err := uc.recoveryRepo.CompleteRecovery(ctx, completeRequest)
	if err != nil {
		return nil, nil, err
	}

	// Get or create user record
	existingUser, err := uc.userRepo.GetByEmail(ctx, recoveryData.Email)
	if err != nil || existingUser == nil {
		// Create a new user record if not found
		existingUser = &user.User{
			Email:     recoveryData.Email,
			Status:    user.UserStatusActive,
			CreatedAt: time.Now(),
		}
	}

	// Update user with new encryption data
	currentTime := time.Now()
	existingUser.PasswordSalt = newSalt
	existingUser.EncryptedMasterKey = keys.EncryptedMasterKey{
		Ciphertext: encryptedMasterKey.Ciphertext,
		Nonce:      encryptedMasterKey.Nonce,
		KeyVersion: existingUser.EncryptedMasterKey.KeyVersion + 1,
		RotatedAt:  &currentTime,
	}
	existingUser.PublicKey.Key = recoveryData.PublicKey
	existingUser.EncryptedPrivateKey = keys.EncryptedPrivateKey{
		Ciphertext: encryptedPrivateKey.Ciphertext,
		Nonce:      encryptedPrivateKey.Nonce,
	}
	existingUser.LastPasswordChange = currentTime
	existingUser.ModifiedAt = currentTime
	existingUser.LastLoginAt = currentTime

	// Save tokens if provided
	if response.AccessToken != "" && response.RefreshToken != "" {
		uc.tokenRepository.Save(
			ctx,
			recoveryData.Email,
			response.AccessToken,
			&response.AccessTokenExpiryTime,
			response.RefreshToken,
			&response.RefreshTokenExpiryTime,
		)
	}

	uc.logger.Info("✅ Account recovery completed successfully", zap.String("email", recoveryData.Email))

	return response, existingUser, nil
}
