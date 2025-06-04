// monorepo/native/desktop/maplefile-cli/internal/service/authdto/recovery.go
package authdto

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_authdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/authdto"
)

// RecoveryService provides high-level functionality for account recovery
type RecoveryService interface {
	// InitiateRecovery starts the recovery process with email and recovery key
	InitiateRecovery(ctx context.Context, email, recoveryKey string) error

	// SetNewPassword sets a new password to complete recovery
	SetNewPassword(ctx context.Context, newPassword string) error

	// GetRecoveryStatus returns the current recovery session status
	GetRecoveryStatus(ctx context.Context) (*RecoveryStatus, error)
}

// RecoveryStatus represents the current state of recovery
type RecoveryStatus struct {
	InProgress bool
	Email      string
	ExpiresAt  time.Time
}

// recoveryService implements the RecoveryService interface
type recoveryService struct {
	logger        *zap.Logger
	useCase       uc_authdto.RecoveryUseCase
	userRepo      user.Repository
	configService config.ConfigService

	// In-memory storage for recovery session
	mu           sync.Mutex
	recoveryData *uc_authdto.RecoveryData
}

// NewRecoveryService creates a new recovery service
func NewRecoveryService(
	logger *zap.Logger,
	useCase uc_authdto.RecoveryUseCase,
	userRepo user.Repository,
	configService config.ConfigService,
) RecoveryService {
	logger = logger.Named("RecoveryService")
	return &recoveryService{
		logger:        logger,
		useCase:       useCase,
		userRepo:      userRepo,
		configService: configService,
	}
}

// InitiateRecovery starts the recovery process
func (s *recoveryService) InitiateRecovery(ctx context.Context, email, recoveryKey string) error {
	s.logger.Info("🔐 Starting account recovery", zap.String("email", email))

	// Call use case to initiate recovery
	recoveryData, err := s.useCase.InitiateRecovery(ctx, email, recoveryKey)
	if err != nil {
		return errors.NewAppError("failed to initiate recovery", err)
	}

	// Store recovery data in memory
	s.mu.Lock()
	s.recoveryData = recoveryData
	s.mu.Unlock()

	s.logger.Info("✅ Recovery initiated successfully",
		zap.String("email", email),
		zap.Time("expiresAt", recoveryData.ExpiresAt))

	return nil
}

// SetNewPassword sets a new password to complete recovery
func (s *recoveryService) SetNewPassword(ctx context.Context, newPassword string) error {
	s.mu.Lock()
	recoveryData := s.recoveryData
	s.mu.Unlock()

	if recoveryData == nil {
		return errors.NewAppError("no recovery session in progress", nil)
	}

	// Check if session has expired
	if time.Now().After(recoveryData.ExpiresAt) {
		s.mu.Lock()
		s.recoveryData = nil
		s.mu.Unlock()
		return errors.NewAppError("recovery session has expired", nil)
	}

	s.logger.Info("🔐 Setting new password", zap.String("email", recoveryData.Email))

	// Start a transaction to update the user
	if err := s.userRepo.OpenTransaction(); err != nil {
		return errors.NewAppError("failed to open transaction", err)
	}

	// Call use case to complete recovery
	response, updatedUser, err := s.useCase.CompleteRecovery(ctx, recoveryData, newPassword)
	if err != nil {
		s.userRepo.DiscardTransaction()
		return errors.NewAppError("failed to complete recovery", err)
	}

	// Save the updated user
	if err := s.userRepo.UpsertByEmail(ctx, updatedUser); err != nil {
		s.userRepo.DiscardTransaction()
		return errors.NewAppError("failed to update user data", err)
	}

	// Commit the transaction
	if err := s.userRepo.CommitTransaction(); err != nil {
		s.userRepo.DiscardTransaction()
		return errors.NewAppError("failed to commit transaction", err)
	}

	// Save authenticated credentials to configuration
	if response.AccessToken != "" && response.RefreshToken != "" {
		s.configService.SetLoggedInUserCredentials(
			ctx,
			recoveryData.Email,
			response.AccessToken,
			&response.AccessTokenExpiryTime,
			response.RefreshToken,
			&response.RefreshTokenExpiryTime,
		)
	}

	// Clear recovery data
	s.mu.Lock()
	s.recoveryData = nil
	s.mu.Unlock()

	s.logger.Info("✅ Password reset successfully", zap.String("email", recoveryData.Email))

	return nil
}

// GetRecoveryStatus returns the current recovery session status
func (s *recoveryService) GetRecoveryStatus(ctx context.Context) (*RecoveryStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.recoveryData == nil {
		return &RecoveryStatus{
			InProgress: false,
		}, nil
	}

	// Check if expired
	if time.Now().After(s.recoveryData.ExpiresAt) {
		s.recoveryData = nil
		return &RecoveryStatus{
			InProgress: false,
		}, nil
	}

	return &RecoveryStatus{
		InProgress: true,
		Email:      s.recoveryData.Email,
		ExpiresAt:  s.recoveryData.ExpiresAt,
	}, nil
}

// RecoveryKeyService provides functionality for managing recovery keys
type RecoveryKeyService interface {
	// ShowRecoveryKey displays the user's recovery key
	ShowRecoveryKey(ctx context.Context, password string) (string, error)

	// RegenerateRecoveryKey creates a new recovery key
	RegenerateRecoveryKey(ctx context.Context, password string) (string, error)
}

// recoveryKeyService implements the RecoveryKeyService interface
type recoveryKeyService struct {
	logger   *zap.Logger
	userRepo user.Repository
}

// NewRecoveryKeyService creates a new recovery key service
func NewRecoveryKeyService(
	logger *zap.Logger,
	userRepo user.Repository,
) RecoveryKeyService {
	logger = logger.Named("RecoveryKeyService")
	return &recoveryKeyService{
		logger:   logger,
		userRepo: userRepo,
	}
}

// ShowRecoveryKey displays the user's recovery key
func (s *recoveryKeyService) ShowRecoveryKey(ctx context.Context, password string) (string, error) {
	// This would be implemented to decrypt and show the existing recovery key
	// For now, returning a placeholder
	return "", errors.NewAppError("show recovery key not implemented", nil)
}

// RegenerateRecoveryKey creates a new recovery key
func (s *recoveryKeyService) RegenerateRecoveryKey(ctx context.Context, password string) (string, error) {
	// This would be implemented to generate a new recovery key
	// For now, returning a placeholder
	return "", errors.NewAppError("regenerate recovery key not implemented", nil)
}
