// internal/service/security/password_validation.go
package security

import (
	"fmt"
	"unicode"
)

// PasswordValidationService provides consistent password validation across all crypto operations
type PasswordValidationService interface {
	ValidateForCryptoOperations(password string) error
	ValidateStrength(password string) error
}

type passwordValidationService struct{}

func NewPasswordValidationService() PasswordValidationService {
	return &passwordValidationService{}
}

// ValidateForCryptoOperations validates password for E2EE operations
func (s *passwordValidationService) ValidateForCryptoOperations(password string) error {
	if password == "" {
		return fmt.Errorf("password is required for E2EE operations")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters for secure E2EE operations")
	}

	return nil
}

// ValidateStrength validates password strength (for registration/changes)
func (s *passwordValidationService) ValidateStrength(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
