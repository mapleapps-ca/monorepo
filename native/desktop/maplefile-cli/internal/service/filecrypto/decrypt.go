// native/desktop/maplefile-cli/internal/service/filecrypto/decrypt.go
package filecrypto

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// FileDecryptionService handles decryption of file-related data
type FileDecryptionService interface {
	// DecryptFileKey decrypts a file key using the collection key
	DecryptFileKey(ctx context.Context, encryptedFileKey keys.EncryptedFileKey, collectionKey []byte) ([]byte, error)

	// DecryptFileMetadata decrypts file metadata using the file key
	DecryptFileMetadata(ctx context.Context, encryptedMetadata string, fileKey []byte) (*dom_file.FileMetadata, error)

	// DecryptFileContent decrypts file content using the file key
	DecryptFileContent(ctx context.Context, encryptedData []byte, fileKey []byte) ([]byte, error)

	// DecryptFileKeyChain performs the complete chain: collection key -> file key -> decrypted file key
	DecryptFileKeyChain(ctx context.Context, encryptedFileKey keys.EncryptedFileKey, collectionKey []byte) ([]byte, error)
}

// fileDecryptionService implements FileDecryptionService
type fileDecryptionService struct {
	logger *zap.Logger
}

// NewFileDecryptionService creates a new file decryption service
func NewFileDecryptionService(logger *zap.Logger) FileDecryptionService {
	logger = logger.Named("FileDecryptionService")
	return &fileDecryptionService{
		logger: logger,
	}
}

// DecryptFileKey decrypts a file key using the collection key
func (s *fileDecryptionService) DecryptFileKey(ctx context.Context, encryptedFileKey keys.EncryptedFileKey, collectionKey []byte) ([]byte, error) {
	s.logger.Debug("🔑 Decrypting file key with collection key")

	if len(collectionKey) == 0 {
		return nil, errors.NewAppError("collection key is required", nil)
	}

	if len(encryptedFileKey.Ciphertext) == 0 || len(encryptedFileKey.Nonce) == 0 {
		s.logger.Error("❌ Encrypted file key is invalid",
			zap.Int("ciphertextLen", len(encryptedFileKey.Ciphertext)),
			zap.Int("nonceLen", len(encryptedFileKey.Nonce)))
		return nil, errors.NewAppError("encrypted file key is invalid", nil)
	}

	// Decrypt the file key using collection key
	fileKey, err := crypto.DecryptWithSecretBox(
		encryptedFileKey.Ciphertext,
		encryptedFileKey.Nonce,
		collectionKey,
	)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt file key", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt file key: %w", err)
	}

	s.logger.Debug("✅ Successfully decrypted file key")
	return fileKey, nil
}

// DecryptFileMetadata decrypts file metadata using the file key
func (s *fileDecryptionService) DecryptFileMetadata(ctx context.Context, encryptedMetadata string, fileKey []byte) (*dom_file.FileMetadata, error) {
	s.logger.Debug("🔑 Decrypting file metadata")

	if encryptedMetadata == "" {
		return nil, errors.NewAppError("encrypted metadata is required", nil)
	}

	if len(fileKey) == 0 {
		return nil, errors.NewAppError("file key is required", nil)
	}

	// The encrypted metadata is stored as base64 encoded (nonce + ciphertext)
	// Format: base64(12-byte-nonce + ciphertext) for ChaCha20-Poly1305
	combined, err := base64.StdEncoding.DecodeString(encryptedMetadata)
	if err != nil {
		s.logger.Error("❌ Failed to decode encrypted metadata from base64", zap.Error(err))
		return nil, fmt.Errorf("failed to decode encrypted metadata: %w", err)
	}

	// Split nonce and ciphertext for ChaCha20-Poly1305 (12-byte nonce)
	if len(combined) < crypto.ChaCha20Poly1305NonceSize {
		s.logger.Error("❌ Combined data too short",
			zap.Int("expectedMinSize", crypto.ChaCha20Poly1305NonceSize),
			zap.Int("actualSize", len(combined)))
		return nil, fmt.Errorf("combined data too short: expected at least %d bytes for ChaCha20-Poly1305, got %d", crypto.ChaCha20Poly1305NonceSize, len(combined))
	}

	nonce := make([]byte, crypto.ChaCha20Poly1305NonceSize)
	copy(nonce, combined[:crypto.ChaCha20Poly1305NonceSize])

	ciphertext := make([]byte, len(combined)-crypto.ChaCha20Poly1305NonceSize)
	copy(ciphertext, combined[crypto.ChaCha20Poly1305NonceSize:])

	// Decrypt metadata using ChaCha20-Poly1305
	decryptedBytes, err := crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt metadata", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt metadata: %w", err)
	}

	// Parse JSON metadata
	var metadata dom_file.FileMetadata
	if err := json.Unmarshal(decryptedBytes, &metadata); err != nil {
		s.logger.Error("❌ Failed to parse decrypted metadata JSON", zap.Error(err))
		return nil, fmt.Errorf("failed to parse decrypted metadata: %w", err)
	}

	s.logger.Debug("✅ Successfully decrypted file metadata",
		zap.String("fileName", metadata.Name),
		zap.String("mimeType", metadata.MimeType),
		zap.Int64("size", metadata.Size))

	return &metadata, nil
}

// DecryptFileContent decrypts file content using the file key
func (s *fileDecryptionService) DecryptFileContent(ctx context.Context, encryptedData []byte, fileKey []byte) ([]byte, error) {
	s.logger.Debug("🔑 Decrypting file content", zap.Int("encryptedSize", len(encryptedData)))

	if len(encryptedData) == 0 {
		return nil, errors.NewAppError("encrypted data is required", nil)
	}

	if len(fileKey) == 0 {
		return nil, errors.NewAppError("file key is required", nil)
	}

	// The encrypted data should be in the format: nonce (12 bytes) + ciphertext for ChaCha20-Poly1305
	if len(encryptedData) < crypto.ChaCha20Poly1305NonceSize {
		s.logger.Error("❌ Encrypted data too short",
			zap.Int("expectedMinSize", crypto.ChaCha20Poly1305NonceSize),
			zap.Int("actualSize", len(encryptedData)))
		return nil, fmt.Errorf("encrypted data too short: expected at least %d bytes for ChaCha20-Poly1305, got %d",
			crypto.ChaCha20Poly1305NonceSize, len(encryptedData))
	}

	// Extract nonce and ciphertext from combined data
	nonce := make([]byte, crypto.ChaCha20Poly1305NonceSize)
	copy(nonce, encryptedData[:crypto.ChaCha20Poly1305NonceSize])

	ciphertext := make([]byte, len(encryptedData)-crypto.ChaCha20Poly1305NonceSize)
	copy(ciphertext, encryptedData[crypto.ChaCha20Poly1305NonceSize:])

	// Decrypt the content using ChaCha20-Poly1305
	decryptedData, err := crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		s.logger.Error("❌ Failed to decrypt file content", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt file content: %w", err)
	}

	s.logger.Debug("✅ Successfully decrypted file content",
		zap.Int("decryptedSize", len(decryptedData)))

	return decryptedData, nil
}

// DecryptFileKeyChain performs the complete chain: collection key -> file key -> decrypted file key
func (s *fileDecryptionService) DecryptFileKeyChain(ctx context.Context, encryptedFileKey keys.EncryptedFileKey, collectionKey []byte) ([]byte, error) {
	s.logger.Debug("🔗 Starting file key chain decryption")

	// This is just a convenience method that calls DecryptFileKey
	// It's here for consistency and potential future expansion
	return s.DecryptFileKey(ctx, encryptedFileKey, collectionKey)
}
