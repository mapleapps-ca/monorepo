// cloud/mapleapps-backend/internal/iam/domain/recovery/interface.go
package recovery

import (
	"context"
	"time"

	"github.com/gocql/gocql"
)

// RecoveryRepository defines the interface for recovery-related data operations
type RecoveryRepository interface {
	// Recovery attempts
	CreateRecoveryAttempt(ctx context.Context, attempt *RecoveryAttempt) error
	GetRecoveryAttemptByID(ctx context.Context, id gocql.UUID) (*RecoveryAttempt, error)
	GetRecentRecoveryAttempts(ctx context.Context, userID gocql.UUID, limit int) ([]*RecoveryAttempt, error)
	UpdateRecoveryAttempt(ctx context.Context, attempt *RecoveryAttempt) error
	CountFailedAttemptsInWindow(ctx context.Context, email string, window time.Duration) (int, error)

	// Recovery sessions
	CreateRecoverySession(ctx context.Context, session *RecoverySession) error
	GetRecoverySessionByID(ctx context.Context, sessionID string) (*RecoverySession, error)
	UpdateRecoverySession(ctx context.Context, session *RecoverySession) error
	DeleteRecoverySession(ctx context.Context, sessionID string) error
	CleanupExpiredSessions(ctx context.Context) error

	// Recovery key rotations
	CreateRecoveryKeyRotation(ctx context.Context, rotation *RecoveryKeyRotation) error
	GetRecoveryKeyRotationHistory(ctx context.Context, userID gocql.UUID, limit int) ([]*RecoveryKeyRotation, error)
}
