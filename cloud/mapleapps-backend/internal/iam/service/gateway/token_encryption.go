// cloud/mapleapps-backend/internal/iam/service/gateway/token_encryption.go
package gateway

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/nacl/box"

	dom_auth "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/auth"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/security/crypto"
)

// tokenEncryptionService implements TokenEncryptionService
type tokenEncryptionService struct {
	logger *zap.Logger
}

// NewTokenEncryptionService creates a new token encryption service
func NewTokenEncryptionService(logger *zap.Logger) dom_auth.TokenEncryptionService {
	logger = logger.Named("TokenEncryptionService")
	return &tokenEncryptionService{
		logger: logger,
	}
}

// EncryptTokens encrypts access and refresh tokens separately with the user's public key
func (s *tokenEncryptionService) EncryptTokens(
	accessToken, refreshToken string,
	publicKey []byte,
	accessExpiry, refreshExpiry time.Time,
) (*dom_auth.EncryptedTokenResponse, error) {

	if len(publicKey) != crypto.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", crypto.PublicKeySize, len(publicKey))
	}

	s.logger.Debug("Encrypting tokens separately for user",
		zap.Int("access_token_size", len(accessToken)),
		zap.Int("refresh_token_size", len(refreshToken)),
		zap.String("public_key_preview", base64.StdEncoding.EncodeToString(publicKey[:8])))

	// Generate ephemeral keypair for this encryption
	ephemeralPubKey, ephemeralPrivKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral keypair: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, 24) // box.NonceSize
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Convert public key to array
	var recipientPubKey [32]byte
	copy(recipientPubKey[:], publicKey)

	// Create nonce array
	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Encrypt access token using box.Seal
	encryptedAccessToken := box.Seal(nil, []byte(accessToken), &nonceArray, &recipientPubKey, ephemeralPrivKey)

	// Encrypt refresh token using box.Seal (reuse same nonce and keys)
	encryptedRefreshToken := box.Seal(nil, []byte(refreshToken), &nonceArray, &recipientPubKey, ephemeralPrivKey)

	// Combine ephemeral public key + nonce + encrypted data for each token
	// Format: [32 bytes ephemeral pubkey][24 bytes nonce][encrypted data]
	combinedAccessToken := make([]byte, 32+24+len(encryptedAccessToken))
	copy(combinedAccessToken[:32], ephemeralPubKey[:])
	copy(combinedAccessToken[32:56], nonce)
	copy(combinedAccessToken[56:], encryptedAccessToken)

	combinedRefreshToken := make([]byte, 32+24+len(encryptedRefreshToken))
	copy(combinedRefreshToken[:32], ephemeralPubKey[:])
	copy(combinedRefreshToken[32:56], nonce)
	copy(combinedRefreshToken[56:], encryptedRefreshToken)

	s.logger.Info("Successfully encrypted tokens separately",
		zap.Int("encrypted_access_token_size", len(combinedAccessToken)),
		zap.Int("encrypted_refresh_token_size", len(combinedRefreshToken)),
		zap.Time("access_expiry", accessExpiry),
		zap.Time("refresh_expiry", refreshExpiry))

	// Return encrypted response with separate tokens
	return &dom_auth.EncryptedTokenResponse{
		EncryptedAccessToken:   base64.StdEncoding.EncodeToString(combinedAccessToken),
		EncryptedRefreshToken:  base64.StdEncoding.EncodeToString(combinedRefreshToken),
		AccessTokenExpiryTime:  accessExpiry,
		RefreshTokenExpiryTime: refreshExpiry,
		Nonce:                  base64.StdEncoding.EncodeToString(nonce),
	}, nil
}
