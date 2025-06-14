// native/desktop/maplefile-cli/internal/domain/recovery/validation.go
package recovery

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// ValidateRecoveryInitiateRequest validates the initiate recovery request
func ValidateRecoveryInitiateRequest(request *RecoveryInitiateRequest) error {
	if request == nil {
		return NewValidationError("request", "request cannot be nil")
	}

	// Validate email
	if err := ValidateEmail(request.Email); err != nil {
		return err
	}

	// Validate method (set default if empty)
	if request.Method == "" {
		request.Method = DefaultRecoveryMethod
	}

	if err := ValidateRecoveryMethod(request.Method); err != nil {
		return err
	}

	return nil
}

// ValidateRecoveryVerifyRequest validates the verify recovery request
func ValidateRecoveryVerifyRequest(request *RecoveryVerifyRequest) error {
	if request == nil {
		return NewValidationError("request", "request cannot be nil")
	}

	// Validate session ID
	if strings.TrimSpace(request.SessionID) == "" {
		return NewValidationError("session_id", "session ID is required")
	}

	// Validate decrypted challenge
	if err := ValidateDecryptedChallenge(request.DecryptedChallenge); err != nil {
		return err
	}

	return nil
}

// ValidateRecoveryCompleteRequest validates the complete recovery request
func ValidateRecoveryCompleteRequest(request *RecoveryCompleteRequest) error {
	if request == nil {
		return NewValidationError("request", "request cannot be nil")
	}

	// Validate recovery token
	if err := ValidateRecoveryToken(request.RecoveryToken); err != nil {
		return err
	}

	// Validate new salt
	if err := ValidateBase64Field("new_salt", request.NewSalt); err != nil {
		return err
	}

	// Validate encrypted master key
	if err := ValidateBase64Field("new_encrypted_master_key", request.NewEncryptedMasterKey); err != nil {
		return err
	}

	// Validate encrypted private key
	if err := ValidateBase64Field("new_encrypted_private_key", request.NewEncryptedPrivateKey); err != nil {
		return err
	}

	// Validate encrypted recovery key
	if err := ValidateBase64Field("new_encrypted_recovery_key", request.NewEncryptedRecoveryKey); err != nil {
		return err
	}

	// Validate master key encrypted with recovery key
	if err := ValidateBase64Field("new_master_key_encrypted_with_recovery_key", request.NewMasterKeyEncryptedWithRecoveryKey); err != nil {
		return err
	}

	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return NewValidationError("email", "email is required")
	}

	// Parse and validate email format
	_, err := mail.ParseAddress(email)
	if err != nil {
		return NewValidationError("email", "invalid email format")
	}

	// Additional email validation rules
	if len(email) > 254 {
		return NewValidationError("email", "email too long (max 254 characters)")
	}

	return nil
}

// ValidateRecoveryMethod validates the recovery method
func ValidateRecoveryMethod(method string) error {
	if strings.TrimSpace(method) == "" {
		return NewValidationError("method", "recovery method is required")
	}

	switch method {
	case RecoveryMethodRecoveryKey:
		return nil
	default:
		return NewValidationError("method", fmt.Sprintf("unsupported recovery method: %s", method))
	}
}

// ValidateDecryptedChallenge validates a decrypted challenge
func ValidateDecryptedChallenge(challenge string) error {
	if strings.TrimSpace(challenge) == "" {
		return NewValidationError("decrypted_challenge", "decrypted challenge is required")
	}

	// Validate base64 encoding
	if err := ValidateBase64Field("decrypted_challenge", challenge); err != nil {
		return err
	}

	// Decode and check minimum length
	decoded, err := base64.StdEncoding.DecodeString(challenge)
	if err != nil {
		// Try URL-safe encoding
		decoded, err = base64.RawURLEncoding.DecodeString(challenge)
		if err != nil {
			return NewValidationError("decrypted_challenge", "invalid base64 encoding")
		}
	}

	// Check minimum challenge length (32 bytes for security)
	if len(decoded) < 32 {
		return NewValidationError("decrypted_challenge", "challenge too short (minimum 32 bytes)")
	}

	return nil
}

// ValidateRecoveryToken validates a recovery token
func ValidateRecoveryToken(token string) error {
	if strings.TrimSpace(token) == "" {
		return NewValidationError("recovery_token", "recovery token is required")
	}

	// Check token format and length
	if len(token) < 32 {
		return NewValidationError("recovery_token", "recovery token too short")
	}

	return nil
}

// ValidateBase64Field validates that a field contains valid base64 data
func ValidateBase64Field(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(fieldName, fmt.Sprintf("%s is required", fieldName))
	}

	// Try standard base64 first
	_, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Try URL-safe base64
		_, err = base64.RawURLEncoding.DecodeString(value)
		if err != nil {
			return NewValidationError(fieldName, fmt.Sprintf("%s contains invalid base64 data", fieldName))
		}
	}

	return nil
}

// ValidateRecoverySession validates a recovery session
func ValidateRecoverySession(session *RecoverySession) error {
	if session == nil {
		return NewValidationError("session", "recovery session cannot be nil")
	}

	// Validate email
	if err := ValidateEmail(session.Email); err != nil {
		return err
	}

	// Validate session ID
	if session.SessionID.String() == "" {
		return NewValidationError("session_id", "session ID is required")
	}

	// Validate challenge ID
	if session.ChallengeID.String() == "" {
		return NewValidationError("challenge_id", "challenge ID is required")
	}

	// Validate user ID
	if session.UserID.String() == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	// Validate encrypted challenge
	if len(session.EncryptedChallenge) == 0 {
		return NewValidationError("encrypted_challenge", "encrypted challenge is required")
	}

	// Validate timestamps
	if session.CreatedAt.IsZero() {
		return NewValidationError("created_at", "creation timestamp is required")
	}

	if session.ExpiresAt.IsZero() {
		return NewValidationError("expires_at", "expiration timestamp is required")
	}

	if session.ExpiresAt.Before(session.CreatedAt) {
		return NewValidationError("expires_at", "expiration time cannot be before creation time")
	}

	// Validate verification timestamp if verified
	if session.IsVerified && session.VerifiedAt == nil {
		return NewValidationError("verified_at", "verification timestamp is required when session is verified")
	}

	return nil
}

// ValidateRecoveryChallenge validates a recovery challenge
func ValidateRecoveryChallenge(challenge *RecoveryChallenge) error {
	if challenge == nil {
		return NewValidationError("challenge", "recovery challenge cannot be nil")
	}

	// Validate challenge ID
	if challenge.ChallengeID.String() == "" {
		return NewValidationError("challenge_id", "challenge ID is required")
	}

	// Validate session ID
	if challenge.SessionID.String() == "" {
		return NewValidationError("session_id", "session ID is required")
	}

	// Validate user ID
	if challenge.UserID.String() == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	// Validate challenge data
	if len(challenge.Challenge) == 0 {
		return NewValidationError("challenge", "challenge data is required")
	}

	// Validate minimum challenge length
	if len(challenge.Challenge) < 32 {
		return NewValidationError("challenge", "challenge data too short (minimum 32 bytes)")
	}

	// Validate timestamps
	if challenge.CreatedAt.IsZero() {
		return NewValidationError("created_at", "creation timestamp is required")
	}

	if challenge.ExpiresAt.IsZero() {
		return NewValidationError("expires_at", "expiration timestamp is required")
	}

	if challenge.ExpiresAt.Before(challenge.CreatedAt) {
		return NewValidationError("expires_at", "expiration time cannot be before creation time")
	}

	// Validate used timestamp if used
	if challenge.Used && challenge.UsedAt == nil {
		return NewValidationError("used_at", "used timestamp is required when challenge is marked as used")
	}

	return nil
}

// ValidateRecoveryToken validates a recovery token
func ValidateRecoveryTokenEntity(token *RecoveryToken) error {
	if token == nil {
		return NewValidationError("token", "recovery token cannot be nil")
	}

	// Validate token value
	if err := ValidateRecoveryToken(token.Token); err != nil {
		return err
	}

	// Validate session ID
	if token.SessionID.String() == "" {
		return NewValidationError("session_id", "session ID is required")
	}

	// Validate user ID
	if token.UserID.String() == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	// Validate timestamps
	if token.CreatedAt.IsZero() {
		return NewValidationError("created_at", "creation timestamp is required")
	}

	if token.ExpiresAt.IsZero() {
		return NewValidationError("expires_at", "expiration timestamp is required")
	}

	if token.ExpiresAt.Before(token.CreatedAt) {
		return NewValidationError("expires_at", "expiration time cannot be before creation time")
	}

	// Validate used timestamp if used
	if token.Used && token.UsedAt == nil {
		return NewValidationError("used_at", "used timestamp is required when token is marked as used")
	}

	return nil
}

// ValidateRecoveryAttempt validates a recovery attempt
func ValidateRecoveryAttempt(attempt *RecoveryAttempt) error {
	if attempt == nil {
		return NewValidationError("attempt", "recovery attempt cannot be nil")
	}

	// Validate email
	if err := ValidateEmail(attempt.Email); err != nil {
		return err
	}

	// Validate IP address (basic validation)
	if strings.TrimSpace(attempt.IPAddress) == "" {
		return NewValidationError("ip_address", "IP address is required")
	}

	// Validate method
	if err := ValidateRecoveryMethod(attempt.Method); err != nil {
		return err
	}

	// Validate timestamp
	if attempt.AttemptedAt.IsZero() {
		return NewValidationError("attempted_at", "attempt timestamp is required")
	}

	return nil
}

// ValidateTimeWindow validates a time window for rate limiting
func ValidateTimeWindow(since time.Time) error {
	now := time.Now()

	// Check if the time is not in the future
	if since.After(now) {
		return NewValidationError("since", "time cannot be in the future")
	}

	// Check if the time is not too far in the past (e.g., more than 24 hours)
	if since.Before(now.Add(-24 * time.Hour)) {
		return NewValidationError("since", "time window too large (maximum 24 hours)")
	}

	return nil
}

// ValidateRecoveryFilter validates a recovery filter
func ValidateRecoveryFilter(filter *RecoveryFilter) error {
	if filter == nil {
		return nil // Filter is optional
	}

	// Validate email if provided
	if filter.Email != nil {
		if err := ValidateEmail(*filter.Email); err != nil {
			return err
		}
	}

	// Validate method if provided
	if filter.Method != nil {
		if err := ValidateRecoveryMethod(*filter.Method); err != nil {
			return err
		}
	}

	// Validate time range
	if filter.CreatedFrom != nil && filter.CreatedTo != nil {
		if filter.CreatedFrom.After(*filter.CreatedTo) {
			return NewValidationError("created_from", "start time cannot be after end time")
		}
	}

	// Validate pagination
	if filter.Limit < 0 {
		return NewValidationError("limit", "limit cannot be negative")
	}

	if filter.Offset < 0 {
		return NewValidationError("offset", "offset cannot be negative")
	}

	// Set reasonable limits
	if filter.Limit > 1000 {
		return NewValidationError("limit", "limit too large (maximum 1000)")
	}

	return nil
}

// ValidateRecoverySessionFilter validates a recovery session filter
func ValidateRecoverySessionFilter(filter *RecoverySessionFilter) error {
	if filter == nil {
		return nil // Filter is optional
	}

	// Validate email if provided
	if filter.Email != nil {
		if err := ValidateEmail(*filter.Email); err != nil {
			return err
		}
	}

	// Validate state if provided
	if filter.State != nil {
		validStates := []string{
			RecoverySessionStatePending,
			RecoverySessionStateVerified,
			RecoverySessionStateCompleted,
			RecoverySessionStateExpired,
		}

		isValid := false
		for _, validState := range validStates {
			if *filter.State == validState {
				isValid = true
				break
			}
		}

		if !isValid {
			return NewValidationError("state", fmt.Sprintf("invalid state: %s", *filter.State))
		}
	}

	// Validate time range
	if filter.CreatedFrom != nil && filter.CreatedTo != nil {
		if filter.CreatedFrom.After(*filter.CreatedTo) {
			return NewValidationError("created_from", "start time cannot be after end time")
		}
	}

	// Validate pagination
	if filter.Limit < 0 {
		return NewValidationError("limit", "limit cannot be negative")
	}

	if filter.Offset < 0 {
		return NewValidationError("offset", "offset cannot be negative")
	}

	// Set reasonable limits
	if filter.Limit > 1000 {
		return NewValidationError("limit", "limit too large (maximum 1000)")
	}

	return nil
}
