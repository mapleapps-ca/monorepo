// native/desktop/maplefile-cli/internal/repo/recovery/attempt.go
package recovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// Attempt tracking operations for rate limiting

func (r *recoveryRepository) CreateAttempt(ctx context.Context, attempt *recovery.RecoveryAttempt) error {
	r.logger.Debug("Creating recovery attempt",
		zap.String("email", attempt.Email),
		zap.String("ipAddress", attempt.IPAddress),
		zap.String("method", attempt.Method),
		zap.Bool("success", attempt.Success))

	// Validate attempt
	if err := recovery.ValidateRecoveryAttempt(attempt); err != nil {
		r.logger.Error("Invalid recovery attempt", zap.Error(err))
		return errors.NewAppError("invalid recovery attempt", err)
	}

	// Serialize attempt
	attemptBytes, err := r.serializeAttempt(attempt)
	if err != nil {
		r.logger.Error("Failed to serialize recovery attempt", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery attempt", err)
	}

	// Generate key using attempt ID
	key := r.generateAttemptKey(attempt.ID.String())

	// Save to database
	if err := r.dbClient.Set(key, attemptBytes); err != nil {
		r.logger.Error("Failed to save recovery attempt to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save recovery attempt to local storage", err)
	}

	r.logger.Info("Successfully created recovery attempt",
		zap.String("email", attempt.Email),
		zap.String("ipAddress", attempt.IPAddress),
		zap.String("method", attempt.Method),
		zap.Bool("success", attempt.Success))

	return nil
}

func (r *recoveryRepository) GetAttemptsByEmail(ctx context.Context, email string, since time.Time) ([]*recovery.RecoveryAttempt, error) {
	r.logger.Debug("Getting recovery attempts by email",
		zap.String("email", email),
		zap.Time("since", since))

	var attempts []*recovery.RecoveryAttempt

	// Iterate through all attempts
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, attemptKeyPrefix) {
			return nil // Skip non-attempt keys
		}

		// Deserialize attempt
		attempt, err := r.deserializeAttempt(value)
		if err != nil {
			r.logger.Error("Failed to deserialize attempt during email search",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if this attempt matches the email and time criteria
		if attempt.Email == email && attempt.AttemptedAt.After(since) {
			attempts = append(attempts, attempt)
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery attempts", zap.Error(err))
		return nil, errors.NewAppError("failed to search recovery attempts by email", err)
	}

	r.logger.Debug("Successfully found recovery attempts by email",
		zap.String("email", email),
		zap.Int("count", len(attempts)))

	return attempts, nil
}

func (r *recoveryRepository) GetAttemptsByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*recovery.RecoveryAttempt, error) {
	r.logger.Debug("Getting recovery attempts by IP address",
		zap.String("ipAddress", ipAddress),
		zap.Time("since", since))

	var attempts []*recovery.RecoveryAttempt

	// Iterate through all attempts
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, attemptKeyPrefix) {
			return nil // Skip non-attempt keys
		}

		// Deserialize attempt
		attempt, err := r.deserializeAttempt(value)
		if err != nil {
			r.logger.Error("Failed to deserialize attempt during IP search",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if this attempt matches the IP address and time criteria
		if attempt.IPAddress == ipAddress && attempt.AttemptedAt.After(since) {
			attempts = append(attempts, attempt)
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery attempts", zap.Error(err))
		return nil, errors.NewAppError("failed to search recovery attempts by IP", err)
	}

	r.logger.Debug("Successfully found recovery attempts by IP address",
		zap.String("ipAddress", ipAddress),
		zap.Int("count", len(attempts)))

	return attempts, nil
}

func (r *recoveryRepository) CountAttemptsByEmail(ctx context.Context, email string, since time.Time) (int64, error) {
	r.logger.Debug("Counting recovery attempts by email",
		zap.String("email", email),
		zap.Time("since", since))

	var count int64

	// Iterate through all attempts
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, attemptKeyPrefix) {
			return nil // Skip non-attempt keys
		}

		// Deserialize attempt
		attempt, err := r.deserializeAttempt(value)
		if err != nil {
			r.logger.Error("Failed to deserialize attempt during email count",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if this attempt matches the email and time criteria
		if attempt.Email == email && attempt.AttemptedAt.After(since) {
			count++
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery attempts for count", zap.Error(err))
		return 0, errors.NewAppError("failed to count recovery attempts by email", err)
	}

	r.logger.Debug("Successfully counted recovery attempts by email",
		zap.String("email", email),
		zap.Int64("count", count))

	return count, nil
}

func (r *recoveryRepository) CountAttemptsByIPAddress(ctx context.Context, ipAddress string, since time.Time) (int64, error) {
	r.logger.Debug("Counting recovery attempts by IP address",
		zap.String("ipAddress", ipAddress),
		zap.Time("since", since))

	var count int64

	// Iterate through all attempts
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, attemptKeyPrefix) {
			return nil // Skip non-attempt keys
		}

		// Deserialize attempt
		attempt, err := r.deserializeAttempt(value)
		if err != nil {
			r.logger.Error("Failed to deserialize attempt during IP count",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if this attempt matches the IP address and time criteria
		if attempt.IPAddress == ipAddress && attempt.AttemptedAt.After(since) {
			count++
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery attempts for count", zap.Error(err))
		return 0, errors.NewAppError("failed to count recovery attempts by IP", err)
	}

	r.logger.Debug("Successfully counted recovery attempts by IP address",
		zap.String("ipAddress", ipAddress),
		zap.Int64("count", count))

	return count, nil
}

func (r *recoveryRepository) DeleteOldAttempts(ctx context.Context, before time.Time) error {
	r.logger.Debug("Deleting old recovery attempts",
		zap.Time("before", before))

	deletedCount := 0

	// Iterate through all attempts
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, attemptKeyPrefix) {
			return nil // Skip non-attempt keys
		}

		// Deserialize attempt
		attempt, err := r.deserializeAttempt(value)
		if err != nil {
			r.logger.Error("Failed to deserialize attempt during cleanup",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if attempt is old
		if attempt.AttemptedAt.Before(before) {
			// Delete old attempt
			if err := r.dbClient.Delete(keyStr); err != nil {
				r.logger.Error("Failed to delete old attempt",
					zap.String("key", keyStr),
					zap.Error(err))
				return nil // Continue iteration despite error
			}
			deletedCount++
			r.logger.Debug("Deleted old recovery attempt",
				zap.String("email", attempt.Email),
				zap.Time("attemptedAt", attempt.AttemptedAt))
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery attempts for cleanup", zap.Error(err))
		return errors.NewAppError("failed to cleanup old attempts", err)
	}

	r.logger.Info("Successfully cleaned up old recovery attempts",
		zap.Int("deletedCount", deletedCount))

	return nil
}

// Cleanup operations

func (r *recoveryRepository) CleanupExpiredData(ctx context.Context) error {
	r.logger.Debug("Starting cleanup of expired recovery data")

	// Clean up expired sessions
	if err := r.DeleteExpiredSessions(ctx); err != nil {
		r.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
		return err
	}

	// Clean up expired challenges
	if err := r.DeleteExpiredChallenges(ctx); err != nil {
		r.logger.Error("Failed to cleanup expired challenges", zap.Error(err))
		return err
	}

	// Clean up expired tokens
	if err := r.DeleteExpiredTokens(ctx); err != nil {
		r.logger.Error("Failed to cleanup expired tokens", zap.Error(err))
		return err
	}

	// Clean up old attempts (older than 24 hours)
	before := time.Now().Add(-24 * time.Hour)
	if err := r.DeleteOldAttempts(ctx, before); err != nil {
		r.logger.Error("Failed to cleanup old attempts", zap.Error(err))
		return err
	}

	r.logger.Info("Successfully completed cleanup of expired recovery data")

	return nil
}

// Helper methods for attempts

func (r *recoveryRepository) generateAttemptKey(attemptID string) string {
	return fmt.Sprintf("%s%s", attemptKeyPrefix, attemptID)
}

func (r *recoveryRepository) serializeAttempt(attempt *recovery.RecoveryAttempt) ([]byte, error) {
	return attempt.Serialize() // This method would need to be implemented in the domain model
}

func (r *recoveryRepository) deserializeAttempt(data []byte) (*recovery.RecoveryAttempt, error) {
	return recovery.NewRecoveryAttemptFromDeserialized(data) // This method would need to be implemented in the domain model
}
