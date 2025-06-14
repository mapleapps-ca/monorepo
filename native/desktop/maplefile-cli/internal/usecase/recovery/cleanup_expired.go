// native/desktop/maplefile-cli/internal/usecase/recovery/cleanup_expired.go
package recovery

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// CleanupExpiredRecoveryDataUseCase defines the interface for cleaning up expired recovery data
type CleanupExpiredRecoveryDataUseCase interface {
	Execute(ctx context.Context) error
}

// cleanupExpiredRecoveryDataUseCase implements the CleanupExpiredRecoveryDataUseCase interface
type cleanupExpiredRecoveryDataUseCase struct {
	logger       *zap.Logger
	recoveryRepo recovery.RecoveryRepository
}

// NewCleanupExpiredRecoveryDataUseCase creates a new cleanup expired recovery data use case
func NewCleanupExpiredRecoveryDataUseCase(
	logger *zap.Logger,
	recoveryRepo recovery.RecoveryRepository,
) CleanupExpiredRecoveryDataUseCase {
	logger = logger.Named("CleanupExpiredRecoveryDataUseCase")
	return &cleanupExpiredRecoveryDataUseCase{
		logger:       logger,
		recoveryRepo: recoveryRepo,
	}
}

// Execute cleans up all expired recovery data
func (uc *cleanupExpiredRecoveryDataUseCase) Execute(ctx context.Context) error {
	uc.logger.Info("Starting cleanup of expired recovery data")

	startTime := time.Now()
	var errors []error

	//
	// STEP 1: Clean up expired sessions
	//
	uc.logger.Debug("Cleaning up expired recovery sessions")
	if err := uc.recoveryRepo.DeleteExpiredSessions(ctx); err != nil {
		uc.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
		errors = append(errors, err)
	}

	//
	// STEP 2: Clean up expired challenges
	//
	uc.logger.Debug("Cleaning up expired recovery challenges")
	if err := uc.recoveryRepo.DeleteExpiredChallenges(ctx); err != nil {
		uc.logger.Error("Failed to cleanup expired challenges", zap.Error(err))
		errors = append(errors, err)
	}

	//
	// STEP 3: Clean up expired tokens
	//
	uc.logger.Debug("Cleaning up expired recovery tokens")
	if err := uc.recoveryRepo.DeleteExpiredTokens(ctx); err != nil {
		uc.logger.Error("Failed to cleanup expired tokens", zap.Error(err))
		errors = append(errors, err)
	}

	//
	// STEP 4: Clean up old attempts (older than 24 hours)
	//
	oldAttemptsCutoff := time.Now().Add(-24 * time.Hour)
	uc.logger.Debug("Cleaning up old recovery attempts",
		zap.Time("before", oldAttemptsCutoff))

	if err := uc.recoveryRepo.DeleteOldAttempts(ctx, oldAttemptsCutoff); err != nil {
		uc.logger.Error("Failed to cleanup old attempts", zap.Error(err))
		errors = append(errors, err)
	}

	//
	// STEP 5: Log cleanup summary
	//
	duration := time.Since(startTime)
	if len(errors) > 0 {
		uc.logger.Warn("Cleanup completed with errors",
			zap.Duration("duration", duration),
			zap.Int("errorCount", len(errors)))
		// Return the first error
		return errors[0]
	}

	uc.logger.Info("Successfully completed cleanup of expired recovery data",
		zap.Duration("duration", duration))

	return nil
}
