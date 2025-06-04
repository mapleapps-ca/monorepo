// monorepo/native/desktop/maplefile-cli/internal/service/authdto/verifyemail_service.go
package authdto

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	authUseCase "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/auth"
)

// EmailVerificationService provides high-level functionality for email verification
type EmailVerificationService interface {
	VerifyEmail(ctx context.Context, code string) (*VerificationResult, error)
}

// VerificationResult holds data returned after email verification
type VerificationResult struct {
	Message  string
	UserRole int
	Status   int
}

// emailVerificationService implements the EmailVerificationService interface
type emailVerificationService struct {
	logger         *zap.Logger
	useCase        authUseCase.EmailVerificationUseCase
	userRepository user.Repository
}

// NewEmailVerificationService creates a new email verification service
func NewEmailVerificationService(
	logger *zap.Logger,
	useCase authUseCase.EmailVerificationUseCase,
	userRepository user.Repository,
) EmailVerificationService {
	logger = logger.Named("EmailVerificationService")
	return &emailVerificationService{
		logger:         logger,
		useCase:        useCase,
		userRepository: userRepository,
	}
}

// VerifyEmail handles the entire flow of email verification
func (s *emailVerificationService) VerifyEmail(ctx context.Context, code string) (*VerificationResult, error) {
	// Call the use case to verify the email
	response, currentUser, err := s.useCase.VerifyEmail(ctx, code)
	if err != nil {
		return nil, errors.NewAppError("failed to verify email", err)
	}

	// Start a transaction to update the user
	if err := s.userRepository.OpenTransaction(); err != nil {
		return nil, errors.NewAppError("failed to open transaction", err)
	}

	// Update the user's verification status
	currentUser.WasEmailVerified = true
	currentUser.Role = int8(response.UserRole)
	currentUser.Status = user.UserStatusActive
	currentUser.ModifiedAt = time.Now()

	// Save the updated user
	if err := s.userRepository.UpsertByEmail(ctx, currentUser); err != nil {
		s.userRepository.DiscardTransaction()
		return nil, errors.NewAppError("failed to update user verification status", err)
	}

	// Commit the transaction
	if err := s.userRepository.CommitTransaction(); err != nil {
		s.userRepository.DiscardTransaction()
		return nil, errors.NewAppError("failed to commit transaction", err)
	}

	// Log success
	s.logger.Info("✅ Email verification successful",
		zap.String("email", currentUser.Email),
		zap.Int("role", int(currentUser.Role)))

	// Return the result
	return &VerificationResult{
		Message:  response.Message,
		UserRole: response.UserRole,
		Status:   response.Status,
	}, nil
}
