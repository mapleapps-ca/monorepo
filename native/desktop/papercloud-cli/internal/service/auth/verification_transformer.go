// monorepo/native/desktop/papercloud-cli/internal/service/auth/verification_transformer.go
package auth

import (
	"encoding/base64"
	"fmt"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/user"
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
		return fmt.Errorf("user cannot be nil")
	}

	if data == nil {
		return fmt.Errorf("verification data cannot be nil")
	}

	// Store Encrypted Challenge
	encryptedChallengeBytes, err := base64.StdEncoding.DecodeString(data.EncryptedChallenge)
	if err != nil {
		encryptedChallengeBytes, err = base64.RawURLEncoding.DecodeString(data.EncryptedChallenge)
		if err != nil {
			return fmt.Errorf("error decoding encrypted challenge: %v", err)
		}
	}
	user.EncryptedChallenge = encryptedChallengeBytes

	// Store Salt
	salt, err := base64.StdEncoding.DecodeString(data.Salt)
	if err != nil {
		salt, err = base64.RawURLEncoding.DecodeString(data.Salt)
		if err != nil {
			return fmt.Errorf("error decoding password salt: %v", err)
		}
	}
	user.PasswordSalt = salt

	// Store Public Key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(data.PublicKey)
	if err != nil {
		publicKeyBytes, err = base64.RawURLEncoding.DecodeString(data.PublicKey)
		if err != nil {
			return fmt.Errorf("error decoding public key: %v", err)
		}
	}
	user.PublicKey.Key = publicKeyBytes

	// Store Encrypted Master Key
	encMasterKeyBytes, err := base64.StdEncoding.DecodeString(data.EncryptedMasterKey)
	if err == nil && len(encMasterKeyBytes) >= 24 {
		// Assuming the first 24 bytes are the nonce
		nonceSize := 24 // sodium.crypto_secretbox_NONCEBYTES
		user.EncryptedMasterKey.Nonce = encMasterKeyBytes[:nonceSize]
		user.EncryptedMasterKey.Ciphertext = encMasterKeyBytes[nonceSize:]
	} else {
		return fmt.Errorf("error decoding encrypted master key: %v", err)
	}

	// Store Encrypted Private Key
	encPrivateKeyBytes, err := base64.StdEncoding.DecodeString(data.EncryptedPrivateKey)
	if err == nil && len(encPrivateKeyBytes) >= 24 {
		nonceSize := 24 // sodium.crypto_secretbox_NONCEBYTES
		user.EncryptedPrivateKey.Nonce = encPrivateKeyBytes[:nonceSize]
		user.EncryptedPrivateKey.Ciphertext = encPrivateKeyBytes[nonceSize:]
	} else {
		return fmt.Errorf("error decoding encrypted private key: %v", err)
	}

	// Store ChallengeID
	user.VerificationID = data.ChallengeID

	return nil
}
