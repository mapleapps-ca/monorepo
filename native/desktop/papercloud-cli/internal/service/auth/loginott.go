// monorepo/native/desktop/papercloud-cli/internal/service/auth/loginott.go
package auth

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/usecase/auth"
)

// LoginOTTService provides high-level functionality for login OTT operations
type LoginOTTService interface {
	RequestLoginOTT(ctx context.Context, email string) error
}

// loginOTTService implements the LoginOTTService interface
type loginOTTService struct {
	logger  *zap.Logger
	useCase auth.LoginOTTUseCase
}

// NewLoginOTTService creates a new login OTT service
func NewLoginOTTService(logger *zap.Logger, useCase auth.LoginOTTUseCase) LoginOTTService {
	return &loginOTTService{
		logger:  logger,
		useCase: useCase,
	}
}

// RequestLoginOTT handles the entire flow of requesting a login OTT
func (s *loginOTTService) RequestLoginOTT(ctx context.Context, email string) error {
	response, err := s.useCase.RequestLoginOTT(ctx, email)
	if err != nil {
		return errors.NewAppError("failed to request login one-time token", err)
	}

	// Log success
	s.logger.Info("Login OTT request successful", zap.String("email", email))

	// Additional service-level logic could be added here if needed
	_ = response

	return nil
}
