// native/desktop/maplefile-cli/internal/repo/recovery/challenge.go
package recovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/recovery"
)

// Challenge management operations

func (r *recoveryRepository) CreateChallenge(ctx context.Context, challenge *recovery.RecoveryChallenge) error {
	r.logger.Debug("Creating recovery challenge",
		zap.String("challengeID", challenge.ChallengeID.String()),
		zap.String("sessionID", challenge.SessionID.String()))

	// Validate challenge
	if err := recovery.ValidateRecoveryChallenge(challenge); err != nil {
		r.logger.Error("Invalid recovery challenge", zap.Error(err))
		return errors.NewAppError("invalid recovery challenge", err)
	}

	// Serialize challenge
	challengeBytes, err := r.serializeChallenge(challenge)
	if err != nil {
		r.logger.Error("Failed to serialize recovery challenge", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery challenge", err)
	}

	// Generate key
	key := r.generateChallengeKey(challenge.ChallengeID.String())

	// Save to database
	if err := r.dbClient.Set(key, challengeBytes); err != nil {
		r.logger.Error("Failed to save recovery challenge to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save recovery challenge to local storage", err)
	}

	r.logger.Info("Successfully created recovery challenge",
		zap.String("challengeID", challenge.ChallengeID.String()),
		zap.String("sessionID", challenge.SessionID.String()))

	return nil
}

func (r *recoveryRepository) GetChallengeByID(ctx context.Context, challengeID gocql.UUID) (*recovery.RecoveryChallenge, error) {
	r.logger.Debug("Getting recovery challenge by ID",
		zap.String("challengeID", challengeID.String()))

	// Generate key
	key := r.generateChallengeKey(challengeID.String())

	// Get from database
	challengeBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve recovery challenge from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve recovery challenge from local storage", err)
	}

	// Check if challenge was found
	if challengeBytes == nil {
		r.logger.Debug("Recovery challenge not found",
			zap.String("challengeID", challengeID.String()))
		return nil, nil
	}

	// Deserialize challenge
	challenge, err := r.deserializeChallenge(challengeBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize recovery challenge", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize recovery challenge", err)
	}

	r.logger.Debug("Successfully retrieved recovery challenge",
		zap.String("challengeID", challengeID.String()),
		zap.String("sessionID", challenge.SessionID.String()))

	return challenge, nil
}

func (r *recoveryRepository) GetChallengeBySessionID(ctx context.Context, sessionID gocql.UUID) (*recovery.RecoveryChallenge, error) {
	r.logger.Debug("Getting recovery challenge by session ID",
		zap.String("sessionID", sessionID.String()))

	var foundChallenge *recovery.RecoveryChallenge

	// Iterate through all challenges to find the one with matching session ID
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, challengeKeyPrefix) {
			return nil // Skip non-challenge keys
		}

		// Deserialize challenge
		challenge, err := r.deserializeChallenge(value)
		if err != nil {
			r.logger.Error("Failed to deserialize challenge during session search",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if this challenge belongs to the session
		if challenge.SessionID == sessionID {
			foundChallenge = challenge
			return fmt.Errorf("found") // Use error to break iteration
		}

		return nil
	})

	// Check if we found the challenge (error "found" is expected)
	if err != nil && err.Error() != "found" {
		r.logger.Error("Error iterating through recovery challenges", zap.Error(err))
		return nil, errors.NewAppError("failed to search recovery challenges", err)
	}

	if foundChallenge != nil {
		r.logger.Debug("Successfully found recovery challenge by session ID",
			zap.String("challengeID", foundChallenge.ChallengeID.String()),
			zap.String("sessionID", sessionID.String()))
	} else {
		r.logger.Debug("Recovery challenge not found for session",
			zap.String("sessionID", sessionID.String()))
	}

	return foundChallenge, nil
}

func (r *recoveryRepository) UpdateChallenge(ctx context.Context, challenge *recovery.RecoveryChallenge) error {
	r.logger.Debug("Updating recovery challenge",
		zap.String("challengeID", challenge.ChallengeID.String()),
		zap.String("sessionID", challenge.SessionID.String()))

	// Validate challenge
	if err := recovery.ValidateRecoveryChallenge(challenge); err != nil {
		r.logger.Error("Invalid recovery challenge for update", zap.Error(err))
		return errors.NewAppError("invalid recovery challenge", err)
	}

	// Check if challenge exists
	existing, err := r.GetChallengeByID(ctx, challenge.ChallengeID)
	if err != nil {
		return err
	}
	if existing == nil {
		r.logger.Error("Recovery challenge not found for update",
			zap.String("challengeID", challenge.ChallengeID.String()))
		return errors.NewAppError("recovery challenge not found", recovery.ErrChallengeNotFound)
	}

	// Serialize challenge
	challengeBytes, err := r.serializeChallenge(challenge)
	if err != nil {
		r.logger.Error("Failed to serialize recovery challenge for update", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery challenge", err)
	}

	// Generate key
	key := r.generateChallengeKey(challenge.ChallengeID.String())

	// Update in database
	if err := r.dbClient.Set(key, challengeBytes); err != nil {
		r.logger.Error("Failed to update recovery challenge in local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to update recovery challenge in local storage", err)
	}

	r.logger.Info("Successfully updated recovery challenge",
		zap.String("challengeID", challenge.ChallengeID.String()),
		zap.String("sessionID", challenge.SessionID.String()))

	return nil
}

func (r *recoveryRepository) DeleteChallenge(ctx context.Context, challengeID gocql.UUID) error {
	r.logger.Debug("Deleting recovery challenge",
		zap.String("challengeID", challengeID.String()))

	// Generate key
	key := r.generateChallengeKey(challengeID.String())

	// Delete from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete recovery challenge from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete recovery challenge from local storage", err)
	}

	r.logger.Info("Successfully deleted recovery challenge",
		zap.String("challengeID", challengeID.String()))

	return nil
}

func (r *recoveryRepository) DeleteExpiredChallenges(ctx context.Context) error {
	r.logger.Debug("Deleting expired recovery challenges")

	deletedCount := 0
	now := time.Now()

	// Iterate through all challenges
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, challengeKeyPrefix) {
			return nil // Skip non-challenge keys
		}

		// Deserialize challenge
		challenge, err := r.deserializeChallenge(value)
		if err != nil {
			r.logger.Error("Failed to deserialize challenge during cleanup",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if expired
		if challenge.ExpiresAt.Before(now) {
			// Delete expired challenge
			if err := r.dbClient.Delete(keyStr); err != nil {
				r.logger.Error("Failed to delete expired challenge",
					zap.String("key", keyStr),
					zap.Error(err))
				return nil // Continue iteration despite error
			}
			deletedCount++
			r.logger.Debug("Deleted expired recovery challenge",
				zap.String("challengeID", challenge.ChallengeID.String()),
				zap.Time("expiredAt", challenge.ExpiresAt))
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery challenges for cleanup", zap.Error(err))
		return errors.NewAppError("failed to cleanup expired challenges", err)
	}

	r.logger.Info("Successfully cleaned up expired recovery challenges",
		zap.Int("deletedCount", deletedCount))

	return nil
}

// Helper methods for challenges

func (r *recoveryRepository) generateChallengeKey(challengeID string) string {
	return fmt.Sprintf("%s%s", challengeKeyPrefix, challengeID)
}

func (r *recoveryRepository) serializeChallenge(challenge *recovery.RecoveryChallenge) ([]byte, error) {
	return challenge.Serialize() // This method would need to be implemented in the domain model
}

func (r *recoveryRepository) deserializeChallenge(data []byte) (*recovery.RecoveryChallenge, error) {
	return recovery.NewRecoveryChallengeFromDeserialized(data) // This method would need to be implemented in the domain model
}
