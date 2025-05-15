// monorepo/native/desktop/papercloud-cli/internal/service/crypto/crypto.go
package crypto

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/mapleapps-ca/monorepo/native/desktop/papercloud-cli/internal/domain/auth"
)

// Constants for cryptographic operations
const (
	NonceSize     = 24
	KeySize       = 32
	PublicKeySize = 32
	SecretKeySize = 32
)

// cryptoService implements the CryptoService interface
type cryptoService struct{}

// NewCryptoService creates a new crypto service
func NewCryptoService() auth.CryptoService {
	return &cryptoService{}
}

// DeriveKeyFromPassword derives a key from a password and salt using Argon2id
func (s *cryptoService) DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != 16 {
		return nil, fmt.Errorf("salt must be 16 bytes, got %d", len(salt))
	}

	// Use the same Argon2id implementation used in registration
	key := argon2.IDKey(
		[]byte(password),
		salt,
		1,           // Argon2OpsLimit (1 iteration as defined in register.go)
		4*1024*1024, // Argon2MemLimit (4 MB as defined in register.go)
		1,           // Argon2Parallelism
		32,          // Argon2KeySize
	)

	return key, nil
}

// DecryptWithSecretBox decrypts data using NaCl secretbox
func (s *cryptoService) DecryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", KeySize, len(key))
	}

	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", NonceSize, len(nonce))
	}

	var keyArray [KeySize]byte
	var nonceArray [NonceSize]byte

	copy(keyArray[:], key)
	copy(nonceArray[:], nonce)

	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt: invalid key, nonce, or corrupted ciphertext")
	}

	return plaintext, nil
}

// DecryptWithBox decrypts data with NaCl box
func (s *cryptoService) DecryptWithBox(encryptedData, publicKey, privateKey []byte) ([]byte, error) {
	if len(publicKey) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", PublicKeySize, len(publicKey))
	}

	if len(privateKey) != SecretKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", SecretKeySize, len(privateKey))
	}

	var pubKeyArray, privKeyArray [32]byte
	copy(pubKeyArray[:], publicKey)
	copy(privKeyArray[:], privateKey)

	// Decrypt the sealed box challenge
	decryptedData, ok := box.OpenAnonymous(nil, encryptedData, &pubKeyArray, &privKeyArray)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt challenge: invalid keys or corrupted challenge")
	}

	return decryptedData, nil
}

// EncodeToBase64 encodes a byte slice to base64
func (s *cryptoService) EncodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
