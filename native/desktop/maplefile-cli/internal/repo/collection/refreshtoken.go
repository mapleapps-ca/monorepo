// monorepo/native/desktop/maplefile-cli/internal/repo/auth/refreshtoken.go (DEPRECATED)
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// refreshTokenIfNeeded checks if the access token is expired and refreshes it if needed
func (r *collectionRepository) refreshTokenIfNeeded(ctx context.Context, userData *user.User) error {
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
