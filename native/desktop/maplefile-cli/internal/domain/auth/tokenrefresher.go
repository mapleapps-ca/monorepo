// internal/domain/auth/tokenrefresher.go
package auth

import (
	"context"
	"time"
)

// TokenRefreshResponse represents the response from a token refresh operation
type TokenRefreshResponse struct {
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryDate  time.Time `json:"access_token_expiry_date"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryDate time.Time `json:"refresh_token_expiry_date"`
}

// TokenRefresher defines the interface for token refresh operations
type TokenRefresher interface {
	RefreshToken(ctx context.Context, refreshToken string) (*TokenRefreshResponse, error)
}
