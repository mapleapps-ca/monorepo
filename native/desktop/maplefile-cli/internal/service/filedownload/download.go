// native/desktop/maplefile-cli/internal/service/filedownload/download.go
package filedownload

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/crypto"
	common_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/filedto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	pkg_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// DecryptedFileMetadata represents decrypted file metadata
type DecryptedFileMetadata struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Created  int64  `json:"created"`
}

// DownloadResult represents the result of a file download with decryption
type DownloadResult struct {
	FileID            primitive.ObjectID     `json:"file_id"`
	DecryptedData     []byte                 `json:"decrypted_data"`
	DecryptedMetadata *DecryptedFileMetadata `json:"decrypted_metadata"`
	ThumbnailData     []byte                 `json:"thumbnail_data,omitempty"`
	OriginalSize      int64                  `json:"original_size"`
	ThumbnailSize     int64                  `json:"thumbnail_size"`
}

// DownloadService handles file download operations with E2EE decryption
type DownloadService interface {
	DownloadAndDecryptFile(ctx context.Context, fileID primitive.ObjectID, userPassword string, urlDuration time.Duration) (*DownloadResult, error)
}

type downloadService struct {
	logger                         *zap.Logger
	getPresignedDownloadURLUseCase filedto.GetPresignedDownloadURLUseCase
	downloadFileUseCase            filedto.DownloadFileUseCase
	getFileUseCase                 uc_file.GetFileUseCase
	getUserByIsLoggedInUseCase     uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase           uc_collection.GetCollectionUseCase
}

func NewDownloadService(
	logger *zap.Logger,
	getPresignedDownloadURLUseCase filedto.GetPresignedDownloadURLUseCase,
	downloadFileUseCase filedto.DownloadFileUseCase,
	getFileUseCase uc_file.GetFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) DownloadService {
	return &downloadService{
		logger:                         logger,
		getPresignedDownloadURLUseCase: getPresignedDownloadURLUseCase,
		downloadFileUseCase:            downloadFileUseCase,
		getFileUseCase:                 getFileUseCase,
		getUserByIsLoggedInUseCase:     getUserByIsLoggedInUseCase,
		getCollectionUseCase:           getCollectionUseCase,
	}
}

func (s *downloadService) DownloadAndDecryptFile(ctx context.Context, fileID primitive.ObjectID, userPassword string, urlDuration time.Duration) (*DownloadResult, error) {
	s.logger.Info("Starting E2EE file download and decryption", zap.String("fileID", fileID.Hex()))

	//
	// Step 1: Validate inputs
	//
	if fileID.IsZero() {
		return nil, errors.NewAppError("file ID is required", nil)
	}
	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE decryption", nil)
	}

	//
	// Step 2: Get file metadata (contains encrypted file key and metadata)
	//
	file, err := s.getFileUseCase.Execute(ctx, fileID)
	if err != nil {
		return nil, errors.NewAppError("failed to get file metadata", err)
	}
	if file == nil {
		return nil, errors.NewAppError("file not found", nil)
	}

	//
	// Step 3: Get user and collection for E2EE key chain
	//
	user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, errors.NewAppError("user not found", nil)
	}

	collection, err := s.getCollectionUseCase.Execute(ctx, file.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collection == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	//
	// Step 4: Decrypt the E2EE key chain
	//
	collectionKey, err := s.decryptCollectionKeyChain(user, collection, userPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// Decrypt the file key using collection key
	fileKey, err := common_crypto.DecryptWithSecretBox(
		file.EncryptedFileKey.Ciphertext,
		file.EncryptedFileKey.Nonce,
		collectionKey,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file key", err)
	}
	defer crypto.ClearBytes(fileKey)

	//
	// Step 5: Decrypt file metadata
	//
	decryptedMetadata, err := s.decryptFileMetadata(file.EncryptedMetadata, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file metadata", err)
	}

	//
	// Step 6: Get presigned download URLs
	//
	urlResponse, err := s.getPresignedDownloadURLUseCase.Execute(ctx, fileID, urlDuration)
	if err != nil {
		return nil, errors.NewAppError("failed to get presigned download URLs", err)
	}

	if !urlResponse.Success {
		return nil, errors.NewAppError("server failed to generate presigned URLs: "+urlResponse.Message, nil)
	}

	//
	// Step 7: Download encrypted file content
	//
	downloadRequest := &filedto.DownloadRequest{
		PresignedURL:          urlResponse.PresignedDownloadURL,
		PresignedThumbnailURL: urlResponse.PresignedThumbnailURL,
	}

	downloadResponse, err := s.downloadFileUseCase.Execute(ctx, downloadRequest)
	if err != nil {
		return nil, errors.NewAppError("failed to download file content", err)
	}

	//
	// Step 8: Decrypt the file content
	//
	decryptedData, err := s.decryptFileContent(downloadResponse.FileData, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file content", err)
	}

	//
	// Step 9: Decrypt thumbnail if present
	//
	var thumbnailData []byte
	if downloadResponse.ThumbnailData != nil && len(downloadResponse.ThumbnailData) > 0 {
		thumbnailData, err = s.decryptFileContent(downloadResponse.ThumbnailData, fileKey)
		if err != nil {
			s.logger.Warn("Failed to decrypt thumbnail, continuing without it", zap.Error(err))
			thumbnailData = nil
		}
	}

	result := &DownloadResult{
		FileID:            fileID,
		DecryptedData:     decryptedData,
		DecryptedMetadata: decryptedMetadata,
		ThumbnailData:     thumbnailData,
		OriginalSize:      int64(len(decryptedData)),
		ThumbnailSize:     int64(len(thumbnailData)),
	}

	s.logger.Info("Successfully completed E2EE file download and decryption",
		zap.String("fileID", fileID.Hex()),
		zap.String("fileName", decryptedMetadata.Name),
		zap.Int64("originalSize", result.OriginalSize))

	return result, nil
}

// decryptCollectionKeyChain decrypts the complete E2EE chain to get the collection key
func (s *downloadService) decryptCollectionKeyChain(user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	// STEP 1: Derive keyEncryptionKey from password
	keyEncryptionKey, err := pkg_crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer pkg_crypto.ClearBytes(keyEncryptionKey)

	// STEP 2: Decrypt masterKey with keyEncryptionKey
	masterKey, err := pkg_crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key - incorrect password?: %w", err)
	}
	defer pkg_crypto.ClearBytes(masterKey)

	// STEP 3: Decrypt collectionKey with masterKey
	if collection.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	collectionKey, err := pkg_crypto.DecryptWithSecretBox(
		collection.EncryptedCollectionKey.Ciphertext,
		collection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key: %w", err)
	}

	return collectionKey, nil
}

// decryptFileMetadata decrypts the encrypted file metadata
func (s *downloadService) decryptFileMetadata(encryptedMetadata string, fileKey []byte) (*DecryptedFileMetadata, error) {
	// Decode from base64
	combined, err := pkg_crypto.DecodeFromBase64(encryptedMetadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted metadata: %w", err)
	}

	// Split nonce and ciphertext
	nonce, ciphertext, err := pkg_crypto.SplitNonceAndCiphertext(combined, pkg_crypto.SecretBoxNonceSize)
	if err != nil {
		return nil, fmt.Errorf("failed to split nonce and ciphertext: %w", err)
	}

	// Decrypt metadata
	decryptedBytes, err := common_crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt metadata: %w", err)
	}

	// Parse JSON metadata
	var metadata DecryptedFileMetadata
	if err := json.Unmarshal(decryptedBytes, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted metadata: %w", err)
	}

	return &metadata, nil
}

// decryptFileContent decrypts the encrypted file content
func (s *downloadService) decryptFileContent(encryptedData, fileKey []byte) ([]byte, error) {
	s.logger.Debug("Decrypting file content", zap.Int("encryptedSize", len(encryptedData)))

	// The encrypted data should be in the format: nonce (24 bytes) + ciphertext
	if len(encryptedData) < common_crypto.NonceSize {
		return nil, fmt.Errorf("encrypted data too short: expected at least %d bytes, got %d",
			common_crypto.NonceSize, len(encryptedData))
	}

	// Extract nonce and ciphertext from combined data
	nonce := make([]byte, common_crypto.NonceSize)
	copy(nonce, encryptedData[:common_crypto.NonceSize])

	ciphertext := make([]byte, len(encryptedData)-common_crypto.NonceSize)
	copy(ciphertext, encryptedData[common_crypto.NonceSize:])

	// Decrypt the content
	decryptedData, err := common_crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file content: %w", err)
	}

	s.logger.Debug("Successfully decrypted file content",
		zap.Int("decryptedSize", len(decryptedData)))

	return decryptedData, nil
}
