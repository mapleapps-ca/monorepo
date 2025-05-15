// monorepo/native/desktop/papercloud-cli/internal/domain/auth/completelogin.go
package auth

import (
	"context"
	"time"
)

// CompleteLoginRequest represents the data sent to the server to complete login
type CompleteLoginRequest struct {
	Email         string `json:"email"`
	ChallengeID   string `json:"challengeId"`
	DecryptedData string `json:"decryptedData"`
}

// TokenResponse represents the response from the server with auth tokens
type TokenResponse struct {
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryTime  time.Time `json:"access_token_expiry_time"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`
}

// CompleteLoginRepository defines methods for the login completion process
type CompleteLoginRepository interface {
	CompleteLogin(ctx context.Context, request *CompleteLoginRequest) (*TokenResponse, error)
}

// CryptoService defines cryptographic operations needed for login
type CryptoService interface {
	DeriveKeyFromPassword(password string, salt []byte) ([]byte, error)
	DecryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error)
	DecryptWithBox(encryptedData, publicKey, privateKey []byte) ([]byte, error)
	EncodeToBase64(data []byte) string
}
