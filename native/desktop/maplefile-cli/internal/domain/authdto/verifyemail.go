// monorepo/native/desktop/maplefile-cli/internal/domain/authdto/verifyemail.go
package authdto

import (
	"context"
	"time"
)

// EmailVerificationDTO represents an email verification request
type EmailVerificationDTO struct {
	Code      string    `json:"code"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// VerifyEmailRequestDTO represents the data needed to verify an email
type VerifyEmailRequestDTO struct {
	Code string `json:"code"`
}

// VerifyEmailResponseDTO represents the expected response from the server
type VerifyEmailResponseDTO struct {
	Message  string `json:"message"`
	UserRole int    `json:"user_role"`
	Status   int    `json:"profile_verification_status,omitempty"`
}

// EmailVerificationDTORepository defines the interface for email verification
type EmailVerificationDTORepository interface {
	VerifyEmail(ctx context.Context, code string) (*VerifyEmailResponseDTO, error)
}
