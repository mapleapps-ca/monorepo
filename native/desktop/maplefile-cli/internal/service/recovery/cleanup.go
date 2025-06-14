// native/desktop/maplefile-cli/internal/service/recovery/cleanup.go
package recovery

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc_recovery "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/recovery"
)

// RecoveryCleanupService provides functionality for cleaning up expired recovery data
type RecoveryCleanupService interface {
	// CleanupExpiredData removes all expired recovery sessions, tokens, and old attempts
	CleanupExpiredData(ctx context.Context) (*CleanupResult, error)

	// StartPeriodicCleanup starts a background goroutine that periodically cleans up expired data
	StartPeriodicCleanup(ctx context.Context, interval time.Duration) error

	// StopPeriodicCleanup stops the periodic cleanup
	StopPeriodicCleanup()
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	StartTime            time.Time     `json:"start_time"`
	EndTime              time.Time     `json:"end_time"`
	Duration             time.Duration `json:"duration"`
	ExpiredSessionsCount int           `json:"expired_sessions_count"`
	ExpiredTokensCount   int           `json:"expired_tokens_count"`
	OldAttemptsCount     int           `json:"old_attempts_count"`
	Success              bool          `json:"success"`
	Error                string        `json:"error,omitempty"`
}

// recoveryCleanupService implements the RecoveryCleanupService interface
type recoveryCleanupService struct {
	logger                            *zap.Logger
	cleanupExpiredRecoveryDataUseCase uc_recovery.CleanupExpiredRecoveryDataUseCase

	// For periodic cleanup
	stopChan chan bool
	stopped  bool
}

// NewRecoveryCleanupService creates a new recovery cleanup service
func NewRecoveryCleanupService(
	logger *zap.Logger,
	cleanupExpiredRecoveryDataUseCase uc_recovery.CleanupExpiredRecoveryDataUseCase,
) RecoveryCleanupService {
	logger = logger.Named("RecoveryCleanupService")
	return &recoveryCleanupService{
		logger:                            logger,
		cleanupExpiredRecoveryDataUseCase: cleanupExpiredRecoveryDataUseCase,
		stopChan:                          make(chan bool),
		stopped:                           false,
	}
}

// CleanupExpiredData removes all expired recovery sessions, tokens, and old attempts
func (s *recoveryCleanupService) CleanupExpiredData(ctx context.Context) (*CleanupResult, error) {
	s.logger.Info("🧹 Starting recovery data cleanup")

	result := &CleanupResult{
		StartTime: time.Now(),
	}

	//
	// Execute cleanup
	//
	err := s.cleanupExpiredRecoveryDataUseCase.Execute(ctx)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		s.logger.Error("❌ Recovery cleanup failed",
			zap.Error(err),
			zap.Duration("duration", result.Duration))

		result.Success = false
		result.Error = err.Error()
		return result, errors.NewAppError("recovery cleanup failed", err)
	}

	result.Success = true

	// Note: In a real implementation, the use case would return counts
	// For now, we'll just indicate success
	s.logger.Info("✅ Recovery cleanup completed successfully",
		zap.Duration("duration", result.Duration))

	return result, nil
}

// StartPeriodicCleanup starts a background goroutine that periodically cleans up expired data
func (s *recoveryCleanupService) StartPeriodicCleanup(ctx context.Context, interval time.Duration) error {
	if s.stopped {
		return errors.NewAppError("cleanup service has been stopped", nil)
	}

	// Validate interval
	if interval < time.Minute {
		return errors.NewAppError("cleanup interval must be at least 1 minute", nil)
	}

	s.logger.Info("🔄 Starting periodic recovery cleanup",
		zap.Duration("interval", interval))

	// Start cleanup goroutine
	go s.runPeriodicCleanup(ctx, interval)

	return nil
}

// StopPeriodicCleanup stops the periodic cleanup
func (s *recoveryCleanupService) StopPeriodicCleanup() {
	if !s.stopped {
		s.logger.Info("🛑 Stopping periodic recovery cleanup")
		s.stopped = true
		close(s.stopChan)
	}
}

// runPeriodicCleanup runs the cleanup task periodically
func (s *recoveryCleanupService) runPeriodicCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run initial cleanup
	s.performCleanup(ctx)

	for {
		select {
		case <-ticker.C:
			s.performCleanup(ctx)
		case <-s.stopChan:
			s.logger.Info("Periodic recovery cleanup stopped")
			return
		case <-ctx.Done():
			s.logger.Info("Periodic recovery cleanup stopped due to context cancellation")
			return
		}
	}
}

// performCleanup executes a single cleanup operation
func (s *recoveryCleanupService) performCleanup(ctx context.Context) {
	s.logger.Debug("Running scheduled recovery cleanup")

	// Create a timeout context for the cleanup operation
	cleanupCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	result, err := s.CleanupExpiredData(cleanupCtx)
	if err != nil {
		s.logger.Error("Scheduled recovery cleanup failed",
			zap.Error(err),
			zap.Duration("duration", result.Duration))
	} else {
		s.logger.Debug("Scheduled recovery cleanup completed",
			zap.Duration("duration", result.Duration))
	}
}
