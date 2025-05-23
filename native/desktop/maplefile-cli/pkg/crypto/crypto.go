// monorepo/native/desktop/maplefile-cli/pkg/crypto/crypto.go
package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// SecretBoxKeySize is the size of a NaCl secretbox key (32 bytes)
	SecretBoxKeySize = 32
	// SecretBoxNonceSize is the size of a NaCl secretbox nonce (24 bytes)
	SecretBoxNonceSize = 24
	// BoxPublicKeySize is the size of a NaCl box public key (32 bytes)
	BoxPublicKeySize = 32
	// BoxSecretKeySize is the size of a NaCl box secret key (32 bytes)
	BoxSecretKeySize = 32
	// BoxNonceSize is the size of a NaCl box nonce (24 bytes)
	BoxNonceSize = 24
	// SaltSize is the size of the salt for password hashing (16 bytes)
	SaltSize = 16

	// Argon2 parameters - must match between platforms
	Argon2IDAlgorithm = "argon2id"
	Argon2MemLimit    = 67108864 // 64 MB
	Argon2OpsLimit    = 4
	Argon2Parallelism = 1
	Argon2KeySize     = 32
	Argon2SaltSize    = 16
)

// KeyParams defines the parameters for key derivation
type KeyParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams returns the default parameters for key derivation
func DefaultParams() *KeyParams {
	return &KeyParams{
		Memory:      64 * 1024,        // 64MB
		Iterations:  3,                // 3 iterations
		Parallelism: 2,                // 2 parallel threads
		SaltLength:  SaltSize,         // 16 byte salt
		KeyLength:   SecretBoxKeySize, // 32 byte key
	}
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

// GenerateKeyPair generates a NaCl box keypair for asymmetric encryption
func GenerateKeyPair() (*[BoxPublicKeySize]byte, *[BoxSecretKeySize]byte, error) {
	return box.GenerateKey(rand.Reader)
}

// DeriveKeyFromPassword derives a key from a password using Argon2
// This matches the crypto_pwhash function in libsodium
func DeriveKeyFromPassword(password string, salt []byte, params *KeyParams) ([]byte, error) {
	if len(salt) != int(params.SaltLength) {
		return nil, errors.New("salt length does not match expected length")
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return key, nil
}

// EncryptWithSecretBox encrypts data using NaCl secretbox (matches crypto_secretbox_easy)
func EncryptWithSecretBox(data []byte, key []byte) ([]byte, []byte, error) {
	if len(key) != SecretBoxKeySize {
		return nil, nil, errors.New("key must be 32 bytes")
	}

	var keyArray [SecretBoxKeySize]byte
	copy(keyArray[:], key)

	// Generate random nonce
	nonce, err := GenerateRandomBytes(SecretBoxNonceSize)
	if err != nil {
		return nil, nil, err
	}

	var nonceArray [SecretBoxNonceSize]byte
	copy(nonceArray[:], nonce)

	// Encrypt
	ciphertext := secretbox.Seal(nil, data, &nonceArray, &keyArray)

	return ciphertext, nonce, nil
}

// DecryptWithSecretBox decrypts data using NaCl secretbox (matches crypto_secretbox_open_easy)
func DecryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error) {
	if len(key) != SecretBoxKeySize {
		return nil, errors.New("key must be 32 bytes")
	}
	if len(nonce) != SecretBoxNonceSize {
		return nil, errors.New("nonce must be 24 bytes")
	}

	var keyArray [SecretBoxKeySize]byte
	var nonceArray [SecretBoxNonceSize]byte
	copy(keyArray[:], key)
	copy(nonceArray[:], nonce)

	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return nil, errors.New("decryption failed, invalid key or ciphertext")
	}

	return plaintext, nil
}

// EncryptWithBoxSeal encrypts data with a recipient's public key
// This implements the functionality of libsodium's crypto_box_seal
func EncryptWithBoxSeal(data []byte, recipientPK []byte) ([]byte, error) {
	if len(recipientPK) != BoxPublicKeySize {
		return nil, errors.New("recipient public key must be 32 bytes")
	}

	// Create a fixed-size array for the recipient's public key
	var recipientPKArray [BoxPublicKeySize]byte
	copy(recipientPKArray[:], recipientPK)

	// Generate an ephemeral keypair
	ephemeralPK, ephemeralSK, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce
	var nonce [BoxNonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// Encrypt the message
	ciphertext := box.Seal(nil, data, &nonce, &recipientPKArray, ephemeralSK)

	// Return the ephemeral public key + nonce + ciphertext
	result := make([]byte, BoxPublicKeySize+BoxNonceSize+len(ciphertext))
	copy(result[:BoxPublicKeySize], ephemeralPK[:])
	copy(result[BoxPublicKeySize:BoxPublicKeySize+BoxNonceSize], nonce[:])
	copy(result[BoxPublicKeySize+BoxNonceSize:], ciphertext)

	return result, nil
}

// DecryptWithBoxSealOpen decrypts data sealed with EncryptWithBoxSeal
func DecryptWithBoxSealOpen(sealedData, publicKey, privateKey []byte) ([]byte, error) {
	if len(publicKey) != BoxPublicKeySize {
		return nil, errors.New("public key must be 32 bytes")
	}
	if len(privateKey) != BoxSecretKeySize {
		return nil, errors.New("private key must be 32 bytes")
	}
	if len(sealedData) < BoxPublicKeySize+BoxNonceSize {
		return nil, errors.New("sealed data too short")
	}

	// Extract components
	ephemeralPK := sealedData[:BoxPublicKeySize]
	nonce := sealedData[BoxPublicKeySize : BoxPublicKeySize+BoxNonceSize]
	ciphertext := sealedData[BoxPublicKeySize+BoxNonceSize:]

	// Convert to fixed-size arrays
	var pubKeyArray, ephemeralPKArray [BoxPublicKeySize]byte
	var privKeyArray [BoxSecretKeySize]byte
	var nonceArray [BoxNonceSize]byte

	copy(pubKeyArray[:], publicKey)
	copy(privKeyArray[:], privateKey)
	copy(ephemeralPKArray[:], ephemeralPK)
	copy(nonceArray[:], nonce)

	// Decrypt
	plaintext, ok := box.Open(nil, ciphertext, &nonceArray, &ephemeralPKArray, &privKeyArray)
	if !ok {
		return nil, errors.New("decryption failed, invalid keys or ciphertext")
	}

	return plaintext, nil
}

// CombineNonceAndCiphertext combines nonce and ciphertext into a single byte slice
// This matches the combineNonceAndCiphertext function in the frontend
func CombineNonceAndCiphertext(nonce, ciphertext []byte) []byte {
	combined := make([]byte, len(nonce)+len(ciphertext))
	copy(combined[:len(nonce)], nonce)
	copy(combined[len(nonce):], ciphertext)
	return combined
}

// SplitNonceAndCiphertext splits a combined byte slice into nonce and ciphertext
// This matches the splitNonceAndCiphertext function in the frontend
func SplitNonceAndCiphertext(combined []byte) ([]byte, []byte, error) {
	if len(combined) < SecretBoxNonceSize {
		return nil, nil, errors.New("combined data too short to contain a nonce")
	}

	nonce := combined[:SecretBoxNonceSize]
	ciphertext := combined[SecretBoxNonceSize:]
	return nonce, ciphertext, nil
}

// ToBase64 encodes bytes to base64 URL-safe string without padding
// This matches the to_base64 function in libsodium with URLSAFE_NO_PADDING variant
func ToBase64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// FromBase64 decodes a base64 URL-safe string without padding to bytes
// This matches the from_base64 function in libsodium with URLSAFE_NO_PADDING variant
func FromBase64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
