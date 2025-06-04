// monorepo/native/desktop/maplefile-cli/internal/domain/authdto/loginott.go
package authdto

import (
	"context"
	"time"
)

// LoginOTT represents a one-time token for authentication
type LoginOTTDTO struct {
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginOTTRequest represents the request to get a one-time token
type LoginOTTRequestDTO struct {
	Email string `json:"email"`
}

// LoginOTTResponse represents the server response when requesting a token
type LoginOTTResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LoginOTTDTORepository defines the interface for interacting with login tokens
type LoginOTTDTORepository interface {
	RequestLoginOTTFromCloud(ctx context.Context, request *LoginOTTRequestDTO) (*LoginOTTResponseDTO, error)
}
