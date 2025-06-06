package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
)

// GenerateRandomKey generates a new random key using crypto_secretbox_keygen
// JavaScript equivalent: sodium.randombytes_buf(crypto.MasterKeySize)
func GenerateRandomKey(size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("key size must be positive")
	}

	key := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

// GenerateKeyPair generates a public/private key pair using NaCl box
// JavaScript equivalent: sodium.crypto_box_keypair()
func GenerateKeyPair() (publicKey, privateKey []byte, verificationID string, err error) {
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Convert from fixed-size arrays to slices
	publicKey = pubKey[:]
	privateKey = privKey[:]

	// Generate deterministic verification ID
	verificationID, err = GenerateVerificationID(publicKey[:])
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to generate verification ID: %w", err)
	}

	return publicKey, privateKey, verificationID, nil
}

// DeriveKeyFromPassword derives a key encryption key from a password using Argon2id
// JavaScript equivalent: sodium.crypto_pwhash()
func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != Argon2SaltSize {
		return nil, fmt.Errorf("invalid salt size: expected %d, got %d", Argon2SaltSize, len(salt))
	}

	// These parameters must match between Go and JavaScript
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

// GenerateRandomNonce generates a random nonce for ChaCha20-Poly1305 encryption operations
// JavaScript equivalent: sodium.randombytes_buf(crypto.NonceSize)
func GenerateRandomNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize) // NonceSize is now 12 for ChaCha20-Poly1305
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random nonce: %w", err)
	}
	return nonce, nil
}

// GenerateVerificationID creates a human-readable representation of a public key
// JavaScript equivalent: The same BIP39 mnemonic implementation
// Generate VerificationID from public key (deterministic)
func GenerateVerificationID(publicKey []byte) (string, error) {
	if len(publicKey) == 0 {
		return "", errors.New("public key cannot be empty")
	}

	// 1. Hash the public key with SHA256
	hash := sha256.Sum256(publicKey)

	// 2. Use the hash as entropy for BIP39
	mnemonic, err := bip39.NewMnemonic(hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to generate verification ID: %w", err)
	}

	return mnemonic, nil
}

// VerifyVerificationID checks if a verification ID matches a public key
func VerifyVerificationID(publicKey []byte, verificationID string) bool {
	expectedID, err := GenerateVerificationID(publicKey)
	if err != nil {
		log.Printf("pkg.crypto.VerifyVerificationID - Failed to generate verification ID with error: %v\n", err)
		return false
	}
	return expectedID == verificationID
}
