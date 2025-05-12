package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/box"
)

// GenerateRandomKey generates a new random key using crypto_secretbox_keygen
// JavaScript equivalent: sodium.randombytes_buf(crypto.MasterKeySize)
func GenerateRandomKey(size int) ([]byte, error) {
	key := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateKeyPair generates a public/private key pair
// JavaScript equivalent: sodium.crypto_box_keypair()
func GenerateKeyPair() (publicKey, privateKey []byte, err error) {
	pubKey, privKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Convert from fixed-size arrays to slices
	publicKey = pubKey[:]
	privateKey = privKey[:]

	return publicKey, privateKey, nil
}

// DeriveKeyFromPassword derives a key encryption key from a password using Argon2id
// JavaScript equivalent: sodium.crypto_pwhash()
func DeriveKeyFromPassword(password string, salt []byte) ([]byte, error) {
	if len(salt) != Argon2SaltSize {
		return nil, errors.New("invalid salt size")
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

// GenerateRandomNonce generates a random nonce for encryption operations
// JavaScript equivalent: sodium.randombytes_buf(crypto.NonceSize)
func GenerateRandomNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}

// GenerateVerificationID creates a human-readable representation of a public key
// JavaScript equivalent: The same BIP39 mnemonic implementation
func GenerateVerificationID(publicKey []byte) (string, error) {
	// Implementation using a BIP39 mnemonic library
	// This would need to be the same in both Go and JavaScript
	// For demonstration, we'll use a simplified approach:
	return base64.URLEncoding.EncodeToString(publicKey)[:12], nil
}
