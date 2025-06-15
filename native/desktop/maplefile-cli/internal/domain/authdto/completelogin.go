// monorepo/native/desktop/maplefile-cli/internal/domain/authdto/completelogin.go
package authdto

import (
	"context"
	"time"
)

// CompleteLoginRequestDTO represents the data sent to the server to complete login
type CompleteLoginRequestDTO struct {
	Email         string `json:"email"`
	ChallengeID   string `json:"challengeId"`
	DecryptedData string `json:"decryptedData"`
}

// TokenResponseDTO represents the response from the server with auth tokens
type TokenResponseDTO struct {
	// Legacy plaintext fields
	AccessToken            string    `json:"access_token,omitempty"`
	AccessTokenExpiryTime  time.Time `json:"access_token_expiry_time"`
	RefreshToken           string    `json:"refresh_token,omitempty"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`

	// New encrypted token fields
	EncryptedTokens string `json:"encrypted_tokens,omitempty"`
	TokenNonce      string `json:"token_nonce,omitempty"`
}

// CompleteLoginRepository defines methods for the login completion process
type CompleteLoginDTORepository interface {
	CompleteLogin(ctx context.Context, request *CompleteLoginRequestDTO) (*TokenResponseDTO, error)
}
