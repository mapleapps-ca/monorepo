// monorepo/cloud/mapleapps-backend/internal/iam/repo/recovery/impl.go
package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/fx"
	"go.uber.org/zap"

	dom_recovery "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/recovery"
)

type Params struct {
	fx.In
	Session *gocql.Session
	Logger  *zap.Logger
}

type recoveryRepository struct {
	session *gocql.Session
	logger  *zap.Logger
}

// NewRepository creates a new Cassandra repository for recovery operations
func NewRepository(p Params) dom_recovery.RecoveryRepository {
	p.Logger = p.Logger.Named("RecoveryRepository")
	return &recoveryRepository{
		session: p.Session,
		logger:  p.Logger,
	}
}

// CreateRecoveryAttempt creates a new recovery attempt record
func (r *recoveryRepository) CreateRecoveryAttempt(ctx context.Context, attempt *dom_recovery.RecoveryAttempt) error {
	if attempt.ID == (gocql.UUID{}) {
		attempt.ID = gocql.TimeUUID()
	}

	query := `
		INSERT INTO iam_recovery_attempts (
			id, user_id, email, method, ip_address, user_agent,
			status, failure_reason, challenge_id, attempted_at,
			completed_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	err := r.session.Query(query,
		attempt.ID, attempt.UserID, attempt.Email, string(attempt.Method),
		attempt.IPAddress, attempt.UserAgent, attempt.Status,
		attempt.FailureReason, attempt.ChallengeID, attempt.AttemptedAt,
		attempt.CompletedAt, attempt.ExpiresAt,
	).WithContext(ctx).Exec()

	if err != nil {
		r.logger.Error("Failed to create recovery attempt",
			zap.String("attempt_id", attempt.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to create recovery attempt: %w", err)
	}

	return nil
}

// GetRecoveryAttemptByID retrieves a recovery attempt by ID
func (r *recoveryRepository) GetRecoveryAttemptByID(ctx context.Context, id gocql.UUID) (*dom_recovery.RecoveryAttempt, error) {
	var attempt dom_recovery.RecoveryAttempt
	var method string

	query := `
		SELECT id, user_id, email, method, ip_address, user_agent,
			   status, failure_reason, challenge_id, attempted_at,
			   completed_at, expires_at
		FROM iam_recovery_attempts
		WHERE id = ?`

	err := r.session.Query(query, id).WithContext(ctx).Scan(
		&attempt.ID, &attempt.UserID, &attempt.Email, &method,
		&attempt.IPAddress, &attempt.UserAgent, &attempt.Status,
		&attempt.FailureReason, &attempt.ChallengeID, &attempt.AttemptedAt,
		&attempt.CompletedAt, &attempt.ExpiresAt,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("Failed to get recovery attempt",
			zap.String("attempt_id", id.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get recovery attempt: %w", err)
	}

	attempt.Method = dom_recovery.RecoveryMethod(method)
	return &attempt, nil
}

// GetRecentRecoveryAttempts retrieves recent recovery attempts for a user
func (r *recoveryRepository) GetRecentRecoveryAttempts(ctx context.Context, userID gocql.UUID, limit int) ([]*dom_recovery.RecoveryAttempt, error) {
	query := `
		SELECT id, user_id, email, method, ip_address, user_agent,
			   status, failure_reason, challenge_id, attempted_at,
			   completed_at, expires_at
		FROM iam_recovery_attempts_by_user_id
		WHERE user_id = ?
		ORDER BY attempted_at DESC
		LIMIT ?`

	iter := r.session.Query(query, userID, limit).WithContext(ctx).Iter()
	defer iter.Close()

	var attempts []*dom_recovery.RecoveryAttempt
	for {
		var attempt dom_recovery.RecoveryAttempt
		var method string

		if !iter.Scan(
			&attempt.ID, &attempt.UserID, &attempt.Email, &method,
			&attempt.IPAddress, &attempt.UserAgent, &attempt.Status,
			&attempt.FailureReason, &attempt.ChallengeID, &attempt.AttemptedAt,
			&attempt.CompletedAt, &attempt.ExpiresAt,
		) {
			break
		}

		attempt.Method = dom_recovery.RecoveryMethod(method)
		attempts = append(attempts, &attempt)
	}

	if err := iter.Close(); err != nil {
		r.logger.Error("Failed to get recent recovery attempts",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get recent recovery attempts: %w", err)
	}

	return attempts, nil
}

// UpdateRecoveryAttempt updates an existing recovery attempt
func (r *recoveryRepository) UpdateRecoveryAttempt(ctx context.Context, attempt *dom_recovery.RecoveryAttempt) error {
	query := `
		UPDATE iam_recovery_attempts
		SET status = ?, failure_reason = ?, completed_at = ?
		WHERE id = ?`

	err := r.session.Query(query,
		attempt.Status, attempt.FailureReason, attempt.CompletedAt, attempt.ID,
	).WithContext(ctx).Exec()

	if err != nil {
		r.logger.Error("Failed to update recovery attempt",
			zap.String("attempt_id", attempt.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to update recovery attempt: %w", err)
	}

	return nil
}

// CountFailedAttemptsInWindow counts failed recovery attempts within a time window
func (r *recoveryRepository) CountFailedAttemptsInWindow(ctx context.Context, email string, window time.Duration) (int, error) {
	startTime := time.Now().Add(-window)

	query := `
		SELECT COUNT(*)
		FROM iam_recovery_attempts_by_email
		WHERE email = ? AND attempted_at >= ? AND status = 'failed'
		ALLOW FILTERING`

	var count int
	err := r.session.Query(query, email, startTime).WithContext(ctx).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count failed attempts",
			zap.String("email", email),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count failed attempts: %w", err)
	}

	return count, nil
}

// CreateRecoverySession creates a new recovery session
func (r *recoveryRepository) CreateRecoverySession(ctx context.Context, session *dom_recovery.RecoverySession) error {
	// Serialize session data to JSON for storage
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal recovery session: %w", err)
	}

	// Store in cache table with TTL
	ttl := int(time.Until(session.ExpiresAt).Seconds())
	if ttl <= 0 {
		return fmt.Errorf("recovery session already expired")
	}

	cacheKey := fmt.Sprintf("recovery_session:%s", session.SessionID)
	query := `
		INSERT INTO pkg_cache_by_key_with_asc_expire_at (key, value, expires_at)
		VALUES (?, ?, ?)
		USING TTL ?`

	err = r.session.Query(query, cacheKey, sessionData, session.ExpiresAt, ttl).
		WithContext(ctx).Exec()

	if err != nil {
		r.logger.Error("Failed to create recovery session",
			zap.String("session_id", session.SessionID),
			zap.Error(err))
		return fmt.Errorf("failed to create recovery session: %w", err)
	}

	return nil
}

// GetRecoverySessionByID retrieves a recovery session by ID
func (r *recoveryRepository) GetRecoverySessionByID(ctx context.Context, sessionID string) (*dom_recovery.RecoverySession, error) {
	cacheKey := fmt.Sprintf("recovery_session:%s", sessionID)

	var sessionData []byte
	query := `SELECT value FROM pkg_cache_by_key_with_asc_expire_at WHERE key = ?`

	err := r.session.Query(query, cacheKey).WithContext(ctx).Scan(&sessionData)
	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("Failed to get recovery session",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get recovery session: %w", err)
	}

	var session dom_recovery.RecoverySession
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recovery session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, nil
	}

	return &session, nil
}

// UpdateRecoverySession updates an existing recovery session
func (r *recoveryRepository) UpdateRecoverySession(ctx context.Context, session *dom_recovery.RecoverySession) error {
	return r.CreateRecoverySession(ctx, session) // Upsert behavior
}

// DeleteRecoverySession deletes a recovery session
func (r *recoveryRepository) DeleteRecoverySession(ctx context.Context, sessionID string) error {
	cacheKey := fmt.Sprintf("recovery_session:%s", sessionID)
	query := `DELETE FROM pkg_cache_by_key_with_asc_expire_at WHERE key = ?`

	err := r.session.Query(query, cacheKey).WithContext(ctx).Exec()
	if err != nil {
		r.logger.Error("Failed to delete recovery session",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return fmt.Errorf("failed to delete recovery session: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired recovery sessions
func (r *recoveryRepository) CleanupExpiredSessions(ctx context.Context) error {
	// The TTL mechanism in Cassandra will automatically remove expired sessions
	// This method is here for interface compliance and manual cleanup if needed
	return nil
}

// CreateRecoveryKeyRotation creates a recovery key rotation record
func (r *recoveryRepository) CreateRecoveryKeyRotation(ctx context.Context, rotation *dom_recovery.RecoveryKeyRotation) error {
	if rotation.ID == (gocql.UUID{}) {
		rotation.ID = gocql.TimeUUID()
	}

	query := `
		INSERT INTO iam_recovery_key_rotations (
			id, user_id, old_recovery_key_hash, new_recovery_key_hash,
			rotated_at, rotated_by, reason
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	err := r.session.Query(query,
		rotation.ID, rotation.UserID, rotation.OldRecoveryKeyHash,
		rotation.NewRecoveryKeyHash, rotation.RotatedAt, rotation.RotatedBy,
		rotation.Reason,
	).WithContext(ctx).Exec()

	if err != nil {
		r.logger.Error("Failed to create recovery key rotation",
			zap.String("rotation_id", rotation.ID.String()),
			zap.Error(err))
		return fmt.Errorf("failed to create recovery key rotation: %w", err)
	}

	return nil
}

// GetRecoveryKeyRotationHistory retrieves recovery key rotation history for a user
func (r *recoveryRepository) GetRecoveryKeyRotationHistory(ctx context.Context, userID gocql.UUID, limit int) ([]*dom_recovery.RecoveryKeyRotation, error) {
	query := `
		SELECT id, user_id, old_recovery_key_hash, new_recovery_key_hash,
			   rotated_at, rotated_by, reason
		FROM iam_recovery_key_rotations_by_user_id
		WHERE user_id = ?
		ORDER BY rotated_at DESC
		LIMIT ?`

	iter := r.session.Query(query, userID, limit).WithContext(ctx).Iter()
	defer iter.Close()

	var rotations []*dom_recovery.RecoveryKeyRotation
	for {
		var rotation dom_recovery.RecoveryKeyRotation

		if !iter.Scan(
			&rotation.ID, &rotation.UserID, &rotation.OldRecoveryKeyHash,
			&rotation.NewRecoveryKeyHash, &rotation.RotatedAt,
			&rotation.RotatedBy, &rotation.Reason,
		) {
			break
		}

		rotations = append(rotations, &rotation)
	}

	if err := iter.Close(); err != nil {
		r.logger.Error("Failed to get recovery key rotation history",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get recovery key rotation history: %w", err)
	}

	return rotations, nil
}
