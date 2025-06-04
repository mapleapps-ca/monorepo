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
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryTime  time.Time `json:"access_token_expiry_time"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`
}

// CompleteLoginRepository defines methods for the login completion process
type CompleteLoginDTORepository interface {
	CompleteLogin(ctx context.Context, request *CompleteLoginRequestDTO) (*TokenResponseDTO, error)
}
