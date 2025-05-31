// monorepo/native/desktop/maplefile-cli/internal/service/auth/verification_transformer.go
package auth

import (
	"encoding/base64"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// userVerificationDataTransformer implements auth.UserVerificationDataTransformer
type userVerificationDataTransformer struct{}

// NewUserVerificationDataTransformer creates a new transformer
func NewUserVerificationDataTransformer() auth.UserVerificationDataTransformer {
	return &userVerificationDataTransformer{}
}

// UpdateUserWithVerificationData updates a user model with verification data
func (t *userVerificationDataTransformer) UpdateUserWithVerificationData(user *user.User, data *auth.VerifyLoginOTTResponse) error {
	if user == nil {
		return fmt.Errorf("❌ user cannot be nil")
	}

	if data == nil {
		return fmt.Errorf("❌ verification data cannot be nil")
	}

	// Store Encrypted Challenge
	encryptedChallengeBytes, err := base64.StdEncoding.DecodeString(data.EncryptedChallenge)
	if err != nil {
		encryptedChallengeBytes, err = base64.RawURLEncoding.DecodeString(data.EncryptedChallenge)
		if err != nil {
			return fmt.Errorf("💔 error decoding encrypted challenge: %v", err)
		}
	}
	user.EncryptedChallenge = encryptedChallengeBytes

	// Store Salt
	salt, err := base64.StdEncoding.DecodeString(data.Salt)
	if err != nil {
		salt, err = base64.RawURLEncoding.DecodeString(data.Salt)
		if err != nil {
			return fmt.Errorf("🔑💔 error decoding password salt: %v", err)
		}
	}
	user.PasswordSalt = salt

	// Store Public Key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(data.PublicKey)
	if err != nil {
		publicKeyBytes, err = base64.RawURLEncoding.DecodeString(data.PublicKey)
		if err != nil {
			return fmt.Errorf("🔑💔 error decoding public key: %v", err)
		}
	}
	user.PublicKey.Key = publicKeyBytes

	// Store Encrypted Master Key - Updated for ChaCha20-Poly1305 (12-byte nonces)
	encMasterKeyBytes, err := base64.StdEncoding.DecodeString(data.EncryptedMasterKey)
	if err == nil && len(encMasterKeyBytes) >= crypto.ChaCha20Poly1305NonceSize {
		// Use ChaCha20-Poly1305 nonce size (12 bytes)
		nonceSize := crypto.ChaCha20Poly1305NonceSize
		user.EncryptedMasterKey.Nonce = encMasterKeyBytes[:nonceSize]
		user.EncryptedMasterKey.Ciphertext = encMasterKeyBytes[nonceSize:]
	} else {
		return fmt.Errorf("🔑💔 error decoding encrypted master key: %v", err)
	}

	// Store Encrypted Private Key - Updated for ChaCha20-Poly1305 (12-byte nonces)
	encPrivateKeyBytes, err := base64.StdEncoding.DecodeString(data.EncryptedPrivateKey)
	if err == nil && len(encPrivateKeyBytes) >= crypto.ChaCha20Poly1305NonceSize {
		// Use ChaCha20-Poly1305 nonce size (12 bytes)
		nonceSize := crypto.ChaCha20Poly1305NonceSize
		user.EncryptedPrivateKey.Nonce = encPrivateKeyBytes[:nonceSize]
		user.EncryptedPrivateKey.Ciphertext = encPrivateKeyBytes[nonceSize:]
	} else {
		return fmt.Errorf("🔑💔 error decoding encrypted private key: %v", err)
	}

	// Store ChallengeID
	user.VerificationID = data.ChallengeID

	// Store extra security data such as KDF parameters and key rotation policy.
	user.KDFParams = data.KDFParams
	user.LastPasswordChange = data.LastPasswordChange
	user.KDFParamsNeedUpgrade = data.KDFParamsNeedUpgrade
	user.CurrentKeyVersion = data.CurrentKeyVersion
	user.LastKeyRotation = data.LastKeyRotation
	user.KeyRotationPolicy = data.KeyRotationPolicy

	return nil
}
