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
type RefreshTokenUseCase interface {
	Execute(ctx context.Context) error
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

// Execute performs token refresh without password (for plaintext tokens only)
func (uc *refreshTokenUseCase) Execute(ctx context.Context) error {
	uc.logger.Info("Starting token refresh")

	// Check if user is logged in
	creds, err := uc.configService.GetLoggedInUserCredentials(ctx)
	if err != nil {
		return errors.NewAppError("failed to get current user credentials", err)
	}

	if creds == nil || creds.Email == "" {
		return errors.NewAppError("no user is currently logged in", nil)
	}

	// Attempt to refresh tokens
	_, err = uc.tokenRefreshService.ForceRefresh(ctx)
	if err != nil {
		// Check if this is an encrypted token error
		if err.Error() == "encrypted tokens received - password required for decryption. Please login again or use RefreshTokenWithPassword" {
			return errors.NewAppError("encrypted tokens detected - password required. Use 'maplefile-cli refreshtoken --password' or login again", nil)
		}
		return errors.NewAppError("failed to refresh token", err)
	}

	uc.logger.Info("Token refresh completed successfully")
	return nil
}

// ExecuteWithPassword performs token refresh with password (for encrypted tokens)
func (uc *refreshTokenUseCase) ExecuteWithPassword(ctx context.Context, password string) error {
	if password == "" {
		return errors.NewAppError("password is required for encrypted token refresh", nil)
	}

	uc.logger.Info("Starting token refresh with password")

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
		return errors.NewAppError("failed to refresh token with password", err)
	}

	uc.logger.Info("Token refresh with password completed successfully")
	return nil
}
