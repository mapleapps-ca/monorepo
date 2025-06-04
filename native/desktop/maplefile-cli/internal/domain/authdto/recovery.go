// monorepo/native/desktop/maplefile-cli/internal/domain/authdto/recovery.go
package authdto

import (
	"context"
	"time"
)

// RecoveryRequest represents a request to recover an account using a recovery key
type RecoveryRequest struct {
	Email       string `json:"email"`
	RecoveryKey string `json:"recovery_key"` // Base64 encoded recovery key
}

// RecoveryVerifyRequest represents the verification step of recovery
type RecoveryVerifyRequest struct {
	Email       string `json:"email"`
	RecoveryKey string `json:"recovery_key"`
	SessionID   string `json:"session_id"`
}

// RecoveryVerifyResponse contains the encrypted data needed for recovery
type RecoveryVerifyResponse struct {
	SessionID                         string    `json:"session_id"`
	MasterKeyEncryptedWithRecoveryKey string    `json:"master_key_encrypted_with_recovery_key"`
	Salt                              string    `json:"salt"`
	KDFParams                         string    `json:"kdf_params"`
	PublicKey                         string    `json:"public_key"`
	EncryptedPrivateKey               string    `json:"encrypted_private_key"`
	ExpiresAt                         time.Time `json:"expires_at"`
}

// RecoveryCompleteRequest represents the final step to complete recovery with new password
type RecoveryCompleteRequest struct {
	Email               string `json:"email"`
	SessionID           string `json:"session_id"`
	NewPassword         string `json:"new_password"`
	EncryptedMasterKey  string `json:"encrypted_master_key"`  // Master key encrypted with new password
	EncryptedPrivateKey string `json:"encrypted_private_key"` // Private key encrypted with master key
	Salt                string `json:"salt"`                  // New salt for password derivation
}

// RecoveryCompleteResponse represents the response after successful recovery
type RecoveryCompleteResponse struct {
	Success                bool      `json:"success"`
	Message                string    `json:"message"`
	AccessToken            string    `json:"access_token"`
	AccessTokenExpiryTime  time.Time `json:"access_token_expiry_time"`
	RefreshToken           string    `json:"refresh_token"`
	RefreshTokenExpiryTime time.Time `json:"refresh_token_expiry_time"`
}

// RecoveryRepository defines the interface for recovery operations
type RecoveryRepository interface {
	// InitiateRecovery starts the recovery process
	InitiateRecovery(ctx context.Context, request *RecoveryRequest) (*RecoveryVerifyResponse, error)

	// CompleteRecovery completes the recovery process with new password
	CompleteRecovery(ctx context.Context, request *RecoveryCompleteRequest) (*RecoveryCompleteResponse, error)
}
