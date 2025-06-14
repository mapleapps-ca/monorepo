// native/desktop/maplefile-cli/internal/domain/recovery/errors.go
package recovery

import (
	"errors"
	"fmt"
)

// Recovery error types
var (
	// Session errors
	ErrSessionNotFound         = errors.New("recovery session not found")
	ErrSessionExpired          = errors.New("recovery session has expired")
	ErrSessionAlreadyVerified  = errors.New("recovery session already verified")
	ErrSessionAlreadyCompleted = errors.New("recovery session already completed")
	ErrSessionInvalid          = errors.New("recovery session is invalid")
	ErrSessionNotVerified      = errors.New("recovery session not verified")

	// Challenge errors
	ErrChallengeNotFound = errors.New("recovery challenge not found")
	ErrChallengeExpired  = errors.New("recovery challenge has expired")
	ErrChallengeUsed     = errors.New("recovery challenge already used")
	ErrChallengeInvalid  = errors.New("recovery challenge is invalid")
	ErrDecryptionFailed  = errors.New("challenge decryption failed")

	// Token errors
	ErrTokenNotFound = errors.New("recovery token not found")
	ErrTokenExpired  = errors.New("recovery token has expired")
	ErrTokenUsed     = errors.New("recovery token already used")
	ErrTokenInvalid  = errors.New("recovery token is invalid")

	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("recovery rate limit exceeded")
	ErrTooManyAttempts   = errors.New("too many recovery attempts")

	// User errors
	ErrUserNotFound     = errors.New("user not found")
	ErrUserNotEligible  = errors.New("user not eligible for recovery")
	ErrNoRecoveryKeySet = errors.New("no recovery key configured for user")

	// Validation errors
	ErrInvalidEmail     = errors.New("invalid email address")
	ErrInvalidMethod    = errors.New("invalid recovery method")
	ErrInvalidChallenge = errors.New("invalid challenge response")
	ErrInvalidKeys      = errors.New("invalid encrypted keys")
	ErrMissingField     = errors.New("required field is missing")

	// Crypto errors
	ErrEncryptionFailed    = errors.New("encryption failed")
	ErrKeyDerivationFailed = errors.New("key derivation failed")
	ErrInvalidKeyFormat    = errors.New("invalid key format")

	// General errors
	ErrInternalError       = errors.New("internal recovery error")
	ErrOperationNotAllowed = errors.New("recovery operation not allowed")
)

// RecoveryError wraps recovery-specific errors with additional context
type RecoveryError struct {
	Code    string
	Message string
	Cause   error
	Details map[string]interface{}
}

// Error implements the error interface
func (e *RecoveryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *RecoveryError) Unwrap() error {
	return e.Cause
}

// NewRecoveryError creates a new recovery error
func NewRecoveryError(code, message string, cause error) *RecoveryError {
	return &RecoveryError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail field to the error
func (e *RecoveryError) WithDetail(key string, value interface{}) *RecoveryError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error codes for recovery operations
const (
	// Session error codes
	ErrCodeSessionNotFound         = "SESSION_NOT_FOUND"
	ErrCodeSessionExpired          = "SESSION_EXPIRED"
	ErrCodeSessionAlreadyVerified  = "SESSION_ALREADY_VERIFIED"
	ErrCodeSessionAlreadyCompleted = "SESSION_ALREADY_COMPLETED"
	ErrCodeSessionInvalid          = "SESSION_INVALID"
	ErrCodeSessionNotVerified      = "SESSION_NOT_VERIFIED"

	// Challenge error codes
	ErrCodeChallengeNotFound = "CHALLENGE_NOT_FOUND"
	ErrCodeChallengeExpired  = "CHALLENGE_EXPIRED"
	ErrCodeChallengeUsed     = "CHALLENGE_USED"
	ErrCodeChallengeInvalid  = "CHALLENGE_INVALID"
	ErrCodeDecryptionFailed  = "DECRYPTION_FAILED"

	// Token error codes
	ErrCodeTokenNotFound = "TOKEN_NOT_FOUND"
	ErrCodeTokenExpired  = "TOKEN_EXPIRED"
	ErrCodeTokenUsed     = "TOKEN_USED"
	ErrCodeTokenInvalid  = "TOKEN_INVALID"

	// Rate limiting error codes
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeTooManyAttempts   = "TOO_MANY_ATTEMPTS"

	// User error codes
	ErrCodeUserNotFound     = "USER_NOT_FOUND"
	ErrCodeUserNotEligible  = "USER_NOT_ELIGIBLE"
	ErrCodeNoRecoveryKeySet = "NO_RECOVERY_KEY_SET"

	// Validation error codes
	ErrCodeInvalidEmail     = "INVALID_EMAIL"
	ErrCodeInvalidMethod    = "INVALID_METHOD"
	ErrCodeInvalidChallenge = "INVALID_CHALLENGE"
	ErrCodeInvalidKeys      = "INVALID_KEYS"
	ErrCodeMissingField     = "MISSING_FIELD"

	// Crypto error codes
	ErrCodeEncryptionFailed    = "ENCRYPTION_FAILED"
	ErrCodeKeyDerivationFailed = "KEY_DERIVATION_FAILED"
	ErrCodeInvalidKeyFormat    = "INVALID_KEY_FORMAT"

	// General error codes
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeOperationNotAllowed = "OPERATION_NOT_ALLOWED"
)

// Helper functions for creating specific recovery errors

// NewSessionNotFoundError creates a session not found error
func NewSessionNotFoundError(sessionID string) *RecoveryError {
	return NewRecoveryError(
		ErrCodeSessionNotFound,
		"Recovery session not found",
		ErrSessionNotFound,
	).WithDetail("session_id", sessionID)
}

// NewSessionExpiredError creates a session expired error
func NewSessionExpiredError(sessionID string) *RecoveryError {
	return NewRecoveryError(
		ErrCodeSessionExpired,
		"Recovery session has expired",
		ErrSessionExpired,
	).WithDetail("session_id", sessionID)
}

// NewRateLimitExceededError creates a rate limit exceeded error
func NewRateLimitExceededError(email string, attempts int) *RecoveryError {
	return NewRecoveryError(
		ErrCodeRateLimitExceeded,
		"Recovery rate limit exceeded",
		ErrRateLimitExceeded,
	).WithDetail("email", email).WithDetail("attempts", attempts)
}

// NewValidationError creates a validation error
func NewValidationError(field, message string) *RecoveryError {
	return NewRecoveryError(
		ErrCodeMissingField,
		fmt.Sprintf("Validation failed for field '%s': %s", field, message),
		ErrMissingField,
	).WithDetail("field", field).WithDetail("validation_message", message)
}

// NewChallengeDecryptionError creates a challenge decryption error
func NewChallengeDecryptionError() *RecoveryError {
	return NewRecoveryError(
		ErrCodeDecryptionFailed,
		"Failed to decrypt challenge with provided recovery key",
		ErrDecryptionFailed,
	)
}

// NewUserNotEligibleError creates a user not eligible error
func NewUserNotEligibleError(email string, reason string) *RecoveryError {
	return NewRecoveryError(
		ErrCodeUserNotEligible,
		"User not eligible for account recovery",
		ErrUserNotEligible,
	).WithDetail("email", email).WithDetail("reason", reason)
}
