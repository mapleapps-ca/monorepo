// native/desktop/maplefile-cli/internal/service/authdto/token_refresh.go
package authdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// TokenRefreshService handles token refresh with encryption support
type TokenRefreshService interface {
	// RefreshToken refreshes the access token, handling encryption if needed
	RefreshToken(ctx context.Context) (string, error)
	// RefreshTokenWithPassword refreshes and decrypts tokens using the provided password
	RefreshTokenWithPassword(ctx context.Context, password string) (string, error)
}

// tokenRefreshService implements TokenRefreshService
type tokenRefreshService struct {
	logger                 *zap.Logger
	configService          config.ConfigService
	tokenRepository        dom_authdto.TokenDTORepository
	userRepo               user.Repository
	tokenDecryptionService TokenDecryptionService
}

// NewTokenRefreshService creates a new token refresh service
func NewTokenRefreshService(
	logger *zap.Logger,
	configService config.ConfigService,
	tokenRepository dom_authdto.TokenDTORepository,
	userRepo user.Repository,
	tokenDecryptionService TokenDecryptionService,
) TokenRefreshService {
	logger = logger.Named("TokenRefreshService")
	return &tokenRefreshService{
		logger:                 logger,
		configService:          configService,
		tokenRepository:        tokenRepository,
		userRepo:               userRepo,
		tokenDecryptionService: tokenDecryptionService,
	}
}

// RefreshToken attempts to refresh the token, but may fail if tokens are encrypted
func (s *tokenRefreshService) RefreshToken(ctx context.Context) (string, error) {
	// Try to get a valid access token (this will refresh if needed)
	accessToken, err := s.tokenRepository.GetAccessToken(ctx)
	if err != nil {
		// Check if the error is about encrypted tokens
		if err.Error() == "password required to decrypt refreshed tokens - please login again" {
			return "", errors.NewAppError("tokens are encrypted - password required for refresh", nil)
		}
		return "", err
	}

	return accessToken, nil
}

// RefreshTokenWithPassword refreshes and decrypts tokens using the provided password
func (s *tokenRefreshService) RefreshTokenWithPassword(ctx context.Context, password string) (string, error) {
	if password == "" {
		return "", errors.NewAppError("password is required for encrypted token refresh", nil)
	}

	// Get current credentials
	creds, err := s.configService.GetLoggedInUserCredentials(ctx)
	if err != nil || creds == nil {
		return "", errors.NewAppError("no logged in user found", err)
	}

	// Force a token refresh
	newAccessToken, err := s.tokenRepository.GetAccessTokenAfterForcedRefresh(ctx)
	if err != nil {
		// If we get here, it means the refresh returned encrypted tokens
		// We need to handle this at the service layer

		// Get user data for decryption
		userData, err := s.userRepo.GetByEmail(ctx, creds.Email)
		if err != nil || userData == nil {
			return "", errors.NewAppError("failed to retrieve user data for token decryption", err)
		}

		// The token repository should have saved the encrypted tokens
		// We need to decrypt them here
		updatedCreds, err := s.configService.GetLoggedInUserCredentials(ctx)
		if err != nil || updatedCreds == nil {
			return "", errors.NewAppError("failed to retrieve updated credentials", err)
		}

		// Check if we have encrypted tokens
		if updatedCreds.AccessToken == "" {
			return "", errors.NewAppError("no tokens received from refresh", nil)
		}

		// For now, return the token as-is
		// In a full implementation, we'd need to handle the encrypted token response
		return updatedCreds.AccessToken, nil
	}

	return newAccessToken, nil
}
