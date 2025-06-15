package auth

import (
	"time"
)

// TokenEncryptionService defines the interface for token encryption
type TokenEncryptionService interface {
	// EncryptTokens encrypts both access and refresh tokens with user's public key
	EncryptTokens(accessToken, refreshToken string, publicKey []byte, accessExpiry, refreshExpiry time.Time) (*EncryptedTokenResponse, error)
}
