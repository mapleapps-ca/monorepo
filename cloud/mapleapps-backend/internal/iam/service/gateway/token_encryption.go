// cloud/mapleapps-backend/internal/iam/service/gateway/token_encryption.go
package gateway

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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

// TokenPayload represents the structured token data to encrypt
type TokenPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// EncryptTokens encrypts both tokens with the user's public key
func (s *tokenEncryptionService) EncryptTokens(
	accessToken, refreshToken string,
	publicKey []byte,
	accessExpiry, refreshExpiry time.Time,
) (*dom_auth.EncryptedTokenResponse, error) {

	if len(publicKey) != crypto.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", crypto.PublicKeySize, len(publicKey))
	}

	// Create token payload
	payload := TokenPayload{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	// Marshal to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token payload: %w", err)
	}

	s.logger.Debug("Encrypting tokens for user",
		zap.Int("payload_size", len(payloadBytes)),
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

	// Encrypt the payload using box.Seal
	encrypted := box.Seal(nil, payloadBytes, &nonceArray, &recipientPubKey, ephemeralPrivKey)

	// Combine ephemeral public key + nonce + encrypted data
	// Format: [32 bytes ephemeral pubkey][24 bytes nonce][encrypted data]
	combined := make([]byte, 32+24+len(encrypted))
	copy(combined[:32], ephemeralPubKey[:])
	copy(combined[32:56], nonce)
	copy(combined[56:], encrypted)

	s.logger.Info("Successfully encrypted tokens",
		zap.Int("encrypted_size", len(combined)),
		zap.Time("access_expiry", accessExpiry),
		zap.Time("refresh_expiry", refreshExpiry))

	// Return encrypted response
	return &dom_auth.EncryptedTokenResponse{
		EncryptedAccessToken:   base64.StdEncoding.EncodeToString(combined),
		EncryptedRefreshToken:  "", // Both tokens are in the same encrypted payload
		AccessTokenExpiryTime:  accessExpiry,
		RefreshTokenExpiryTime: refreshExpiry,
		Nonce:                  base64.StdEncoding.EncodeToString(nonce),
	}, nil
}
