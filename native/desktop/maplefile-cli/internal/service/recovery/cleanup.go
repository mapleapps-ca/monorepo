// native/desktop/maplefile-cli/internal/service/recovery/cleanup.go
package recovery

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc_recovery "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/recovery"
)

// RecoveryCleanupService provides functionality for cleaning up expired recovery data
type RecoveryCleanupService interface {
	// CleanupExpiredSessions removes expired recovery sessions
	CleanupExpiredSessions(ctx context.Context) error
}

// recoveryCleanupService implements the RecoveryCleanupService interface
type recoveryCleanupService struct {
	logger                        *zap.Logger
	cleanupExpiredRecoveryUseCase uc_recovery.CleanupExpiredRecoveryDataUseCase
}

// NewRecoveryCleanupService creates a new recovery cleanup service
func NewRecoveryCleanupService(
	logger *zap.Logger,
	cleanupExpiredRecoveryUseCase uc_recovery.CleanupExpiredRecoveryDataUseCase,
) RecoveryCleanupService {
	logger = logger.Named("RecoveryCleanupService")
	return &recoveryCleanupService{
		logger:                        logger,
		cleanupExpiredRecoveryUseCase: cleanupExpiredRecoveryUseCase,
	}
}

// CleanupExpiredSessions removes expired recovery sessions
func (s *recoveryCleanupService) CleanupExpiredSessions(ctx context.Context) error {
	s.logger.Info("🧹 Starting cleanup of expired recovery sessions")

	err := s.cleanupExpiredRecoveryUseCase.Execute(ctx)
	if err != nil {
		return errors.NewAppError("failed to cleanup expired recovery sessions", err)
	}

	s.logger.Info("✅ Successfully cleaned up expired recovery sessions")

	return nil
}
