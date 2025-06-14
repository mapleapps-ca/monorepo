// native/desktop/maplefile-cli/internal/domain/recoverydto/validation.go
package recoverydto

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
)

// ValidateRecoveryInitiateRequestDTO validates the initiate recovery request
func ValidateRecoveryInitiateRequestDTO(request *RecoveryInitiateRequestDTO) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate email
	if strings.TrimSpace(request.Email) == "" {
		return fmt.Errorf("email is required")
	}

	// Parse and validate email format
	_, err := mail.ParseAddress(request.Email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	// Additional email validation rules
	if len(request.Email) > 254 {
		return fmt.Errorf("email too long (max 254 characters)")
	}

	// Validate method (set default if empty)
	if request.Method == "" {
		request.Method = "recovery_key"
	}

	switch request.Method {
	case "recovery_key":
		// Valid method
	default:
		return fmt.Errorf("unsupported recovery method: %s", request.Method)
	}

	return nil
}

// ValidateRecoveryVerifyRequestDTO validates the verify recovery request
func ValidateRecoveryVerifyRequestDTO(request *RecoveryVerifyRequestDTO) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate session ID
	if strings.TrimSpace(request.SessionID) == "" {
		return fmt.Errorf("session ID is required")
	}

	// Validate decrypted challenge
	if strings.TrimSpace(request.DecryptedChallenge) == "" {
		return fmt.Errorf("decrypted challenge is required")
	}

	// Validate base64 encoding of decrypted challenge
	_, err := base64.StdEncoding.DecodeString(request.DecryptedChallenge)
	if err != nil {
		// Try URL-safe encoding
		_, err = base64.RawURLEncoding.DecodeString(request.DecryptedChallenge)
		if err != nil {
			return fmt.Errorf("decrypted challenge must be valid base64")
		}
	}

	return nil
}

// ValidateRecoveryCompleteRequestDTO validates the complete recovery request
func ValidateRecoveryCompleteRequestDTO(request *RecoveryCompleteRequestDTO) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate recovery token
	if strings.TrimSpace(request.RecoveryToken) == "" {
		return fmt.Errorf("recovery token is required")
	}

	// Validate new salt
	if err := validateBase64Field("new_salt", request.NewSalt); err != nil {
		return err
	}

	// Validate encrypted master key
	if err := validateBase64Field("new_encrypted_master_key", request.NewEncryptedMasterKey); err != nil {
		return err
	}

	// Validate encrypted private key
	if err := validateBase64Field("new_encrypted_private_key", request.NewEncryptedPrivateKey); err != nil {
		return err
	}

	// Validate encrypted recovery key
	if err := validateBase64Field("new_encrypted_recovery_key", request.NewEncryptedRecoveryKey); err != nil {
		return err
	}

	// Validate master key encrypted with recovery key
	if err := validateBase64Field("new_master_key_encrypted_with_recovery_key", request.NewMasterKeyEncryptedWithRecoveryKey); err != nil {
		return err
	}

	return nil
}

// validateBase64Field validates that a field contains valid base64 data
func validateBase64Field(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	// Try standard base64 first
	_, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Try URL-safe base64
		_, err = base64.RawURLEncoding.DecodeString(value)
		if err != nil {
			return fmt.Errorf("%s contains invalid base64 data", fieldName)
		}
	}

	return nil
}
