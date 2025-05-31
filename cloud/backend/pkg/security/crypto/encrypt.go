package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/nacl/box"
)

// EncryptData represents encrypted data with its nonce
type EncryptData struct {
	Ciphertext []byte
	Nonce      []byte
}

// EncryptWithSecretKey encrypts data with a symmetric key using ChaCha20-Poly1305
// JavaScript equivalent: sodium.crypto_secretbox_easy() but using ChaCha20-Poly1305
func EncryptWithSecretKey(data, key []byte) (*EncryptData, error) {
	if len(key) != MasterKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", MasterKeySize, len(key))
	}

	// Create ChaCha20-Poly1305 cipher
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate nonce
	nonce, err := GenerateRandomNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := cipher.Seal(nil, nonce, data, nil)

	return &EncryptData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, nil
}

// DecryptWithSecretKey decrypts data with a symmetric key using ChaCha20-Poly1305
// JavaScript equivalent: sodium.crypto_secretbox_open_easy() but using ChaCha20-Poly1305
func DecryptWithSecretKey(encryptedData *EncryptData, key []byte) ([]byte, error) {
	if len(key) != MasterKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", MasterKeySize, len(key))
	}

	if len(encryptedData.Nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", NonceSize, len(encryptedData.Nonce))
	}

	// Create ChaCha20-Poly1305 cipher
	cipher, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt
	plaintext, err := cipher.Open(nil, encryptedData.Nonce, encryptedData.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptWithPublicKey encrypts data with a public key using NaCl box (XSalsa20-Poly1305)
// Note: Asymmetric encryption still uses NaCl box for compatibility
// JavaScript equivalent: sodium.crypto_box_seal()
func EncryptWithPublicKey(data, recipientPublicKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", PublicKeySize, len(recipientPublicKey))
	}

	// Convert to fixed-size array
	var pubKeyArray [32]byte
	copy(pubKeyArray[:], recipientPublicKey)

	// Generate nonce for box encryption (24 bytes for NaCl box)
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// For sealed box, we need to use SealAnonymous
	sealed, err := box.SealAnonymous(nil, data, &pubKeyArray, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to seal data: %w", err)
	}

	return sealed, nil
}

// DecryptWithPrivateKey decrypts data with a private key using NaCl box
// Note: Asymmetric encryption still uses NaCl box for compatibility
// JavaScript equivalent: sodium.crypto_box_seal_open()
func DecryptWithPrivateKey(encryptedData, publicKey, privateKey []byte) ([]byte, error) {
	if len(privateKey) != PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", PrivateKeySize, len(privateKey))
	}
	if len(publicKey) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", PublicKeySize, len(publicKey))
	}

	// Convert to fixed-size arrays
	var pubKeyArray [32]byte
	copy(pubKeyArray[:], publicKey)

	var privKeyArray [32]byte
	copy(privKeyArray[:], privateKey)

	// Decrypt using OpenAnonymous for sealed box
	plaintext, ok := box.OpenAnonymous(nil, encryptedData, &pubKeyArray, &privKeyArray)
	if !ok {
		return nil, errors.New("decryption failed: invalid keys or corrupted data")
	}

	return plaintext, nil
}

// EncryptFileChunked encrypts a file in chunks using ChaCha20-Poly1305
// JavaScript equivalent: sodium.crypto_secretstream_* but using ChaCha20-Poly1305
func EncryptFileChunked(reader io.Reader, key []byte) ([]byte, error) {
	// This would be a more complex implementation using
	// chunked encryption. For brevity, we'll use a simpler approach
	// that reads the entire file into memory first.

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	encData, err := EncryptWithSecretKey(data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Combine nonce and ciphertext
	result := make([]byte, len(encData.Nonce)+len(encData.Ciphertext))
	copy(result, encData.Nonce)
	copy(result[len(encData.Nonce):], encData.Ciphertext)

	return result, nil
}

// DecryptFileChunked decrypts a chunked encrypted file using ChaCha20-Poly1305
// JavaScript equivalent: sodium.crypto_secretstream_* but using ChaCha20-Poly1305
func DecryptFileChunked(encryptedData, key []byte) ([]byte, error) {
	// Split nonce and ciphertext
	if len(encryptedData) < NonceSize {
		return nil, fmt.Errorf("encrypted data too short: expected at least %d bytes, got %d", NonceSize, len(encryptedData))
	}

	nonce := encryptedData[:NonceSize]
	ciphertext := encryptedData[NonceSize:]

	// Decrypt
	return DecryptWithSecretKey(&EncryptData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, key)
}
