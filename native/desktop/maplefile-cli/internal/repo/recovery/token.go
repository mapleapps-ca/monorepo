// native/desktop/maplefile-cli/internal/repo/recovery/token.go
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

// Token management operations

func (r *recoveryRepository) CreateToken(ctx context.Context, token *recovery.RecoveryToken) error {
	r.logger.Debug("Creating recovery token",
		zap.String("sessionID", token.SessionID.String()),
		zap.String("userID", token.UserID.String()))

	// Validate token
	if err := recovery.ValidateRecoveryTokenEntity(token); err != nil {
		r.logger.Error("Invalid recovery token", zap.Error(err))
		return errors.NewAppError("invalid recovery token", err)
	}

	// Serialize token
	tokenBytes, err := r.serializeToken(token)
	if err != nil {
		r.logger.Error("Failed to serialize recovery token", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery token", err)
	}

	// Generate key using token value
	key := r.generateTokenKey(token.Token)

	// Save to database
	if err := r.dbClient.Set(key, tokenBytes); err != nil {
		r.logger.Error("Failed to save recovery token to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save recovery token to local storage", err)
	}

	r.logger.Info("Successfully created recovery token",
		zap.String("sessionID", token.SessionID.String()),
		zap.String("userID", token.UserID.String()))

	return nil
}

func (r *recoveryRepository) GetTokenByValue(ctx context.Context, tokenValue string) (*recovery.RecoveryToken, error) {
	r.logger.Debug("Getting recovery token by value")

	// Generate key
	key := r.generateTokenKey(tokenValue)

	// Get from database
	tokenBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve recovery token from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve recovery token from local storage", err)
	}

	// Check if token was found
	if tokenBytes == nil {
		r.logger.Debug("Recovery token not found")
		return nil, nil
	}

	// Deserialize token
	token, err := r.deserializeToken(tokenBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize recovery token", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize recovery token", err)
	}

	r.logger.Debug("Successfully retrieved recovery token",
		zap.String("sessionID", token.SessionID.String()),
		zap.String("userID", token.UserID.String()))

	return token, nil
}

func (r *recoveryRepository) GetTokenBySessionID(ctx context.Context, sessionID gocql.UUID) (*recovery.RecoveryToken, error) {
	r.logger.Debug("Getting recovery token by session ID",
		zap.String("sessionID", sessionID.String()))

	var foundToken *recovery.RecoveryToken

	// Iterate through all tokens to find the one with matching session ID
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, tokenKeyPrefix) {
			return nil // Skip non-token keys
		}

		// Deserialize token
		token, err := r.deserializeToken(value)
		if err != nil {
			r.logger.Error("Failed to deserialize token during session search",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if this token belongs to the session
		if token.SessionID == sessionID {
			foundToken = token
			return fmt.Errorf("found") // Use error to break iteration
		}

		return nil
	})

	// Check if we found the token (error "found" is expected)
	if err != nil && err.Error() != "found" {
		r.logger.Error("Error iterating through recovery tokens", zap.Error(err))
		return nil, errors.NewAppError("failed to search recovery tokens", err)
	}

	if foundToken != nil {
		r.logger.Debug("Successfully found recovery token by session ID",
			zap.String("sessionID", sessionID.String()),
			zap.String("userID", foundToken.UserID.String()))
	} else {
		r.logger.Debug("Recovery token not found for session",
			zap.String("sessionID", sessionID.String()))
	}

	return foundToken, nil
}

func (r *recoveryRepository) UpdateToken(ctx context.Context, token *recovery.RecoveryToken) error {
	r.logger.Debug("Updating recovery token",
		zap.String("sessionID", token.SessionID.String()),
		zap.String("userID", token.UserID.String()))

	// Validate token
	if err := recovery.ValidateRecoveryTokenEntity(token); err != nil {
		r.logger.Error("Invalid recovery token for update", zap.Error(err))
		return errors.NewAppError("invalid recovery token", err)
	}

	// Check if token exists
	existing, err := r.GetTokenByValue(ctx, token.Token)
	if err != nil {
		return err
	}
	if existing == nil {
		r.logger.Error("Recovery token not found for update")
		return errors.NewAppError("recovery token not found", recovery.ErrTokenNotFound)
	}

	// Serialize token
	tokenBytes, err := r.serializeToken(token)
	if err != nil {
		r.logger.Error("Failed to serialize recovery token for update", zap.Error(err))
		return errors.NewAppError("failed to serialize recovery token", err)
	}

	// Generate key
	key := r.generateTokenKey(token.Token)

	// Update in database
	if err := r.dbClient.Set(key, tokenBytes); err != nil {
		r.logger.Error("Failed to update recovery token in local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to update recovery token in local storage", err)
	}

	r.logger.Info("Successfully updated recovery token",
		zap.String("sessionID", token.SessionID.String()),
		zap.String("userID", token.UserID.String()))

	return nil
}

func (r *recoveryRepository) DeleteToken(ctx context.Context, tokenValue string) error {
	r.logger.Debug("Deleting recovery token")

	// Generate key
	key := r.generateTokenKey(tokenValue)

	// Delete from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete recovery token from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete recovery token from local storage", err)
	}

	r.logger.Info("Successfully deleted recovery token")

	return nil
}

func (r *recoveryRepository) DeleteExpiredTokens(ctx context.Context) error {
	r.logger.Debug("Deleting expired recovery tokens")

	deletedCount := 0
	now := time.Now()

	// Iterate through all tokens
	err := r.dbClient.Iterate(func(key, value []byte) error {
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, tokenKeyPrefix) {
			return nil // Skip non-token keys
		}

		// Deserialize token
		token, err := r.deserializeToken(value)
		if err != nil {
			r.logger.Error("Failed to deserialize token during cleanup",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Check if expired
		if token.ExpiresAt.Before(now) {
			// Delete expired token
			if err := r.dbClient.Delete(keyStr); err != nil {
				r.logger.Error("Failed to delete expired token",
					zap.String("key", keyStr),
					zap.Error(err))
				return nil // Continue iteration despite error
			}
			deletedCount++
			r.logger.Debug("Deleted expired recovery token",
				zap.String("sessionID", token.SessionID.String()),
				zap.Time("expiredAt", token.ExpiresAt))
		}

		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through recovery tokens for cleanup", zap.Error(err))
		return errors.NewAppError("failed to cleanup expired tokens", err)
	}

	r.logger.Info("Successfully cleaned up expired recovery tokens",
		zap.Int("deletedCount", deletedCount))

	return nil
}

// Helper methods for tokens

func (r *recoveryRepository) generateTokenKey(tokenValue string) string {
	return fmt.Sprintf("%s%s", tokenKeyPrefix, tokenValue)
}

func (r *recoveryRepository) serializeToken(token *recovery.RecoveryToken) ([]byte, error) {
	return token.Serialize() // This method would need to be implemented in the domain model
}

func (r *recoveryRepository) deserializeToken(data []byte) (*recovery.RecoveryToken, error) {
	return recovery.NewRecoveryTokenFromDeserialized(data) // This method would need to be implemented in the domain model
}
