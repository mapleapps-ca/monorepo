// monorepo/native/desktop/maplefile-cli/internal/domain/authdto/verifyloginott.go
package authdto

import (
	"context"
	"time"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
)

// VerifyLoginOTTRequestDTO represents the data structure sent to the verify-ott endpoint
type VerifyLoginOTTRequestDTO struct {
	Email string `json:"email"`
	OTT   string `json:"ott"`
}

// VerifyLoginOTTResponseDTO represents the response from the verify-ott API
type VerifyLoginOTTResponseDTO struct {
	Salt                string         `json:"salt"`
	KDFParams           keys.KDFParams `json:"kdf_params" bson:"kdf_params"`
	PublicKey           string         `json:"publicKey"`
	EncryptedMasterKey  string         `json:"encryptedMasterKey"`
	EncryptedPrivateKey string         `json:"encryptedPrivateKey"`
	EncryptedChallenge  string         `json:"encryptedChallenge"`
	ChallengeID         string         `json:"challengeId"`

	// KDF upgrade and key rotation fields.
	LastPasswordChange   time.Time               `json:"last_password_change" bson:"last_password_change"`
	KDFParamsNeedUpgrade bool                    `json:"kdf_params_need_upgrade" bson:"kdf_params_need_upgrade"`
	CurrentKeyVersion    int                     `json:"current_key_version" bson:"current_key_version"`
	LastKeyRotation      *time.Time              `json:"last_key_rotation,omitempty" bson:"last_key_rotation,omitempty"`
	KeyRotationPolicy    *keys.KeyRotationPolicy `json:"key_rotation_policy,omitempty" bson:"key_rotation_policy,omitempty"`
}

// LoginOTTVerificationRepository defines methods for verifying login OTTs
type LoginOTTVerificationDTORepository interface {
	VerifyLoginOTT(ctx context.Context, request *VerifyLoginOTTRequestDTO) (*VerifyLoginOTTResponseDTO, error)
}
