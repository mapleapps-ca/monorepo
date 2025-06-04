// monorepo/native/desktop/maplefile-cli/internal/domain/authdto/completelogin.go
package authdto

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
