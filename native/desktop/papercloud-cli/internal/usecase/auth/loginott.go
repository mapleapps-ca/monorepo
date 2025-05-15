// monorepo/native/desktop/papercloud-cli/internal/usecase/auth/loginott.go
package auth

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/auth"
)

// LoginOTTUseCase defines the interface for login OTT use cases
type LoginOTTUseCase interface {
	RequestLoginOTT(ctx context.Context, email string) (*auth.LoginOTTResponse, error)
}

// loginOTTUseCase implements the LoginOTTUseCase interface
type loginOTTUseCase struct {
	logger     *zap.Logger
	repository auth.LoginOTTRepository
}

// NewLoginOTTUseCase creates a new login OTT use case
func NewLoginOTTUseCase(logger *zap.Logger, repository auth.LoginOTTRepository) LoginOTTUseCase {
	return &loginOTTUseCase{
		logger:     logger,
		repository: repository,
	}
}

// RequestLoginOTT handles the business logic for requesting a login OTT
func (uc *loginOTTUseCase) RequestLoginOTT(ctx context.Context, email string) (*auth.LoginOTTResponse, error) {
	// Validate input
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}

	// Sanitize input
	email = strings.ToLower(strings.TrimSpace(email))

	// Log the operation
	uc.logger.Info("Requesting login OTT", zap.String("email", email))

	// Create request and forward to repository
	request := &auth.LoginOTTRequest{
		Email: email,
	}

	return uc.repository.RequestLoginOTT(ctx, request)
}
