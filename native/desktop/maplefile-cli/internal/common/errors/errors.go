// monorepo/native/desktop/maplefile-cli/internal/common/errors/errors.go
package errors

import "fmt"

// AppError represents an application-specific error
type AppError struct {
	Message string
	Cause   error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(message string, cause error) *AppError {
	return &AppError{
		Message: message,
		Cause:   cause,
	}
}
