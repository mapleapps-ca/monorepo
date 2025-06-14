// native/desktop/maplefile-cli/internal/usecase/recovery/check_rate_limit.go
package recovery

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// CheckRateLimitUseCase defines the interface for checking recovery rate limits
type CheckRateLimitUseCase interface {
	Execute(ctx context.Context, email string, ipAddress string) error
}

// checkRateLimitUseCase implements the CheckRateLimitUseCase interface
type checkRateLimitUseCase struct {
	logger       *zap.Logger
	recoveryRepo recovery.RecoveryRepository
}

// NewCheckRateLimitUseCase creates a new check rate limit use case
func NewCheckRateLimitUseCase(
	logger *zap.Logger,
	recoveryRepo recovery.RecoveryRepository,
) CheckRateLimitUseCase {
	logger = logger.Named("CheckRateLimitUseCase")
	return &checkRateLimitUseCase{
		logger:       logger,
		recoveryRepo: recoveryRepo,
	}
}

// Execute checks if the recovery rate limit has been exceeded
func (uc *checkRateLimitUseCase) Execute(ctx context.Context, email string, ipAddress string) error {
	//
	// STEP 1: Define the time window for rate limiting
	//
	since := time.Now().Add(-recovery.RecoveryAttemptWindow)

	//
	// STEP 2: Count attempts by email
	//
	emailAttempts, err := uc.recoveryRepo.CountAttemptsByEmail(ctx, email, since)
	if err != nil {
		uc.logger.Error("Failed to count attempts by email",
			zap.String("email", email),
			zap.Error(err))
		// Continue with IP check even if email check fails
		emailAttempts = 0
	}

	//
	// STEP 3: Check if email attempts exceed limit
	//
	if emailAttempts >= recovery.MaxRecoveryAttemptsPerWindow {
		uc.logger.Warn("Recovery rate limit exceeded for email",
			zap.String("email", email),
			zap.Int64("attempts", emailAttempts),
			zap.Int("limit", recovery.MaxRecoveryAttemptsPerWindow))

		return recovery.NewRateLimitExceededError(email, int(emailAttempts))
	}

	//
	// STEP 4: Count attempts by IP address
	//
	ipAttempts, err := uc.recoveryRepo.CountAttemptsByIPAddress(ctx, ipAddress, since)
	if err != nil {
		uc.logger.Error("Failed to count attempts by IP address",
			zap.String("ipAddress", ipAddress),
			zap.Error(err))
		// If we can't check IP attempts, allow the request but log the issue
		return nil
	}

	//
	// STEP 5: Check if IP attempts exceed limit
	//
	// Use a higher limit for IP addresses to account for shared networks
	ipLimit := recovery.MaxRecoveryAttemptsPerWindow * 3
	if ipAttempts >= int64(ipLimit) {
		uc.logger.Warn("Recovery rate limit exceeded for IP address",
			zap.String("ipAddress", ipAddress),
			zap.Int64("attempts", ipAttempts),
			zap.Int("limit", ipLimit))

		return errors.NewAppError("recovery rate limit exceeded", recovery.ErrRateLimitExceeded)
	}

	//
	// STEP 6: Rate limit check passed
	//
	uc.logger.Debug("Recovery rate limit check passed",
		zap.String("email", email),
		zap.String("ipAddress", ipAddress),
		zap.Int64("emailAttempts", emailAttempts),
		zap.Int64("ipAttempts", ipAttempts))

	return nil
}
