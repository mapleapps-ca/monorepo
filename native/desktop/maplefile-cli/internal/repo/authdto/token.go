// native/desktop/maplefile-cli/internal/repo/authdto/token.go
package authdto

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
)

// tokenDTORepositoryImpl implements the TokenDTORepository interface
type tokenDTORepositoryImpl struct {
	logger        *zap.Logger
	configService config.ConfigService
}

// NewTokenDTORepository creates a new instance of TokenDTORepository
func NewTokenDTORepository(
	logger *zap.Logger,
	configService config.ConfigService,
) dom_authdto.TokenDTORepository {
	logger = logger.Named("TokenRepository")
	return &tokenDTORepositoryImpl{
		logger:        logger,
		configService: configService,
	}
}

func (s *tokenDTORepositoryImpl) Save(
	ctx context.Context,
	email string,
	accessToken string,
	accessTokenExpiryDate *time.Time,
	refreshToken string,
	refreshTokenExpiryDate *time.Time,
) error {
	return s.configService.SetLoggedInUserCredentials(ctx, email, accessToken, accessTokenExpiryDate, refreshToken, refreshTokenExpiryDate)
}

func (s *tokenDTORepositoryImpl) GetAccessToken(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// Check if token is expired or will expire soon (within 30 seconds as a buffer)
	if creds.AccessToken == "" || time.Now().Add(30*time.Second).After(*creds.AccessTokenExpiryTime) {
		// Check if we have a refresh token
		if creds.RefreshToken == "" {
			return "", errors.NewAppError("no refresh token available", nil)
		}

		// Check if refresh token is still valid
		if time.Now().After(*creds.RefreshTokenExpiryTime) {
			return "", errors.NewAppError("refresh token has expired, please login again", nil)
		}

		// Since we only support encrypted tokens now, always require service layer for refresh
		return "", errors.NewAppError("access token expired - encrypted token refresh required at service layer with password", nil)
	}

	return creds.AccessToken, nil
}

func (s *tokenDTORepositoryImpl) GetAccessTokenAfterForcedRefresh(ctx context.Context) (string, error) {
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting logged in user credentials: %w", err)
	}
	if creds == nil {
		return "", fmt.Errorf("no logged in user credentials found")
	}

	// For forced refresh, always delegate to service layer since all tokens are encrypted
	return "", errors.NewAppError("forced refresh requires service layer with password for encrypted token decryption", nil)
}
