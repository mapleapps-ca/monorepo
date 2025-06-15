package auth

import (
	"time"
)

// OneTimeToken for email verification
type OneTimeToken struct {
	Token     string    `json:"token" bson:"token"`
	Email     string    `json:"email" bson:"email"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
	Used      bool      `json:"used" bson:"used"`
}

// AuthToken for API authentication
type AuthToken struct {
	Token     string    `json:"token" bson:"token"`
	UserID    string    `json:"user_id" bson:"user_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
}

// EncryptedAuthToken is the auth token encrypted with the user's public key
type EncryptedAuthToken struct {
	Ciphertext []byte `json:"ciphertext" bson:"ciphertext"`
	UserID     string `json:"user_id" bson:"user_id"`
}

// EncryptedTokenResponse represents tokens encrypted with user's public key
type EncryptedTokenResponse struct {
	// Base64 encoded encrypted access token
	EncryptedAccessToken string `json:"encrypted_access_token"`
	// Base64 encoded encrypted refresh token
	EncryptedRefreshToken string `json:"encrypted_refresh_token"`
	// Expiry times are sent in plaintext as they're not sensitive
	AccessTokenExpiryTime  time.Time `json:"access_token_expiry_time"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`
	// Base64 encoded nonce used for encryption
	Nonce string `json:"nonce"`
}
