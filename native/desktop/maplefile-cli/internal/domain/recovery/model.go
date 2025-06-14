// native/desktop/maplefile-cli/internal/domain/recovery/model.go
package recovery

import (
	"time"

	"github.com/gocql/gocql"
)

// RecoverySession represents an active recovery session
type RecoverySession struct {
	SessionID          gocql.UUID `json:"session_id" bson:"session_id"`
	ChallengeID        gocql.UUID `json:"challenge_id" bson:"challenge_id"`
	Email              string     `json:"email" bson:"email"`
	UserID             gocql.UUID `json:"user_id" bson:"user_id"`
	EncryptedChallenge []byte     `json:"encrypted_challenge" bson:"encrypted_challenge"`
	ExpiresAt          time.Time  `json:"expires_at" bson:"expires_at"`
	IsVerified         bool       `json:"is_verified" bson:"is_verified"`
	CreatedAt          time.Time  `json:"created_at" bson:"created_at"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
}

// RecoveryInitiateRequest represents the request to initiate account recovery
type RecoveryInitiateRequest struct {
	Email  string `json:"email"`
	Method string `json:"method"` // Default: "recovery_key"
}

// RecoveryInitiateResponse represents the response from initiate recovery
type RecoveryInitiateResponse struct {
	SessionID          string `json:"session_id"`
	ChallengeID        string `json:"challenge_id"`
	EncryptedChallenge string `json:"encrypted_challenge"`
	ExpiresIn          int    `json:"expires_in"`
}

// RecoveryVerifyRequest represents the request to verify recovery challenge
type RecoveryVerifyRequest struct {
	SessionID          string `json:"session_id"`
	DecryptedChallenge string `json:"decrypted_challenge"`
}

// RecoveryVerifyResponse represents the response from verify recovery
type RecoveryVerifyResponse struct {
	RecoveryToken                     string `json:"recovery_token"`
	MasterKeyEncryptedWithRecoveryKey string `json:"master_key_encrypted_with_recovery_key"`
	ExpiresIn                         int    `json:"expires_in"`
}

// RecoveryCompleteRequest represents the request to complete account recovery
type RecoveryCompleteRequest struct {
	RecoveryToken                        string `json:"recovery_token"`
	NewSalt                              string `json:"new_salt"`
	NewEncryptedMasterKey                string `json:"new_encrypted_master_key"`
	NewEncryptedPrivateKey               string `json:"new_encrypted_private_key"`
	NewEncryptedRecoveryKey              string `json:"new_encrypted_recovery_key"`
	NewMasterKeyEncryptedWithRecoveryKey string `json:"new_master_key_encrypted_with_recovery_key"`
}

// RecoveryCompleteResponse represents the response from complete recovery
type RecoveryCompleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RecoveryChallenge represents the challenge data structure
type RecoveryChallenge struct {
	ChallengeID gocql.UUID `json:"challenge_id" bson:"challenge_id"`
	SessionID   gocql.UUID `json:"session_id" bson:"session_id"`
	UserID      gocql.UUID `json:"user_id" bson:"user_id"`
	Challenge   []byte     `json:"challenge" bson:"challenge"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at" bson:"expires_at"`
	Used        bool       `json:"used" bson:"used"`
	UsedAt      *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
}

// RecoveryToken represents a recovery token for completing the process
type RecoveryToken struct {
	Token     string     `json:"token" bson:"token"`
	SessionID gocql.UUID `json:"session_id" bson:"session_id"`
	UserID    gocql.UUID `json:"user_id" bson:"user_id"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	ExpiresAt time.Time  `json:"expires_at" bson:"expires_at"`
	Used      bool       `json:"used" bson:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty" bson:"used_at,omitempty"`
}

// RecoveryAttempt represents a recovery attempt for rate limiting
type RecoveryAttempt struct {
	ID          gocql.UUID `json:"id" bson:"_id"`
	Email       string     `json:"email" bson:"email"`
	IPAddress   string     `json:"ip_address" bson:"ip_address"`
	AttemptedAt time.Time  `json:"attempted_at" bson:"attempted_at"`
	Success     bool       `json:"success" bson:"success"`
	Method      string     `json:"method" bson:"method"`
	UserAgent   string     `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
}

// Constants for recovery configuration
const (
	// RecoverySessionExpiryDuration is the default expiry time for recovery sessions
	RecoverySessionExpiryDuration = 10 * time.Minute

	// RecoveryTokenExpiryDuration is the default expiry time for recovery tokens
	RecoveryTokenExpiryDuration = 10 * time.Minute

	// MaxRecoveryAttemptsPerWindow is the maximum number of recovery attempts per time window
	MaxRecoveryAttemptsPerWindow = 5

	// RecoveryAttemptWindow is the time window for rate limiting recovery attempts
	RecoveryAttemptWindow = 15 * time.Minute

	// DefaultRecoveryMethod is the default recovery method
	DefaultRecoveryMethod = "recovery_key"
)

// Recovery method constants
const (
	RecoveryMethodRecoveryKey = "recovery_key"
)

// Recovery session states
const (
	RecoverySessionStatePending   = "pending"
	RecoverySessionStateVerified  = "verified"
	RecoverySessionStateCompleted = "completed"
	RecoverySessionStateExpired   = "expired"
)

// IsExpired checks if the recovery session has expired
func (rs *RecoverySession) IsExpired() bool {
	return time.Now().After(rs.ExpiresAt)
}

// CanVerify checks if the session can be verified
func (rs *RecoverySession) CanVerify() bool {
	return !rs.IsExpired() && !rs.IsVerified
}

// CanComplete checks if the session can be completed
func (rs *RecoverySession) CanComplete() bool {
	return !rs.IsExpired() && rs.IsVerified && rs.CompletedAt == nil
}

// GetState returns the current state of the recovery session
func (rs *RecoverySession) GetState() string {
	if rs.IsExpired() {
		return RecoverySessionStateExpired
	}
	if rs.CompletedAt != nil {
		return RecoverySessionStateCompleted
	}
	if rs.IsVerified {
		return RecoverySessionStateVerified
	}
	return RecoverySessionStatePending
}

// IsExpired checks if the recovery challenge has expired
func (rc *RecoveryChallenge) IsExpired() bool {
	return time.Now().After(rc.ExpiresAt)
}

// CanUse checks if the challenge can be used
func (rc *RecoveryChallenge) CanUse() bool {
	return !rc.IsExpired() && !rc.Used
}

// IsExpired checks if the recovery token has expired
func (rt *RecoveryToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// CanUse checks if the token can be used
func (rt *RecoveryToken) CanUse() bool {
	return !rt.IsExpired() && !rt.Used
}
