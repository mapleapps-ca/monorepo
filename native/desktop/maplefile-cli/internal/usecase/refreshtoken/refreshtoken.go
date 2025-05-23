// native/desktop/maplefile-cli/internal/usecase/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/tokenservice"
)

// RefreshTokenUseCase defines the interface for refreshing tokens
type RefreshTokenUseCase interface {
	Execute(ctx context.Context) error
}

// refreshTokenUseCaseImpl implements the RefreshTokenUseCase interface
type refreshTokenUseCaseImpl struct {
	logger          *zap.Logger
	configService   config.ConfigService
	userRepo        user.Repository
	tokenRefreshSvc tokenservice.TokenRefreshService
}

// NewRefreshTokenUseCase creates a new instance of RefreshTokenUseCase
func NewRefreshTokenUseCase(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenRefreshSvc tokenservice.TokenRefreshService,
) RefreshTokenUseCase {
	return &refreshTokenUseCaseImpl{
		logger:          logger,
		configService:   configService,
		userRepo:        userRepo,
		tokenRefreshSvc: tokenRefreshSvc,
	}
}

// Execute performs the token refresh operation
func (uc *refreshTokenUseCaseImpl) Execute(ctx context.Context) error {
	// Get the current user's email from configuration
	email, err := uc.configService.GetLoggedInUserEmail(ctx)
	if err != nil {
		return fmt.Errorf("error getting authenticated user email: %w", err)
	}

	if email == "" {
		return fmt.Errorf("no authenticated user found")
	}

	// Get the user details from the repository
	userData, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("error retrieving user data: %w", err)
	}

	if userData == nil {
		return fmt.Errorf("user with email %s not found", email)
	}

	// Check if the user has a refresh token
	if userData.RefreshToken == "" {
		return fmt.Errorf("no refresh token found")
	}

	// Check if the refresh token has expired
	if time.Now().After(userData.RefreshTokenExpiryTime) {
		return fmt.Errorf("refresh token has expired")
	}

	// Call the service to refresh the token
	tokenResponse, err := uc.tokenRefreshSvc.RefreshToken(ctx, userData.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Start a transaction to update the user with the new tokens
	if err := uc.userRepo.OpenTransaction(); err != nil {
		return fmt.Errorf("error opening transaction: %w", err)
	}

	// Update user with the new tokens
	userData.AccessToken = tokenResponse.AccessToken
	userData.AccessTokenExpiryTime = tokenResponse.AccessTokenExpiryDate
	userData.RefreshToken = tokenResponse.RefreshToken
	userData.RefreshTokenExpiryTime = tokenResponse.RefreshTokenExpiryDate
	userData.ModifiedAt = time.Now()

	// Update the user in the repository
	if err := uc.userRepo.UpsertByEmail(ctx, userData); err != nil {
		uc.userRepo.DiscardTransaction()
		return fmt.Errorf("error updating user with new tokens: %w", err)
	}

	// Commit the transaction
	if err := uc.userRepo.CommitTransaction(); err != nil {
		uc.userRepo.DiscardTransaction()
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
