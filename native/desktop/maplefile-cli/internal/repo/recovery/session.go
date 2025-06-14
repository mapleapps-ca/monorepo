// native/desktop/maplefile-cli/internal/repo/recovery/session.go
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

// Session management operations

func (r *recoveryRepository) CreateSession(ctx context.Context, session *recovery.RecoverySession) error {
	r.logger.Debug("Creating recovery session",
		zap.String("sessionID", session.SessionID.String()),
		zap.String("email", session.Email))

	// Validate session
	if err := recovery.ValidateRecoverySession(session); err != nil {
		r.logger.Error("Invalid recovery session", zap.Error(err))
		return errors.NewAppError("invalid recovery session", err)
	}

	// Serialize session
	sessionBytes, err := r.serializeSession(session)
	if err != nil {
		r.logger.Error("Failed to serialize recovery session", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery session", err)
	}

	// Generate key
	key := r.generateSessionKey(session.SessionID.String())

	// Save to database
	if err := r.dbClient.Set(key, sessionBytes); err != nil {
		r.logger.Error("Failed to save recovery session to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save recovery session to local storage", err)
	}

	r.logger.Info("Successfully created recovery session",
		zap.String("sessionID", session.SessionID.String()),
		zap.String("email", session.Email))

	return nil
}

func (r *recoveryRepository) GetSessionByID(ctx context.Context, sessionID gocql.UUID) (*recovery.RecoverySession, error) {
	r.logger.Debug("Getting recovery session by ID",
		zap.String("sessionID", sessionID.String()))

	// Generate key
	key := r.generateSessionKey(sessionID.String())

	// Get from database
	sessionBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve recovery session from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve recovery session from local storage", err)
	}

	// Check if session was found
	if sessionBytes == nil {
		r.logger.Debug("Recovery session not found",
			zap.String("sessionID", sessionID.String()))
		return nil, nil
	}

	// Deserialize session
	session, err := r.deserializeSession(sessionBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize recovery session", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize recovery session", err)
	}

	r.logger.Debug("Successfully retrieved recovery session",
		zap.String("sessionID", sessionID.String()),
		zap.String("email", session.Email))

	return session, nil
}

func (r *recoveryRepository) GetSessionBySessionIDString(ctx context.Context, sessionID string) (*recovery.RecoverySession, error) {
	r.logger.Debug("Getting recovery session by string ID",
		zap.String("sessionID", sessionID))

	// Parse UUID
	parsedID, err := gocql.ParseUUID(sessionID)
	if err != nil {
		r.logger.Error("Invalid session ID format",
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return nil, errors.NewAppError("invalid session ID format", err)
	}

	return r.GetSessionByID(ctx, parsedID)
}

func (r *recoveryRepository) UpdateSession(ctx context.Context, session *recovery.RecoverySession) error {
	r.logger.Debug("Updating recovery session",
		zap.String("sessionID", session.SessionID.String()),
		zap.String("email", session.Email))

	// Validate session
	if err := recovery.ValidateRecoverySession(session); err != nil {
		r.logger.Error("Invalid recovery session for update", zap.Error(err))
		return errors.NewAppError("invalid recovery session", err)
	}

	// Check if session exists
	existing, err := r.GetSessionByID(ctx, session.SessionID)
	if err != nil {
		return err
	}
	if existing == nil {
		r.logger.Error("Recovery session not found for update",
			zap.String("sessionID", session.SessionID.String()))
		return errors.NewAppError("recovery session not found", recovery.ErrSessionNotFound)
	}

	// Serialize session
	sessionBytes, err := r.serializeSession(session)
	if err != nil {
		r.logger.Error("Failed to serialize recovery session for update", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery session", err)
	}

	// Generate key
	key := r.generateSessionKey(session.SessionID.String())

	// Update in database
	if err := r.dbClient.Set(key, sessionBytes); err != nil {
		r.logger.Error("Failed to update recovery session in local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to update recovery session in local storage", err)
	}

	r.logger.Info("Successfully updated recovery session",
		zap.String("sessionID", session.SessionID.String()),
		zap.String("email", session.Email))

	return nil
}

func (r *recoveryRepository) DeleteSession(ctx context.Context, sessionID gocql.UUID) error {
	r.logger.Debug("Deleting recovery session",
		zap.String("sessionID", sessionID.String()))

	// Generate key
	key := r.generateSessionKey(sessionID.String())

	// Delete from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete recovery session from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete recovery session from local storage", err)
	}

	r.logger.Info("Successfully deleted recovery session",
		zap.String("sessionID", sessionID.String()))

	return nil
}

func (r *recoveryRepository) DeleteExpiredSessions(ctx context.Context) error {
	r.logger.Debug("Deleting expired recovery sessions")

	deletedCount := 0
	now := time.Now()

	// Iterate through all sessions
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, sessionKeyPrefix) {
			return nil // Skip non-session keys
		}

		// Deserialize session
		session, err := r.deserializeSession(value)
		if err != nil {
			r.logger.Error("Failed to deserialize session during cleanup",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if expired
		if session.ExpiresAt.Before(now) {
			// Delete expired session
			if err := r.dbClient.Delete(keyStr); err != nil {
				r.logger.Error("Failed to delete expired session",
					zap.String("key", keyStr),
					zap.Error(err))
				return nil // Continue iteration despite error
			}
			deletedCount++
			r.logger.Debug("Deleted expired recovery session",
				zap.String("sessionID", session.SessionID.String()),
				zap.Time("expiredAt", session.ExpiresAt))
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery sessions for cleanup", zap.Error(err))
		return errors.NewAppError("failed to cleanup expired sessions", err)
	}

	r.logger.Info("Successfully cleaned up expired recovery sessions",
		zap.Int("deletedCount", deletedCount))

	return nil
}

// Helper methods for sessions

func (r *recoveryRepository) generateSessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", sessionKeyPrefix, sessionID)
}

func (r *recoveryRepository) serializeSession(session *recovery.RecoverySession) ([]byte, error) {
	// You can implement JSON or CBOR serialization here
	// For now, using a simple approach similar to your existing code
	return session.Serialize() // This method would need to be implemented in the domain model
}

func (r *recoveryRepository) deserializeSession(data []byte) (*recovery.RecoverySession, error) {
	// Corresponding deserialization method
	return recovery.NewRecoverySessionFromDeserialized(data) // This method would need to be implemented in the domain model
}
