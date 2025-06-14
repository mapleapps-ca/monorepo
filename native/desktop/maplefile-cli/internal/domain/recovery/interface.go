// native/desktop/maplefile-cli/internal/domain/recovery/interface.go
package recovery

import (
	"context"
	"time"

	"github.com/gocql/gocql"
)

// RecoveryRepository defines the interface for recovery operations
type RecoveryRepository interface {
	// Session management
	CreateSession(ctx context.Context, session *RecoverySession) error
	GetSessionByID(ctx context.Context, sessionID gocql.UUID) (*RecoverySession, error)
	GetSessionBySessionIDString(ctx context.Context, sessionID string) (*RecoverySession, error)
	UpdateSession(ctx context.Context, session *RecoverySession) error
	DeleteSession(ctx context.Context, sessionID gocql.UUID) error
	DeleteExpiredSessions(ctx context.Context) error

	// Challenge management
	CreateChallenge(ctx context.Context, challenge *RecoveryChallenge) error
	GetChallengeByID(ctx context.Context, challengeID gocql.UUID) (*RecoveryChallenge, error)
	GetChallengeBySessionID(ctx context.Context, sessionID gocql.UUID) (*RecoveryChallenge, error)
	UpdateChallenge(ctx context.Context, challenge *RecoveryChallenge) error
	DeleteChallenge(ctx context.Context, challengeID gocql.UUID) error
	DeleteExpiredChallenges(ctx context.Context) error

	// Token management
	CreateToken(ctx context.Context, token *RecoveryToken) error
	GetTokenByValue(ctx context.Context, tokenValue string) (*RecoveryToken, error)
	GetTokenBySessionID(ctx context.Context, sessionID gocql.UUID) (*RecoveryToken, error)
	UpdateToken(ctx context.Context, token *RecoveryToken) error
	DeleteToken(ctx context.Context, tokenValue string) error
	DeleteExpiredTokens(ctx context.Context) error

	// Attempt tracking for rate limiting
	CreateAttempt(ctx context.Context, attempt *RecoveryAttempt) error
	GetAttemptsByEmail(ctx context.Context, email string, since time.Time) ([]*RecoveryAttempt, error)
	GetAttemptsByIPAddress(ctx context.Context, ipAddress string, since time.Time) ([]*RecoveryAttempt, error)
	CountAttemptsByEmail(ctx context.Context, email string, since time.Time) (int64, error)
	CountAttemptsByIPAddress(ctx context.Context, ipAddress string, since time.Time) (int64, error)
	DeleteOldAttempts(ctx context.Context, before time.Time) error

	// Cleanup operations
	CleanupExpiredData(ctx context.Context) error

	// Transaction support
	OpenTransaction() error
	CommitTransaction() error
	DiscardTransaction()
}

// RecoveryFilter defines filtering options for recovery queries
type RecoveryFilter struct {
	Email       *string     `json:"email,omitempty"`
	UserID      *gocql.UUID `json:"user_id,omitempty"`
	IPAddress   *string     `json:"ip_address,omitempty"`
	Method      *string     `json:"method,omitempty"`
	Success     *bool       `json:"success,omitempty"`
	CreatedFrom *time.Time  `json:"created_from,omitempty"`
	CreatedTo   *time.Time  `json:"created_to,omitempty"`
	Limit       int64       `json:"limit,omitempty"`
	Offset      int64       `json:"offset,omitempty"`
}

// RecoverySessionFilter defines filtering options for recovery sessions
type RecoverySessionFilter struct {
	Email       *string     `json:"email,omitempty"`
	UserID      *gocql.UUID `json:"user_id,omitempty"`
	IsVerified  *bool       `json:"is_verified,omitempty"`
	IsExpired   *bool       `json:"is_expired,omitempty"`
	State       *string     `json:"state,omitempty"`
	CreatedFrom *time.Time  `json:"created_from,omitempty"`
	CreatedTo   *time.Time  `json:"created_to,omitempty"`
	Limit       int64       `json:"limit,omitempty"`
	Offset      int64       `json:"offset,omitempty"`
}

// RecoveryStats represents recovery statistics
type RecoveryStats struct {
	TotalAttempts         int64         `json:"total_attempts"`
	SuccessfulAttempts    int64         `json:"successful_attempts"`
	FailedAttempts        int64         `json:"failed_attempts"`
	ActiveSessions        int64         `json:"active_sessions"`
	ExpiredSessions       int64         `json:"expired_sessions"`
	CompletedSessions     int64         `json:"completed_sessions"`
	AverageCompletionTime time.Duration `json:"average_completion_time"`
}

// RecoveryService defines the business logic interface for recovery operations
type RecoveryService interface {
	// Core recovery flow
	InitiateRecovery(ctx context.Context, request *RecoveryInitiateRequest, ipAddress string, userAgent string) (*RecoveryInitiateResponse, error)
	VerifyRecovery(ctx context.Context, request *RecoveryVerifyRequest) (*RecoveryVerifyResponse, error)
	CompleteRecovery(ctx context.Context, request *RecoveryCompleteRequest) (*RecoveryCompleteResponse, error)

	// Rate limiting and security
	CheckRateLimit(ctx context.Context, email string, ipAddress string) error
	IsRecoveryAllowed(ctx context.Context, email string) (bool, error)

	// Session management
	GetActiveSession(ctx context.Context, sessionID string) (*RecoverySession, error)
	InvalidateSession(ctx context.Context, sessionID string) error

	// Cleanup operations
	CleanupExpiredData(ctx context.Context) error

	// Statistics and monitoring
	GetRecoveryStats(ctx context.Context, filter *RecoveryFilter) (*RecoveryStats, error)
}

// RecoveryValidator defines validation interface for recovery operations
type RecoveryValidator interface {
	ValidateInitiateRequest(request *RecoveryInitiateRequest) error
	ValidateVerifyRequest(request *RecoveryVerifyRequest) error
	ValidateCompleteRequest(request *RecoveryCompleteRequest) error
	ValidateEmail(email string) error
	ValidateRecoveryMethod(method string) error
	ValidateDecryptedChallenge(challenge string) error
	ValidateRecoveryToken(token string) error
	ValidateEncryptedKeys(request *RecoveryCompleteRequest) error
}

// RecoveryCrypto defines cryptographic operations interface for recovery
type RecoveryCrypto interface {
	GenerateChallenge() ([]byte, error)
	EncryptChallenge(challenge []byte, publicKey []byte) ([]byte, error)
	GenerateRecoveryToken() (string, error)
	ValidateDecryptedChallenge(originalChallenge []byte, decryptedChallenge string) bool
	HashRecoveryToken(token string) string
}

// RecoveryNotifier defines notification interface for recovery events
type RecoveryNotifier interface {
	NotifyRecoveryInitiated(ctx context.Context, email string, sessionID string) error
	NotifyRecoveryCompleted(ctx context.Context, email string, success bool) error
	NotifyRateLimitExceeded(ctx context.Context, email string, ipAddress string) error
	NotifySecurityAlert(ctx context.Context, email string, reason string) error
}
