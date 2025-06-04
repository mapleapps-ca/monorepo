// internal/domain/authdto/token.go
package authdto

import (
	"context"
	"time"
)

// TokenDTO represents the access credentials to the authenticated user in the cloud service.
type TokenDTO struct {
	Email                  string     `json:"email"`
	AccessToken            string     `json:"access_token"`
	AccessTokenExpiryDate  *time.Time `json:"access_token_expiry_date"`
	RefreshToken           string     `json:"refresh_token"`
	RefreshTokenExpiryDate *time.Time `json:"refresh_token_expiry_date"`
}

// TokenDTORepository defines the interface saving, getting and refreshing access tokens for the authenticated user
type TokenDTORepository interface {
	Save(
		ctx context.Context,
		email string,
		accessToken string,
		accessTokenExpiryDate *time.Time,
		refreshToken string,
		refreshTokenExpiryDate *time.Time,
	) error
	GetAccessToken(ctx context.Context) (string, error)
	GetAccessTokenAfterForcedRefresh(ctx context.Context) (string, error)
}
