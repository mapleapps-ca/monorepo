// internal/service/crypto/crypto.go
package crypto

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// CryptoService provides high-level cryptographic operations for the application
type CryptoService interface {
	// User key operations
	DecryptUserMasterKey(ctx context.Context, user *user.User, password string) ([]byte, error)
	DecryptUserPrivateKey(ctx context.Context, user *user.User, masterKey []byte) ([]byte, error)

	// Collection operations
	EncryptCollectionName(ctx context.Context, name string, collectionKey []byte) (string, error)
	DecryptCollectionName(ctx context.Context, encryptedName string, collectionKey []byte) (string, error)
	EncryptCollectionKey(ctx context.Context, collectionKey []byte, masterKey []byte) (*crypto.EncryptedData, error)
	DecryptCollectionKey(ctx context.Context, encryptedKey, nonce, masterKey []byte) ([]byte, error)

	// File operations
	EncryptFileMetadata(ctx context.Context, metadata map[string]interface{}, fileKey []byte) (string, error)
	DecryptFileMetadata(ctx context.Context, encryptedMetadata string, fileKey []byte) (map[string]interface{}, error)
	EncryptFileKey(ctx context.Context, fileKey []byte, collectionKey []byte) (*crypto.EncryptedData, error)
	DecryptFileKey(ctx context.Context, encryptedKey, nonce, collectionKey []byte) ([]byte, error)

	// File content operations
	EncryptFile(ctx context.Context, filePath string, outputPath string, fileKey []byte) error
	DecryptFile(ctx context.Context, encryptedPath string, outputPath string, fileKey []byte) error
}

// cryptoService implements the CryptoService interface
type cryptoService struct {
	logger *zap.Logger
}

// NewCryptoService creates a new crypto service
func NewCryptoService(logger *zap.Logger) CryptoService {
	logger = logger.Named("CryptoService")
	return &cryptoService{
		logger: logger,
	}
}

// DecryptUserMasterKey decrypts the user's master key using their password
func (s *cryptoService) DecryptUserMasterKey(ctx context.Context, user *user.User, password string) ([]byte, error) {
	// Derive key encryption key from password
	kek, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key from password: %w", err)
	}

	// Decrypt master key
	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		kek,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key: %w", err)
	}

	return masterKey, nil
}

// DecryptUserPrivateKey decrypts the user's private key using their master key
func (s *cryptoService) DecryptUserPrivateKey(ctx context.Context, user *user.User, masterKey []byte) ([]byte, error) {
	privateKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedPrivateKey.Ciphertext,
		user.EncryptedPrivateKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return privateKey, nil
}

// EncryptCollectionName encrypts a collection name using the collection key
func (s *cryptoService) EncryptCollectionName(ctx context.Context, name string, collectionKey []byte) (string, error) {
	// Encrypt the name
	encryptedData, err := crypto.EncryptWithSecretBox([]byte(name), collectionKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt collection name: %w", err)
	}

	// Combine nonce and ciphertext
	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)

	// Encode to base64
	return crypto.EncodeToBase64(combined), nil
}

// DecryptCollectionName decrypts a collection name using the collection key
func (s *cryptoService) DecryptCollectionName(ctx context.Context, encryptedName string, collectionKey []byte) (string, error) {
	// Decode from base64
	combined, err := crypto.DecodeFromBase64(encryptedName)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted name: %w", err)
	}

	// Split nonce and ciphertext
	nonce, ciphertext, err := crypto.SplitNonceAndCiphertext(combined, crypto.SecretBoxNonceSize)
	if err != nil {
		return "", fmt.Errorf("failed to split nonce and ciphertext: %w", err)
	}

	// Decrypt
	nameBytes, err := crypto.DecryptWithSecretBox(ciphertext, nonce, collectionKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt collection name: %w", err)
	}

	return string(nameBytes), nil
}

// EncryptCollectionKey encrypts a collection key using the master key
func (s *cryptoService) EncryptCollectionKey(ctx context.Context, collectionKey []byte, masterKey []byte) (*crypto.EncryptedData, error) {
	return crypto.EncryptWithSecretBox(collectionKey, masterKey)
}

// DecryptCollectionKey decrypts a collection key using the master key
func (s *cryptoService) DecryptCollectionKey(ctx context.Context, encryptedKey, nonce, masterKey []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(encryptedKey, nonce, masterKey)
}

// EncryptFileMetadata encrypts file metadata using the file key
func (s *cryptoService) EncryptFileMetadata(ctx context.Context, metadata map[string]interface{}, fileKey []byte) (string, error) {
	// Marshal metadata to JSON
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Encrypt the metadata
	encryptedData, err := crypto.EncryptWithSecretBox(metadataBytes, fileKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt metadata: %w", err)
	}

	// Combine nonce and ciphertext
	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)

	// Encode to base64
	return crypto.EncodeToBase64(combined), nil
}

// DecryptFileMetadata decrypts file metadata using the file key
func (s *cryptoService) DecryptFileMetadata(ctx context.Context, encryptedMetadata string, fileKey []byte) (map[string]interface{}, error) {
	// Decode from base64
	combined, err := crypto.DecodeFromBase64(encryptedMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted metadata: %w", err)
	}

	// Split nonce and ciphertext
	nonce, ciphertext, err := crypto.SplitNonceAndCiphertext(combined, crypto.SecretBoxNonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to split nonce and ciphertext: %w", err)
	}

	// Decrypt
	metadataBytes, err := crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt metadata: %w", err)
	}

	// Unmarshal metadata
	var metadata map[string]interface{}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return metadata, nil
}

// EncryptFileKey encrypts a file key using the collection key
func (s *cryptoService) EncryptFileKey(ctx context.Context, fileKey []byte, collectionKey []byte) (*crypto.EncryptedData, error) {
	return crypto.EncryptWithSecretBox(fileKey, collectionKey)
}

// DecryptFileKey decrypts a file key using the collection key
func (s *cryptoService) DecryptFileKey(ctx context.Context, encryptedKey, nonce, collectionKey []byte) ([]byte, error) {
	return crypto.DecryptWithSecretBox(encryptedKey, nonce, collectionKey)
}

// EncryptFile encrypts a file's content
// TODO: Implement streaming encryption for large files
func (s *cryptoService) EncryptFile(ctx context.Context, filePath string, outputPath string, fileKey []byte) error {
	// This is a placeholder - in production, this should:
	// 1. Read the file in chunks
	// 2. Encrypt each chunk
	// 3. Write to output file
	// 4. Handle large files efficiently
	return fmt.Errorf("file encryption not yet implemented")
}

// DecryptFile decrypts a file's content
// TODO: Implement streaming decryption for large files
func (s *cryptoService) DecryptFile(ctx context.Context, encryptedPath string, outputPath string, fileKey []byte) error {
	// This is a placeholder - in production, this should:
	// 1. Read the encrypted file in chunks
	// 2. Decrypt each chunk
	// 3. Write to output file
	// 4. Handle large files efficiently
	return fmt.Errorf("file decryption not yet implemented")
}
