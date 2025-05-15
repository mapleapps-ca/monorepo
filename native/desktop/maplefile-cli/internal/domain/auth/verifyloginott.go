// monorepo/native/desktop/maplefile-cli/internal/domain/auth/verifyloginott.go
package auth

import (
	"context"
)

// VerifyLoginOTTRequest represents the data structure sent to the verify-ott endpoint
type VerifyLoginOTTRequest struct {
	Email string `json:"email"`
	OTT   string `json:"ott"`
}

// VerifyLoginOTTResponse represents the response from the verify-ott API
type VerifyLoginOTTResponse struct {
	Salt                string `json:"salt"`
	PublicKey           string `json:"publicKey"`
	EncryptedMasterKey  string `json:"encryptedMasterKey"`
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	EncryptedChallenge  string `json:"encryptedChallenge"`
	ChallengeID         string `json:"challengeId"`
}

// LoginOTTVerificationRepository defines methods for verifying login OTTs
type LoginOTTVerificationRepository interface {
	VerifyLoginOTT(ctx context.Context, request *VerifyLoginOTTRequest) (*VerifyLoginOTTResponse, error)
}
