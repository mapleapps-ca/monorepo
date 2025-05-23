// internal/usecase/user/get_by_email.go
package user

import (
	"context"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
)

// GetByIsLoggedInUseCase defines the interface for retrieving a user by whom is logged in locally
type GetByIsLoggedInUseCase interface {
	Execute(ctx context.Context) (*user.User, error)
}

type getByIsLoggedInUseCaseImpl struct {
	configService config.ConfigService
	userRepo      user.Repository
}

// NewGetByIsLoggedInUseCase creates a new GetByIsLoggedInUseCase
func NewGetByIsLoggedInUseCase(configService config.ConfigService, userRepo user.Repository) GetByIsLoggedInUseCase {
	return &getByIsLoggedInUseCaseImpl{
		configService: configService,
		userRepo:      userRepo,
	}
}

// Execute retrieves a user by email
func (uc *getByIsLoggedInUseCaseImpl) Execute(ctx context.Context) (*user.User, error) {
	// Get the current user's email from configuration
	credentials, err := uc.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting authenticated user credentials: %w", err)
	}

	if credentials.Email == "" {
		return nil, fmt.Errorf("no authenticated user found")
	}

	user, err := uc.userRepo.GetByEmail(ctx, credentials.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}
