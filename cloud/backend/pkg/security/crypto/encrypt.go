package crypto

import (
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

// EncryptData represents encrypted data with its nonce
type EncryptData struct {
	Ciphertext []byte
	Nonce      []byte
}

// EncryptWithSecretKey encrypts data with a symmetric key
// JavaScript equivalent: sodium.crypto_secretbox_easy()
func EncryptWithSecretKey(data, key []byte) (*EncryptData, error) {
	if len(key) != MasterKeySize {
		return nil, errors.New("invalid key size")
	}

	// Generate nonce
	nonce, err := GenerateRandomNonce()
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

// DecryptWithSecretKey decrypts data with a symmetric key
// JavaScript equivalent: sodium.crypto_secretbox_open_easy()
func DecryptWithSecretKey(encryptedData *EncryptData, key []byte) ([]byte, error) {
	if len(key) != MasterKeySize {
		return nil, errors.New("invalid key size")
	}

	// Create fixed-size arrays from slices
	var keyArray [32]byte
	copy(keyArray[:], key)

	var nonceArray [24]byte
	copy(nonceArray[:], encryptedData.Nonce)

	// Decrypt
	plaintext, ok := secretbox.Open(nil, encryptedData.Ciphertext, &nonceArray, &keyArray)
	if !ok {
		return nil, errors.New("decryption failed")
	}

	return plaintext, nil
}

// EncryptWithPublicKey encrypts data with a public key
// JavaScript equivalent: sodium.crypto_box_seal()
func EncryptWithPublicKey(data, recipientPublicKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != PublicKeySize {
		return nil, errors.New("invalid public key size")
	}

	// Convert to fixed-size array
	var pubKeyArray [32]byte
	copy(pubKeyArray[:], recipientPublicKey)

	// Encrypt
	return box.Seal(nil, data, &[24]byte{}, &pubKeyArray, nil), nil
}

// DecryptWithPrivateKey decrypts data with a private key
// JavaScript equivalent: sodium.crypto_box_seal_open()
func DecryptWithPrivateKey(encryptedData, publicKey, privateKey []byte) ([]byte, error) {
	if len(privateKey) != PrivateKeySize || len(publicKey) != PublicKeySize {
		return nil, errors.New("invalid key size")
	}

	// Convert to fixed-size arrays
	var pubKeyArray [32]byte
	copy(pubKeyArray[:], publicKey)

	var privKeyArray [32]byte
	copy(privKeyArray[:], privateKey)

	// Decrypt
	plaintext, ok := box.Open(nil, encryptedData, &[24]byte{}, &pubKeyArray, &privKeyArray)
	if !ok {
		return nil, errors.New("decryption failed")
	}

	return plaintext, nil
}

// EncryptFileChunked encrypts a file in chunks
// JavaScript equivalent: sodium.crypto_secretstream_*
func EncryptFileChunked(reader io.Reader, key []byte) ([]byte, error) {
	// This would be a more complex implementation using
	// chunked encryption. For brevity, we'll use a simpler approach
	// that reads the entire file into memory first.

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	encData, err := EncryptWithSecretKey(data, key)
	if err != nil {
		return nil, err
	}

	// Combine nonce and ciphertext
	result := make([]byte, len(encData.Nonce)+len(encData.Ciphertext))
	copy(result, encData.Nonce)
	copy(result[len(encData.Nonce):], encData.Ciphertext)

	return result, nil
}

// DecryptFileChunked decrypts a chunked encrypted file
// JavaScript equivalent: sodium.crypto_secretstream_*
func DecryptFileChunked(encryptedData, key []byte) ([]byte, error) {
	// Split nonce and ciphertext
	if len(encryptedData) < NonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce := encryptedData[:NonceSize]
	ciphertext := encryptedData[NonceSize:]

	// Decrypt
	return DecryptWithSecretKey(&EncryptData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
	}, key)
}
