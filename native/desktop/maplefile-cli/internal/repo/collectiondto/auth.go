// monorepo/native/desktop/maplefile-cli/internal/repo/collectiondto/auth.go
package collectiondto

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// getAccessToken retrieves a valid access token for API calls
func (r *collectionDTORepository) getAccessToken(ctx context.Context) (string, error) {
	// Get authenticated user's email
	email, err := r.configService.GetEmail(ctx)
	if err != nil {
		r.logger.Error("Failed to get authenticated user email", zap.Error(err))
		return "", errors.NewAppError("failed to get authenticated user", err)
	}

	// Get user data to retrieve auth token
	userData, err := r.userRepo.GetByEmail(ctx, email)
	if err != nil {
		r.logger.Error("Failed to retrieve user data", zap.String("email", email), zap.Error(err))
		return "", errors.NewAppError("failed to retrieve user data", err)
	}

	if userData == nil {
		r.logger.Error("User data not found", zap.String("email", email))
		return "", errors.NewAppError("user not found; please login first", nil)
	}

	// Check if access token is valid
	if userData.AccessToken == "" || r.isTokenExpiredOrExpiringSoon(userData.AccessTokenExpiryTime) {
		r.logger.Info("Access token is invalid or expiring soon, refreshing",
			zap.String("email", email))

		if err := r.refreshTokenIfNeeded(ctx, userData); err != nil {
			r.logger.Error("Failed to refresh token", zap.Error(err))
			return "", errors.NewAppError("authentication token has expired; please refresh token", nil)
		}
	}

	return userData.AccessToken, nil
}

// Helper to check if token is expired or expiring soon
func (r *collectionDTORepository) isTokenExpiredOrExpiringSoon(expiryTime time.Time) bool {
	// Create this helper method to make token checks more readable
	return time.Now().After(expiryTime)
}

// refreshTokenIfNeeded checks if the access token is expired and refreshes it if needed
func (r *collectionDTORepository) refreshTokenIfNeeded(ctx context.Context, userData *user.User) error {
	// Check if token is expired or will expire soon (within 30 seconds as a buffer)
	if userData.AccessToken == "" || time.Now().Add(30*time.Second).After(userData.AccessTokenExpiryTime) {
		r.logger.Info("Access token expired or will expire soon, attempting to refresh",
			zap.String("email", userData.Email),
			zap.Time("tokenExpiry", userData.AccessTokenExpiryTime))

		// Check if we have a refresh token
		if userData.RefreshToken == "" {
			return errors.NewAppError("no refresh token available", nil)
		}

		// Check if refresh token is still valid
		if time.Now().After(userData.RefreshTokenExpiryTime) {
			return errors.NewAppError("refresh token has expired, please login again", nil)
		}

		// Refresh the token using the domain interface
		tokenResp, err := r.tokenRefresher.RefreshToken(ctx, userData.RefreshToken)
		if err != nil {
			return errors.NewAppError("failed to refresh token", err)
		}

		// Update user with new tokens
		if err := r.userRepo.OpenTransaction(); err != nil {
			return errors.NewAppError("failed to open transaction for token update", err)
		}

		userData.AccessToken = tokenResp.AccessToken
		userData.AccessTokenExpiryTime = tokenResp.AccessTokenExpiryDate
		userData.RefreshToken = tokenResp.RefreshToken
		userData.RefreshTokenExpiryTime = tokenResp.RefreshTokenExpiryDate
		userData.ModifiedAt = time.Now()

		if err := r.userRepo.UpsertByEmail(ctx, userData); err != nil {
			r.userRepo.DiscardTransaction()
			return errors.NewAppError("failed to update user with new tokens", err)
		}

		if err := r.userRepo.CommitTransaction(); err != nil {
			r.userRepo.DiscardTransaction()
			return errors.NewAppError("failed to commit token update transaction", err)
		}

		r.logger.Info("Successfully refreshed access token",
			zap.String("email", userData.Email),
			zap.Time("newTokenExpiry", userData.AccessTokenExpiryTime))
	}

	return nil
}
