// native/desktop/maplefile-cli/internal/service/authdto/token_decryption.go
package authdto

import (
	"encoding/base64"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// TokenDecryptionService handles decryption of encrypted authentication tokens
type TokenDecryptionService interface {
	// DecryptTokens decrypts the encrypted token payload using user's private key
	DecryptTokens(encryptedTokens string, user *user.User, password string) (accessToken, refreshToken string, err error)
}

// tokenDecryptionService implements TokenDecryptionService
type tokenDecryptionService struct {
	logger *zap.Logger
}

// NewTokenDecryptionService creates a new token decryption service
func NewTokenDecryptionService(logger *zap.Logger) TokenDecryptionService {
	logger = logger.Named("TokenDecryptionService")
	return &tokenDecryptionService{
		logger: logger,
	}
}

// TokenPayload represents the decrypted token data
type TokenPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// DecryptTokens decrypts the encrypted token payload
func (s *tokenDecryptionService) DecryptTokens(encryptedTokens string, user *user.User, password string) (string, string, error) {
	if encryptedTokens == "" {
		return "", "", errors.NewAppError("encrypted tokens are required", nil)
	}

	if user == nil {
		return "", "", errors.NewAppError("user data is required", nil)
	}

	if password == "" {
		return "", "", errors.NewAppError("password is required for decryption", nil)
	}

	s.logger.Debug("Starting token decryption",
		zap.String("email", user.Email),
		zap.Int("encrypted_length", len(encryptedTokens)))

	// Decode the encrypted tokens from base64
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedTokens)
	if err != nil {
		// Try URL-safe encoding as fallback
		encryptedData, err = base64.RawURLEncoding.DecodeString(encryptedTokens)
		if err != nil {
			return "", "", errors.NewAppError("failed to decode encrypted tokens", err)
		}
	}

	// Validate minimum size: 32 bytes ephemeral pubkey + 24 bytes nonce + overhead
	if len(encryptedData) < 32+24+16 {
		return "", "", errors.NewAppError("encrypted data too short", nil)
	}

	// Extract components
	ephemeralPublicKey := encryptedData[:32]
	nonce := encryptedData[32:56]
	ciphertext := encryptedData[56:]

	s.logger.Debug("Extracted encryption components",
		zap.Int("ephemeral_key_size", len(ephemeralPublicKey)),
		zap.Int("nonce_size", len(nonce)),
		zap.Int("ciphertext_size", len(ciphertext)))

	// First, we need to decrypt the user's private key
	// Derive key encryption key from password
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return "", "", errors.NewAppError("failed to derive key from password", err)
	}

	// Decrypt master key
	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return "", "", errors.NewAppError("failed to decrypt master key", err)
	}

	// Decrypt private key using master key
	privateKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedPrivateKey.Ciphertext,
		user.EncryptedPrivateKey.Nonce,
		masterKey,
	)
	if err != nil {
		// Clear sensitive data
		crypto.ClearBytes(masterKey)
		return "", "", errors.NewAppError("failed to decrypt private key", err)
	}

	// Clear master key as we no longer need it
	crypto.ClearBytes(masterKey)

	// Now decrypt the tokens using the ephemeral public key and our private key
	plaintext, err := crypto.DecryptWithBox(ciphertext, nonce, ephemeralPublicKey, privateKey)
	if err != nil {
		// Clear sensitive data
		crypto.ClearBytes(privateKey)
		return "", "", errors.NewAppError("failed to decrypt tokens", err)
	}

	// Clear private key
	crypto.ClearBytes(privateKey)

	// Parse the JSON payload
	var payload TokenPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		// Clear sensitive data
		crypto.ClearBytes(plaintext)
		return "", "", errors.NewAppError("failed to parse decrypted token payload", err)
	}

	// Clear plaintext
	crypto.ClearBytes(plaintext)

	s.logger.Info("Successfully decrypted tokens",
		zap.String("email", user.Email),
		zap.Bool("has_access_token", payload.AccessToken != ""),
		zap.Bool("has_refresh_token", payload.RefreshToken != ""))

	return payload.AccessToken, payload.RefreshToken, nil
}
