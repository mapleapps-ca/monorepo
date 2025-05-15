// monorepo/native/desktop/papercloud-cli/internal/domain/auth/verifyemail.go
package auth

import (
	"context"
	"time"
)

// EmailVerification represents an email verification request
type EmailVerification struct {
	Code      string    `json:"code"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// VerifyEmailRequest represents the data needed to verify an email
type VerifyEmailRequest struct {
	Code string `json:"code"`
}

// VerifyEmailResponse represents the expected response from the server
type VerifyEmailResponse struct {
	Message  string `json:"message"`
	UserRole int    `json:"user_role"`
	Status   int    `json:"profile_verification_status,omitempty"`
}

// EmailVerificationRepository defines the interface for email verification
type EmailVerificationRepository interface {
	VerifyEmail(ctx context.Context, code string) (*VerifyEmailResponse, error)
}
