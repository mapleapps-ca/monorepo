// monorepo/native/desktop/maplefile-cli/internal/service/authdto/completelogin_service.go
package authdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
)

// CompleteLoginService provides high-level functionality for login completion
type CompleteLoginService interface {
	CompleteLogin(ctx context.Context, email, password string) (*dom_authdto.TokenResponse, error)
}

// completeLoginService implements the CompleteLoginService interface
type completeLoginService struct {
	logger        *zap.Logger
	useCase       uc_authdto.CompleteLoginUseCase
	userRepo      user.Repository
	configService config.ConfigService
}

// NewCompleteLoginService creates a new login completion service
func NewCompleteLoginService(
	logger *zap.Logger,
	useCase uc_authdto.CompleteLoginUseCase,
	userRepo user.Repository,
	configService config.ConfigService,
) CompleteLoginService {
	logger = logger.Named("CompleteLoginService")
	return &completeLoginService{
		logger:        logger,
		useCase:       useCase,
		userRepo:      userRepo,
		configService: configService,
	}
}

// CompleteLogin handles the entire flow of login completion
func (s *completeLoginService) CompleteLogin(ctx context.Context, email, password string) (*dom_authdto.TokenResponse, error) {
	// Call the use case to complete login and get token and updated user
	tokenResp, updatedUser, err := s.useCase.CompleteLogin(ctx, email, password)
	if err != nil {
		return nil, errors.NewAppError("failed to complete login", err)
	}

	// Start a transaction to update the user
	if err := s.userRepo.OpenTransaction(); err != nil {
		return nil, errors.NewAppError("failed to open transaction", err)
	}

	// Save the updated user
	if err := s.userRepo.UpsertByEmail(ctx, updatedUser); err != nil {
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to update user data", err)
	}

	// Commit the transaction
	if err := s.userRepo.CommitTransaction(); err != nil {
		s.userRepo.DiscardTransaction()
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	// Save our authenticated credentials to configuration
	s.configService.SetLoggedInUserCredentials(
		ctx,
		email,
		tokenResp.AccessToken,
		&tokenResp.AccessTokenExpiryTime,
		tokenResp.RefreshToken,
		&tokenResp.RefreshTokenExpiryTime,
	)

	// Log success
	s.logger.Info("✅ Login completed successfully",
		zap.String("email", email),
		zap.Time("accessTokenExpiry", tokenResp.AccessTokenExpiryTime),
		zap.Time("refreshTokenExpiry", tokenResp.RefreshTokenExpiryTime))

	return tokenResp, nil
}
