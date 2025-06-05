// native/desktop/maplefile-cli/internal/service/filecrypto/encrypt.go
package filecrypto

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// FileEncryptionService handles encryption of file-related data
type FileEncryptionService interface {
	// GenerateFileKeyAndEncryptWithCollectionKey generates a new file key and encrypts it with the collection key
	GenerateFileKeyAndEncryptWithCollectionKey(ctx context.Context, collectionKey []byte) (*keys.EncryptedFileKey, []byte, error)

	// EncryptFileKey encrypts an existing file key with the collection key
	EncryptFileKey(ctx context.Context, fileKey []byte, collectionKey []byte) (*keys.EncryptedFileKey, error)

	// EncryptFileMetadata encrypts file metadata using the file key
	EncryptFileMetadata(ctx context.Context, metadata *dom_file.FileMetadata, fileKey []byte) (string, error)

	// EncryptFileContent encrypts file content using the file key
	EncryptFileContent(ctx context.Context, fileData []byte, fileKey []byte) ([]byte, error)
}

// fileEncryptionService implements FileEncryptionService
type fileEncryptionService struct {
	logger *zap.Logger
}

// NewFileEncryptionService creates a new file encryption service
func NewFileEncryptionService(logger *zap.Logger) FileEncryptionService {
	logger = logger.Named("FileEncryptionService")
	return &fileEncryptionService{
		logger: logger,
	}
}

// GenerateFileKeyAndEncryptWithCollectionKey generates a new file key and encrypts it with the collection key
func (s *fileEncryptionService) GenerateFileKeyAndEncryptWithCollectionKey(ctx context.Context, collectionKey []byte) (*keys.EncryptedFileKey, []byte, error) {
	s.logger.Debug("🔑 Generating new file key and encrypting with collection key")

	if len(collectionKey) == 0 {
		return nil, nil, errors.NewAppError("collection key is required", nil)
	}

	// Generate a new random file key
	fileKey, err := crypto.GenerateRandomBytes(crypto.FileKeySize)
	if err != nil {
		s.logger.Error("❌ Failed to generate file key", zap.Error(err))
		return nil, nil, errors.NewAppError("failed to generate file key", err)
	}

	// Encrypt the file key with the collection key
	encryptedFileKey, err := s.EncryptFileKey(ctx, fileKey, collectionKey)
	if err != nil {
		crypto.ClearBytes(fileKey) // Clear the key if encryption fails
		return nil, nil, err
	}

	s.logger.Debug("✅ Successfully generated and encrypted file key")

	// Note: The caller is responsible for clearing the fileKey when done
	return encryptedFileKey, fileKey, nil
}

// EncryptFileKey encrypts an existing file key with the collection key
func (s *fileEncryptionService) EncryptFileKey(ctx context.Context, fileKey []byte, collectionKey []byte) (*keys.EncryptedFileKey, error) {
	s.logger.Debug("🔑 Encrypting file key with collection key")

	if len(fileKey) == 0 {
		return nil, errors.NewAppError("file key is required", nil)
	}

	if len(collectionKey) == 0 {
		return nil, errors.NewAppError("collection key is required", nil)
	}

	// Encrypt the file key using the collection key
	encryptedData, err := crypto.EncryptWithSecretBox(fileKey, collectionKey)
	if err != nil {
		s.logger.Error("❌ Failed to encrypt file key", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt file key", err)
	}

	// Create the encrypted file key structure
	currentTime := time.Now()
	historicalKey := keys.EncryptedHistoricalKey{
		Ciphertext:    encryptedData.Ciphertext,
		Nonce:         encryptedData.Nonce,
		KeyVersion:    1,
		RotatedAt:     currentTime,
		RotatedReason: "Initial file key creation",
		Algorithm:     crypto.ChaCha20Poly1305Algorithm,
	}

	encryptedFileKey := &keys.EncryptedFileKey{
		Ciphertext:   encryptedData.Ciphertext,
		Nonce:        encryptedData.Nonce,
		KeyVersion:   1,
		RotatedAt:    &currentTime,
		PreviousKeys: []keys.EncryptedHistoricalKey{historicalKey},
	}

	s.logger.Debug("✅ Successfully encrypted file key")
	return encryptedFileKey, nil
}

// EncryptFileMetadata encrypts file metadata using the file key
func (s *fileEncryptionService) EncryptFileMetadata(ctx context.Context, metadata *dom_file.FileMetadata, fileKey []byte) (string, error) {
	s.logger.Debug("🔑 Encrypting file metadata")

	if metadata == nil {
		return "", errors.NewAppError("metadata is required", nil)
	}

	if len(fileKey) == 0 {
		return "", errors.NewAppError("file key is required", nil)
	}

	// Convert metadata to JSON
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		s.logger.Error("❌ Failed to marshal metadata to JSON", zap.Error(err))
		return "", errors.NewAppError("failed to marshal metadata", err)
	}

	// Encrypt the metadata
	encryptedData, err := crypto.EncryptWithSecretBox(metadataBytes, fileKey)
	if err != nil {
		s.logger.Error("❌ Failed to encrypt metadata", zap.Error(err))
		return "", errors.NewAppError("failed to encrypt metadata", err)
	}

	// Combine nonce and ciphertext
	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)

	// Encode to base64
	encryptedMetadata := crypto.EncodeToBase64(combined)

	s.logger.Debug("✅ Successfully encrypted file metadata")
	return encryptedMetadata, nil
}

// EncryptFileContent encrypts file content using the file key
func (s *fileEncryptionService) EncryptFileContent(ctx context.Context, fileData []byte, fileKey []byte) ([]byte, error) {
	s.logger.Debug("🔑 Encrypting file content", zap.Int("dataSize", len(fileData)))

	if len(fileData) == 0 {
		return nil, errors.NewAppError("file data is required", nil)
	}

	if len(fileKey) == 0 {
		return nil, errors.NewAppError("file key is required", nil)
	}

	// Encrypt the file content
	encryptedData, err := crypto.EncryptWithSecretBox(fileData, fileKey)
	if err != nil {
		s.logger.Error("❌ Failed to encrypt file content", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt file content", err)
	}

	// Combine nonce and ciphertext for storage
	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)

	s.logger.Debug("✅ Successfully encrypted file content",
		zap.Int("originalSize", len(fileData)),
		zap.Int("encryptedSize", len(combined)))

	return combined, nil
}
