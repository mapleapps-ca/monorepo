// monorepo/native/desktop/maplefile-cli/internal/usecase/auth/loginott.go
package auth

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/authdto"
)

// LoginOTTUseCase defines the interface for login OTT use cases
type LoginOTTUseCase interface {
	RequestLoginOTT(ctx context.Context, email string) (*dom_authdto.LoginOTTResponse, error)
}

// loginOTTUseCase implements the LoginOTTUseCase interface
type loginOTTUseCase struct {
	logger     *zap.Logger
	repository dom_authdto.LoginOTTRepository
}

// NewLoginOTTUseCase creates a new login OTT use case
func NewLoginOTTUseCase(logger *zap.Logger, repository dom_authdto.LoginOTTRepository) LoginOTTUseCase {
	logger = logger.Named("LoginOTTUseCase")
	return &loginOTTUseCase{
		logger:     logger,
		repository: repository,
	}
}

// RequestLoginOTT handles the business logic for requesting a login OTT
func (uc *loginOTTUseCase) RequestLoginOTT(ctx context.Context, email string) (*dom_authdto.LoginOTTResponse, error) {
	// Validate input
	if email == "" {
		return nil, errors.NewAppError("email is required", nil)
	}

	// Sanitize input
	email = strings.ToLower(strings.TrimSpace(email))

	// Log the operation
	uc.logger.Info("Requesting login OTT", zap.String("email", email))

	// Create request and forward to repository
	request := &dom_authdto.LoginOTTRequest{
		Email: email,
	}

	return uc.repository.RequestLoginOTT(ctx, request)
}
