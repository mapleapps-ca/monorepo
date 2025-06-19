// native/desktop/maplefile-cli/internal/usecase/refreshtoken/refreshtoken.go
package refreshtoken

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/authdto"
)

// RefreshTokenUseCase defines the interface for token refresh operations
// Note: Legacy Execute() method removed - password is now always required
type RefreshTokenUseCase interface {
	ExecuteWithPassword(ctx context.Context, password string) error
}

// refreshTokenUseCase implements the RefreshTokenUseCase interface
type refreshTokenUseCase struct {
	logger                 *zap.Logger
	configService          config.ConfigService
	userRepo               user.Repository
	tokenDecryptionService svc_authdto.TokenDecryptionService
	tokenRefreshService    svc_authdto.TokenRefreshService
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(
	logger *zap.Logger,
	configService config.ConfigService,
	userRepo user.Repository,
	tokenDecryptionService svc_authdto.TokenDecryptionService,
	tokenRefreshService svc_authdto.TokenRefreshService,
) RefreshTokenUseCase {
	logger = logger.Named("RefreshTokenUseCase")
	return &refreshTokenUseCase{
		logger:                 logger,
		configService:          configService,
		userRepo:               userRepo,
		tokenDecryptionService: tokenDecryptionService,
		tokenRefreshService:    tokenRefreshService,
	}
}

// ExecuteWithPassword performs token refresh with password (required for encrypted tokens)
func (uc *refreshTokenUseCase) ExecuteWithPassword(ctx context.Context, password string) error {
	if password == "" {
		return errors.NewAppError("password is required for encrypted token refresh", nil)
	}

	uc.logger.Info("Starting encrypted token refresh")

	// Check if user is logged in
	creds, err := uc.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return errors.NewAppError("failed to get current user credentials", err)
	}

	if creds == nil || creds.Email == "" {
		return errors.NewAppError("no user is currently logged in", nil)
	}

	// Refresh tokens with password for decryption
	_, err = uc.tokenRefreshService.RefreshTokenWithPassword(ctx, password)
	if err != nil {
		return errors.NewAppError("failed to refresh encrypted tokens", err)
	}

	uc.logger.Info("Encrypted token refresh completed successfully")
	return nil
}
