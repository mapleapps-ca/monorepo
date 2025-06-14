// native/desktop/maplefile-cli/internal/usecase/recovery/track_attempt.go
package recovery

import (
	"context"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// TrackRecoveryAttemptUseCase defines the interface for tracking recovery attempts
type TrackRecoveryAttemptUseCase interface {
	Execute(ctx context.Context, email string, ipAddress string, method string, success bool, userAgent string) error
}

// trackRecoveryAttemptUseCase implements the TrackRecoveryAttemptUseCase interface
type trackRecoveryAttemptUseCase struct {
	logger       *zap.Logger
	recoveryRepo recovery.RecoveryRepository
}

// NewTrackRecoveryAttemptUseCase creates a new track recovery attempt use case
func NewTrackRecoveryAttemptUseCase(
	logger *zap.Logger,
	recoveryRepo recovery.RecoveryRepository,
) TrackRecoveryAttemptUseCase {
	logger = logger.Named("TrackRecoveryAttemptUseCase")
	return &trackRecoveryAttemptUseCase{
		logger:       logger,
		recoveryRepo: recoveryRepo,
	}
}

// Execute tracks a recovery attempt for rate limiting
func (uc *trackRecoveryAttemptUseCase) Execute(ctx context.Context, email string, ipAddress string, method string, success bool, userAgent string) error {
	//
	// STEP 1: Validate inputs
	//
	if email == "" {
		return errors.NewAppError("email is required", nil)
	}
	if ipAddress == "" {
		// Default to localhost if no IP provided
		ipAddress = "127.0.0.1"
	}
	if method == "" {
		method = recovery.DefaultRecoveryMethod
	}

	//
	// STEP 2: Create recovery attempt record
	//
	attempt := &recovery.RecoveryAttempt{
		ID:          gocql.TimeUUID(),
		Email:       email,
		IPAddress:   ipAddress,
		AttemptedAt: time.Now(),
		Success:     success,
		Method:      method,
		UserAgent:   userAgent,
	}

	//
	// STEP 3: Save the attempt
	//
	uc.logger.Debug("Tracking recovery attempt",
		zap.String("email", email),
		zap.String("ipAddress", ipAddress),
		zap.String("method", method),
		zap.Bool("success", success))

	if err := uc.recoveryRepo.CreateAttempt(ctx, attempt); err != nil {
		uc.logger.Error("Failed to track recovery attempt", zap.Error(err))
		return errors.NewAppError("failed to track recovery attempt", err)
	}

	uc.logger.Info("Successfully tracked recovery attempt",
		zap.String("attemptID", attempt.ID.String()),
		zap.String("email", email),
		zap.Bool("success", success))

	return nil
}
