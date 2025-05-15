// monorepo/native/desktop/papercloud-cli/internal/usecase/auth/verifyemail.go
package auth

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
)

// EmailVerificationUseCase defines the interface for email verification use cases
type EmailVerificationUseCase interface {
	VerifyEmail(ctx context.Context, code string) (*auth.VerifyEmailResponse, *user.User, error)
}

// emailVerificationUseCase implements the EmailVerificationUseCase interface
type emailVerificationUseCase struct {
	logger         *zap.Logger
	repository     auth.EmailVerificationRepository
	userRepository user.Repository
}

// NewEmailVerificationUseCase creates a new email verification use case
func NewEmailVerificationUseCase(
	logger *zap.Logger,
	repository auth.EmailVerificationRepository,
	userRepository user.Repository,
) EmailVerificationUseCase {
	return &emailVerificationUseCase{
		logger:         logger,
		repository:     repository,
		userRepository: userRepository,
	}
}

// VerifyEmail verifies an email with the provided code
func (uc *emailVerificationUseCase) VerifyEmail(ctx context.Context, code string) (*auth.VerifyEmailResponse, *user.User, error) {
	// Validate input
	if code == "" {
		return nil, nil, errors.NewAppError("verification code is required", nil)
	}

	// Get a list of users to find a registered one
	users, err := uc.userRepository.ListAll(ctx)
	if err != nil {
		return nil, nil, errors.NewAppError("failed to retrieve users", err)
	}

	if len(users) == 0 {
		return nil, nil, errors.NewAppError("no registered user found", nil)
	}

	// Get the first user - this is simple but in a real app
	// we might want to allow the user to select which account to verify
	currentUser := users[0]

	// Call the repository to verify the email
	response, err := uc.repository.VerifyEmail(ctx, code)
	if err != nil {
		return nil, nil, err
	}

	return response, currentUser, nil
}
