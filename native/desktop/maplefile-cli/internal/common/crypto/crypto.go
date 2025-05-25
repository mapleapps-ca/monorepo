// internal/common/crypto/crypto.go
package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

// Constants for cryptographic operations
const (
	// Key sizes
	MasterKeySize        = 32 // 256-bit
	KeyEncryptionKeySize = 32
	RecoveryKeySize      = 32

	// Sodium/NaCl constants
	NonceSize         = 24
	PublicKeySize     = 32
	PrivateKeySize    = 32
	SealedBoxOverhead = 16

	// Argon2 parameters - reduced for CLI usage
	Argon2MemLimit    = 4 * 1024 * 1024 // 4 MB
	Argon2OpsLimit    = 1               // 1 iteration
	Argon2Parallelism = 1
	Argon2KeySize     = 32
	Argon2SaltSize    = 16

	FileKeySize       = 32 // 256-bit file encryption keys
	CollectionKeySize = 32 // 256-bit collection encryption keys
)

// EncryptData represents encrypted data with its nonce
type EncryptData struct {
	Ciphertext []byte
	Nonce      []byte
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// DeriveKeyFromPassword derives a key from a password using Argon2id
func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != Argon2SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d, got %d", Argon2SaltSize, len(salt))
	}

	// Use modified parameters for CLI use
	key := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2OpsLimit,
		Argon2MemLimit,
		Argon2Parallelism,
		Argon2KeySize,
	)

	return key, nil
}

// EncryptWithSecretBox encrypts data with a symmetric key
func EncryptWithSecretBox(data, key []byte) (*EncryptData, error) {
	if len(key) != MasterKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", MasterKeySize, len(key))
	}

	// Generate nonce
	nonce, err := GenerateRandomBytes(NonceSize)
	if err != nil {
		return nil, err
	}

	// Create a fixed-size array from slice for secretbox
	var keyArray [32]byte
	copy(keyArray[:], key)

	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Encrypt
	ciphertext := secretbox.Seal(nil, data, &nonceArray, &keyArray)

	return &EncryptData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// EncryptWithBoxSeal encrypts data with a recipient's public key
func EncryptWithBoxSeal(message []byte, recipientPK []byte) ([]byte, error) {
	if len(recipientPK) != PublicKeySize {
		return nil, fmt.Errorf("recipient public key must be %d bytes", PublicKeySize)
	}

	// Create a fixed-size array for the recipient's public key
	var recipientPKArray [32]byte
	copy(recipientPKArray[:], recipientPK)

	// Generate an ephemeral keypair
	ephemeralPK, ephemeralSK, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce
	nonce, err := GenerateRandomBytes(NonceSize)
	if err != nil {
		return nil, err
	}
	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Encrypt the message
	ciphertext := box.Seal(nil, message, &nonceArray, &recipientPKArray, ephemeralSK)

	// Result format: ephemeral_public_key || nonce || ciphertext
	result := make([]byte, PublicKeySize+NonceSize+len(ciphertext))
	copy(result[:PublicKeySize], ephemeralPK[:])
	copy(result[PublicKeySize:PublicKeySize+NonceSize], nonce)
	copy(result[PublicKeySize+NonceSize:], ciphertext)

	return result, nil
}

// DecryptWithSecretBox decrypts data with a symmetric key
func DecryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error) {
	if len(key) != MasterKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", MasterKeySize, len(key))
	}

	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", NonceSize, len(nonce))
	}

	// Create fixed-size arrays from slices for secretbox
	var keyArray [32]byte
	copy(keyArray[:], key)

	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Decrypt
	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}

	return plaintext, nil
}

func ClearBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
