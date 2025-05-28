// monorepo/native/desktop/maplefile-cli/internal/service/auth/verifyloginott.go
package auth

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	authUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/auth"
)

// LoginOTTVerificationService provides high-level functionality for login OTT verification
type LoginOTTVerificationService interface {
	VerifyLoginOTT(ctx context.Context, email, ott string) error
}

// loginOTTVerificationService implements the LoginOTTVerificationService interface
type loginOTTVerificationService struct {
	logger   *zap.Logger
	useCase  authUseCase.LoginOTTVerificationUseCase
	userRepo user.Repository
}

// NewLoginOTTVerificationService creates a new login OTT verification service
func NewLoginOTTVerificationService(
	logger *zap.Logger,
	useCase authUseCase.LoginOTTVerificationUseCase,
	userRepo user.Repository,
) LoginOTTVerificationService {
	logger = logger.Named("LoginOTTVerificationService")
	return &loginOTTVerificationService{
		logger:   logger,
		useCase:  useCase,
		userRepo: userRepo,
	}
}

// VerifyLoginOTT handles the entire flow of verifying a login OTT
func (s *loginOTTVerificationService) VerifyLoginOTT(ctx context.Context, email, ott string) error {
	// Call the use case to verify the OTT and get updated user
	_, user, err := s.useCase.VerifyLoginOTT(ctx, email, ott)
	if err != nil {
		return errors.NewAppError("failed to verify login one-time token", err)
	}

	// Start a transaction to update the user
	if err := s.userRepo.OpenTransaction(); err != nil {
		return errors.NewAppError("failed to open transaction", err)
	}

	// Update the user in the repository
	if err := s.userRepo.UpsertByEmail(ctx, user); err != nil {
		s.userRepo.DiscardTransaction()
		return errors.NewAppError("failed to update user with verification data", err)
	}

	// Commit the transaction
	if err := s.userRepo.CommitTransaction(); err != nil {
		s.userRepo.DiscardTransaction()
		return errors.NewAppError("failed to commit transaction", err)
	}

	// Log success
	s.logger.Info("Login OTT verified successfully", zap.String("email", email))

	return nil
}
