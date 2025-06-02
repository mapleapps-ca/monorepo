// internal/usecase/auth/logout.go
package auth

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
)

// LogoutUseCase defines the interface for logout use cases
type LogoutUseCase interface {
	Logout(ctx context.Context) error
}

// logoutUseCase implements the LogoutUseCase interface
type logoutUseCase struct {
	logger        *zap.Logger
	configService config.ConfigService
}

// NewLogoutUseCase creates a new logout use case
func NewLogoutUseCase(
	logger *zap.Logger,
	configService config.ConfigService,
) LogoutUseCase {
	logger = logger.Named("LogoutUseCase")
	return &logoutUseCase{
		logger:        logger,
		configService: configService,
	}
}

// Logout handles the business logic for user logout
func (uc *logoutUseCase) Logout(ctx context.Context) error {
	// Check if user is currently logged in
	credentials, err := uc.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return errors.NewAppError("failed to get current user credentials", err)
	}

	if credentials == nil || credentials.Email == "" {
		return errors.NewAppError("no user is currently logged in", nil)
	}

	currentUserEmail := credentials.Email
	uc.logger.Info("Logging out user", zap.String("email", currentUserEmail))

	// Clear the user credentials from local storage
	if err := uc.configService.ClearLoggedInUserCredentials(ctx); err != nil {
		return errors.NewAppError("failed to clear user credentials", err)
	}

	uc.logger.Info("✅ User logged out successfully", zap.String("email", currentUserEmail))

	return nil
}
