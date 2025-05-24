// monorepo/native/desktop/maplefile-cli/internal/usecase/crypto/crypto.go
package crypto

import (
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/auth"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// cryptoUseCase implements the auth.CryptographyOperations interface
type cryptoUseCase struct{}

// NewCryptoUseCase creates a new crypto use-case
func NewCryptoUseCase() auth.CryptographyOperations {
	return &cryptoUseCase{}
}

// DeriveKeyFromPassword derives a key from a password and salt using Argon2id
func (c *cryptoUseCase) DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	return crypto.DeriveKeyFromPassword(password, salt)
}

// DecryptWithSecretBox decrypts data using NaCl secretbox
func (c *cryptoUseCase) DecryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(ciphertext, nonce, key)
}

// DecryptWithBox decrypts data with NaCl box (anonymous sealed box)
func (c *cryptoUseCase) DecryptWithBox(encryptedData, publicKey, privateKey []byte) ([]byte, error) {
	// The login flow uses box.OpenAnonymous, so we use DecryptWithBoxAnonymous
	return crypto.DecryptWithBoxAnonymous(encryptedData, publicKey, privateKey)
}

// EncodeToBase64 encodes a byte slice to base64
func (c *cryptoUseCase) EncodeToBase64(data []byte) string {
	return crypto.EncodeToBase64(data)
}
