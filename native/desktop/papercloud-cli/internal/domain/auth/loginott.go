// monorepo/native/desktop/papercloud-cli/internal/domain/auth/loginott.go
package auth

import (
	"context"
	"time"
)

// LoginOTT represents a one-time token for authentication
type LoginOTT struct {
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginOTTRequest represents the request to get a one-time token
type LoginOTTRequest struct {
	Email string `json:"email"`
}

// LoginOTTResponse represents the server response when requesting a token
type LoginOTTResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LoginOTTRepository defines the interface for interacting with login tokens
type LoginOTTRepository interface {
	RequestLoginOTT(ctx context.Context, request *LoginOTTRequest) (*LoginOTTResponse, error)
}
