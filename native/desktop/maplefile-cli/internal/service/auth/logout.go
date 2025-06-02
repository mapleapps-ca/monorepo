// internal/service/auth/logout.go
package auth

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	authUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/auth"
)

// LogoutService provides high-level functionality for user logout
type LogoutService interface {
	Logout(ctx context.Context) error
}

// logoutService implements the LogoutService interface
type logoutService struct {
	logger  *zap.Logger
	useCase authUseCase.LogoutUseCase
}

// NewLogoutService creates a new logout service
func NewLogoutService(
	logger *zap.Logger,
	useCase authUseCase.LogoutUseCase,
) LogoutService {
	logger = logger.Named("LogoutService")
	return &logoutService{
		logger:  logger,
		useCase: useCase,
	}
}

// Logout handles the entire flow of user logout
func (s *logoutService) Logout(ctx context.Context) error {
	s.logger.Info("🚪 Processing logout request")

	// Call the use case to perform logout
	if err := s.useCase.Logout(ctx); err != nil {
		s.logger.Error("❌ Logout failed", zap.Error(err))
		return errors.NewAppError("logout failed", err)
	}

	s.logger.Info("✅ Logout completed successfully")
	return nil
}
