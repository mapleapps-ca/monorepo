// monorepo/native/desktop/maplefile-cli/pkg/crypto/crypto.go
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	// Key sizes
	MasterKeySize        = 32 // 256-bit
	KeyEncryptionKeySize = 32
	RecoveryKeySize      = 32
	CollectionKeySize    = 32
	FileKeySize          = 32

	// SecretBox (symmetric encryption) constants
	SecretBoxKeySize   = 32
	SecretBoxNonceSize = 24
	SecretBoxOverhead  = 16

	// Box (asymmetric encryption) constants
	BoxPublicKeySize = 32
	BoxSecretKeySize = 32
	BoxNonceSize     = 24
	BoxOverhead      = box.Overhead
	BoxSealOverhead  = BoxPublicKeySize + BoxOverhead

	// Argon2 parameters - must match between platforms
	Argon2IDAlgorithm = "argon2id"
	Argon2MemLimit    = 4 * 1024 * 1024 // 4 MB (matching your internal/common/crypto settings)
	Argon2OpsLimit    = 1               // 1 iteration (matching your settings)
	Argon2Parallelism = 1
	Argon2KeySize     = 32
	Argon2SaltSize    = 16

	// Encryption algorithm identifiers
	XSalsa20Poly1305Algorithm = "xsalsa20poly1305"
	ChaCha20Poly1305Algorithm = "chacha20poly1305"
)

// EncryptedData represents encrypted data with its nonce
type EncryptedData struct {
	Ciphertext []byte
	Nonce      []byte
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("size must be positive")
	}

	buf := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return buf, nil
}

// GenerateVerificationID creates a human-readable representation of a public key
// JavaScript equivalent: The same BIP39 mnemonic implementation
// Generate VerificationID from public key (deterministic)
func GenerateVerificationID(publicKey []byte) (string, error) {
	if publicKey == nil {
		err := fmt.Errorf("no public key entered")
		log.Printf("pkg.crypto.VerifyVerificationID - Failed to generate verification ID with error: %v\n", err)
		return "", fmt.Errorf("failed to generate verification ID: %w", err)
	}

	// 1. Hash the public key with SHA256
	hash := sha256.Sum256(publicKey[:])

	// 2. Use the hash as entropy for BIP39
	mnemonic, err := bip39.NewMnemonic(hash[:])
	if err != nil {
		log.Printf("pkg.crypto.VerifyVerificationID - Failed to generate verification ID with error: %v\n", err)
		return "", fmt.Errorf("failed to generate verification ID: %w", err)
	}
	return mnemonic, nil
}

// Verify VerificationID matches public key
func VerifyVerificationID(publicKey []byte, verificationID string) bool {
	expectedID, err := GenerateVerificationID(publicKey)
	if err != nil {
		log.Printf("pkg.crypto.VerifyVerificationID - Failed to generate verification ID with error: %v\n", err)
		return false
	}
	return expectedID == verificationID
}

// GenerateKeyPair generates a NaCl box keypair for asymmetric encryption
func GenerateKeyPair() (publicKey []byte, privateKey []byte, verificationID string, err error) {
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Generate deterministic verification ID
	verificationID, err = GenerateVerificationID(pubKey[:])
	if err != nil {
		return nil, nil, "", err
	}

	return pubKey[:], privKey[:], verificationID, nil
}

// DeriveKeyFromPassword derives a key from a password using Argon2id
// This matches the parameters used in your registration and login flows
func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != Argon2SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d, got %d", Argon2SaltSize, len(salt))
	}

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

// EncryptWithSecretBox encrypts data with a symmetric key using NaCl secretbox
func EncryptWithSecretBox(data, key []byte) (*EncryptedData, error) {
	if len(key) != SecretBoxKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", SecretBoxKeySize, len(key))
	}

	// Generate nonce
	nonce, err := GenerateRandomBytes(SecretBoxNonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create fixed-size arrays
	var keyArray [32]byte
	var nonceArray [24]byte
	copy(keyArray[:], key)
	copy(nonceArray[:], nonce)

	// Encrypt
	ciphertext := secretbox.Seal(nil, data, &nonceArray, &keyArray)

	return &EncryptedData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// EncryptDataWithKey is a helper that encrypts data and returns ciphertext and nonce separately
// This is for backward compatibility with existing code
func EncryptDataWithKey(data, key []byte) (ciphertext []byte, nonce []byte, err error) {
	encData, err := EncryptWithSecretBox(data, key)
	if err != nil {
		return nil, nil, err
	}
	return encData.Ciphertext, encData.Nonce, nil
}

// DecryptWithSecretBox decrypts data with a symmetric key using NaCl secretbox
func DecryptWithSecretBox(ciphertext, nonce, key []byte) ([]byte, error) {
	if len(key) != SecretBoxKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", SecretBoxKeySize, len(key))
	}

	if len(nonce) != SecretBoxNonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", SecretBoxNonceSize, len(nonce))
	}

	// Create fixed-size arrays
	var keyArray [32]byte
	var nonceArray [24]byte
	copy(keyArray[:], key)
	copy(nonceArray[:], nonce)

	// Decrypt
	plaintext, ok := secretbox.Open(nil, ciphertext, &nonceArray, &keyArray)
	if !ok {
		return nil, errors.New("failed to decrypt: invalid key, nonce, or corrupted ciphertext")
	}

	return plaintext, nil
}

// EncryptWithBox encrypts data using NaCl box with the recipient's public key and sender's private key
func EncryptWithBox(message []byte, recipientPublicKey, senderPrivateKey []byte) (*EncryptedData, error) {
	if len(recipientPublicKey) != BoxPublicKeySize {
		return nil, fmt.Errorf("recipient public key must be %d bytes", BoxPublicKeySize)
	}
	if len(senderPrivateKey) != BoxSecretKeySize {
		return nil, fmt.Errorf("sender private key must be %d bytes", BoxSecretKeySize)
	}

	// Generate nonce
	nonce, err := GenerateRandomBytes(BoxNonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create fixed-size arrays
	var recipientPubKey [32]byte
	var senderPrivKey [32]byte
	var nonceArray [24]byte
	copy(recipientPubKey[:], recipientPublicKey)
	copy(senderPrivKey[:], senderPrivateKey)
	copy(nonceArray[:], nonce)

	// Encrypt
	ciphertext := box.Seal(nil, message, &nonceArray, &recipientPubKey, &senderPrivKey)

	return &EncryptedData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// DecryptWithBox decrypts data using NaCl box with the sender's public key and recipient's private key
func DecryptWithBox(ciphertext, nonce []byte, senderPublicKey, recipientPrivateKey []byte) ([]byte, error) {
	if len(senderPublicKey) != BoxPublicKeySize {
		return nil, fmt.Errorf("sender public key must be %d bytes", BoxPublicKeySize)
	}
	if len(recipientPrivateKey) != BoxSecretKeySize {
		return nil, fmt.Errorf("recipient private key must be %d bytes", BoxSecretKeySize)
	}
	if len(nonce) != BoxNonceSize {
		return nil, fmt.Errorf("nonce must be %d bytes", BoxNonceSize)
	}

	// Create fixed-size arrays
	var senderPubKey [32]byte
	var recipientPrivKey [32]byte
	var nonceArray [24]byte
	copy(senderPubKey[:], senderPublicKey)
	copy(recipientPrivKey[:], recipientPrivateKey)
	copy(nonceArray[:], nonce)

	// Decrypt
	plaintext, ok := box.Open(nil, ciphertext, &nonceArray, &senderPubKey, &recipientPrivKey)
	if !ok {
		return nil, errors.New("failed to decrypt: invalid keys, nonce, or corrupted ciphertext")
	}

	return plaintext, nil
}

// EncryptWithBoxSeal encrypts data with a recipient's public key using anonymous sender (sealed box)
// This is used for encrypting data where the sender doesn't need to be authenticated
func EncryptWithBoxSeal(message []byte, recipientPublicKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != BoxPublicKeySize {
		return nil, fmt.Errorf("recipient public key must be %d bytes", BoxPublicKeySize)
	}

	// Create a fixed-size array for the recipient's public key
	var recipientPubKey [32]byte
	copy(recipientPubKey[:], recipientPublicKey)

	// Generate an ephemeral keypair
	ephemeralPubKey, ephemeralPrivKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral keypair: %w", err)
	}

	// Generate a random nonce
	nonce, err := GenerateRandomBytes(BoxNonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	var nonceArray [24]byte
	copy(nonceArray[:], nonce)

	// Encrypt the message
	ciphertext := box.Seal(nil, message, &nonceArray, &recipientPubKey, ephemeralPrivKey)

	// Result format: ephemeral_public_key || nonce || ciphertext
	result := make([]byte, BoxPublicKeySize+BoxNonceSize+len(ciphertext))
	copy(result[:BoxPublicKeySize], ephemeralPubKey[:])
	copy(result[BoxPublicKeySize:BoxPublicKeySize+BoxNonceSize], nonce)
	copy(result[BoxPublicKeySize+BoxNonceSize:], ciphertext)

	return result, nil
}

// DecryptWithBoxSeal decrypts data that was encrypted with EncryptWithBoxSeal
func DecryptWithBoxSeal(sealedData []byte, recipientPublicKey, recipientPrivateKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != BoxPublicKeySize {
		return nil, fmt.Errorf("recipient public key must be %d bytes", BoxPublicKeySize)
	}
	if len(recipientPrivateKey) != BoxSecretKeySize {
		return nil, fmt.Errorf("recipient private key must be %d bytes", BoxSecretKeySize)
	}
	if len(sealedData) < BoxPublicKeySize+BoxNonceSize+box.Overhead {
		return nil, errors.New("sealed data too short")
	}

	// Extract components
	ephemeralPublicKey := sealedData[:BoxPublicKeySize]
	nonce := sealedData[BoxPublicKeySize : BoxPublicKeySize+BoxNonceSize]
	ciphertext := sealedData[BoxPublicKeySize+BoxNonceSize:]

	// Create fixed-size arrays
	var ephemeralPubKey [32]byte
	var recipientPrivKey [32]byte
	var nonceArray [24]byte
	copy(ephemeralPubKey[:], ephemeralPublicKey)
	copy(recipientPrivKey[:], recipientPrivateKey)
	copy(nonceArray[:], nonce)

	// Decrypt
	plaintext, ok := box.Open(nil, ciphertext, &nonceArray, &ephemeralPubKey, &recipientPrivKey)
	if !ok {
		return nil, errors.New("failed to decrypt sealed box: invalid keys or corrupted ciphertext")
	}

	return plaintext, nil
}

// DecryptWithBoxAnonymous decrypts data that was encrypted anonymously (without nonce in the data)
// This is used in the login flow for decrypting challenges
func DecryptWithBoxAnonymous(encryptedData []byte, recipientPublicKey, recipientPrivateKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != BoxPublicKeySize {
		return nil, fmt.Errorf("recipient public key must be %d bytes", BoxPublicKeySize)
	}
	if len(recipientPrivateKey) != BoxSecretKeySize {
		return nil, fmt.Errorf("recipient private key must be %d bytes", BoxSecretKeySize)
	}

	// Create fixed-size arrays
	var pubKeyArray, privKeyArray [32]byte
	copy(pubKeyArray[:], recipientPublicKey)
	copy(privKeyArray[:], recipientPrivateKey)

	// Decrypt the sealed box challenge
	decryptedData, ok := box.OpenAnonymous(nil, encryptedData, &pubKeyArray, &privKeyArray)
	if !ok {
		return nil, errors.New("failed to decrypt anonymous box: invalid keys or corrupted data")
	}

	return decryptedData, nil
}

// EncodeToBase64 encodes bytes to base64 standard encoding
func EncodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// EncodeToBase64URL encodes bytes to base64 URL-safe encoding without padding
func EncodeToBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeFromBase64 decodes a base64 standard encoded string to bytes
func DecodeFromBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// DecodeFromBase64URL decodes a base64 URL-safe encoded string without padding to bytes
func DecodeFromBase64URL(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// CombineNonceAndCiphertext combines nonce and ciphertext into a single byte slice
// This is useful for storing encrypted data as a single blob
func CombineNonceAndCiphertext(nonce, ciphertext []byte) []byte {
	combined := make([]byte, len(nonce)+len(ciphertext))
	copy(combined[:len(nonce)], nonce)
	copy(combined[len(nonce):], ciphertext)
	return combined
}

// SplitNonceAndCiphertext splits a combined byte slice into nonce and ciphertext
// Assumes the nonce is SecretBoxNonceSize (24 bytes)
func SplitNonceAndCiphertext(combined []byte, nonceSize int) (nonce []byte, ciphertext []byte, err error) {
	if len(combined) < nonceSize {
		return nil, nil, fmt.Errorf("combined data too short: expected at least %d bytes, got %d", nonceSize, len(combined))
	}

	nonce = combined[:nonceSize]
	ciphertext = combined[nonceSize:]
	return nonce, ciphertext, nil
}

// Helper function to convert EncryptedData to separate slices (for backward compatibility)
func (ed *EncryptedData) Separate() (ciphertext []byte, nonce []byte) {
	return ed.Ciphertext, ed.Nonce
}

// ClearBytes overwrites a byte slice with zeros
// This should be called on sensitive data like keys when they're no longer needed
func ClearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
