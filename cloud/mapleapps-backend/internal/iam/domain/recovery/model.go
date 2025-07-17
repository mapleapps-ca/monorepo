// monorepo/cloud/mapleapps-backend/internal/iam/domain/recovery/model.go
package recovery

import (
	"time"

	"github.com/gocql/gocql"
)

// RecoveryMethod represents different ways a user can recover their account
type RecoveryMethod string

const (
	RecoveryMethodRecoveryKey RecoveryMethod = "recovery_key"
	RecoveryMethodSocialKey   RecoveryMethod = "social_key" // Future: social recovery
)

// RecoveryAttempt tracks recovery attempts for security auditing
type RecoveryAttempt struct {
	ID            gocql.UUID     `json:"id" bson:"id"`
	UserID        gocql.UUID     `json:"user_id" bson:"user_id"`
	Email         string         `json:"email" bson:"email"`
	Method        RecoveryMethod `json:"method" bson:"method"`
	IPAddress     string         `json:"ip_address" bson:"ip_address"`
	UserAgent     string         `json:"user_agent" bson:"user_agent"`
	Status        string         `json:"status" bson:"status"` // "initiated", "challenged", "succeeded", "failed"
	FailureReason string         `json:"failure_reason,omitempty" bson:"failure_reason,omitempty"`
	ChallengeID   string         `json:"challenge_id,omitempty" bson:"challenge_id,omitempty"`
	AttemptedAt   time.Time      `json:"attempted_at" bson:"attempted_at"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	ExpiresAt     time.Time      `json:"expires_at" bson:"expires_at"`
}

// RecoverySession represents an active recovery session
type RecoverySession struct {
	SessionID                string         `json:"session_id" bson:"session_id"`
	UserID                   gocql.UUID     `json:"user_id" bson:"user_id"`
	Email                    string         `json:"email" bson:"email"`
	Method                   RecoveryMethod `json:"method" bson:"method"`
	EncryptedChallenge       []byte         `json:"encrypted_challenge" bson:"encrypted_challenge"`
	ChallengeID              string         `json:"challenge_id" bson:"challenge_id"`
	PublicKey                []byte         `json:"public_key" bson:"public_key"`
	EncryptedMasterKey       []byte         `json:"encrypted_master_key" bson:"encrypted_master_key"`
	EncryptedPrivateKey      []byte         `json:"encrypted_private_key" bson:"encrypted_private_key"`
	MasterKeyWithRecoveryKey []byte         `json:"master_key_with_recovery_key" bson:"master_key_with_recovery_key"`
	CreatedAt                time.Time      `json:"created_at" bson:"created_at"`
	ExpiresAt                time.Time      `json:"expires_at" bson:"expires_at"`
	IsVerified               bool           `json:"is_verified" bson:"is_verified"`
	VerifiedAt               *time.Time     `json:"verified_at,omitempty" bson:"verified_at,omitempty"`
}

// RecoveryKeyRotation tracks when recovery keys are rotated
type RecoveryKeyRotation struct {
	ID                 gocql.UUID `json:"id" bson:"id"`
	UserID             gocql.UUID `json:"user_id" bson:"user_id"`
	OldRecoveryKeyHash string     `json:"old_recovery_key_hash" bson:"old_recovery_key_hash"`
	NewRecoveryKeyHash string     `json:"new_recovery_key_hash" bson:"new_recovery_key_hash"`
	RotatedAt          time.Time  `json:"rotated_at" bson:"rotated_at"`
	RotatedBy          gocql.UUID `json:"rotated_by" bson:"rotated_by"`
	Reason             string     `json:"reason" bson:"reason"`
}
