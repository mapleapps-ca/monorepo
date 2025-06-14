// native/desktop/maplefile-cli/internal/domain/recoverydto/model.go
package recoverydto

import (
	"time"
)

// RecoveryInitiateRequestDTO represents the request to initiate account recovery
type RecoveryInitiateRequestDTO struct {
	Email  string `json:"email"`
	Method string `json:"method"` // Default: "recovery_key"
}

// RecoveryInitiateResponseDTO represents the response from initiate recovery
type RecoveryInitiateResponseDTO struct {
	SessionID          string `json:"session_id"`
	ChallengeID        string `json:"challenge_id"`
	EncryptedChallenge string `json:"encrypted_challenge"`
	ExpiresIn          int    `json:"expires_in"`
}

// RecoveryVerifyRequestDTO represents the request to verify recovery challenge
type RecoveryVerifyRequestDTO struct {
	SessionID          string `json:"session_id"`
	DecryptedChallenge string `json:"decrypted_challenge"`
}

// RecoveryVerifyResponseDTO represents the response from verify recovery
type RecoveryVerifyResponseDTO struct {
	RecoveryToken                     string `json:"recovery_token"`
	MasterKeyEncryptedWithRecoveryKey string `json:"master_key_encrypted_with_recovery_key"`
	ExpiresIn                         int    `json:"expires_in"`
}

// RecoveryCompleteRequestDTO represents the request to complete account recovery
type RecoveryCompleteRequestDTO struct {
	RecoveryToken                        string `json:"recovery_token"`
	NewSalt                              string `json:"new_salt"`
	NewEncryptedMasterKey                string `json:"new_encrypted_master_key"`
	NewEncryptedPrivateKey               string `json:"new_encrypted_private_key"`
	NewEncryptedRecoveryKey              string `json:"new_encrypted_recovery_key"`
	NewMasterKeyEncryptedWithRecoveryKey string `json:"new_master_key_encrypted_with_recovery_key"`
}

// RecoveryCompleteResponseDTO represents the response from complete recovery
type RecoveryCompleteResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RecoveryStatusResponseDTO represents the status of a recovery session
type RecoveryStatusResponseDTO struct {
	InProgress bool      `json:"in_progress"`
	Email      string    `json:"email,omitempty"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
}
