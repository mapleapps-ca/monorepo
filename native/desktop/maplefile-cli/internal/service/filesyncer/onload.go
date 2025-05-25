// internal/service/filesyncer/onload.go
package filesyncer

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/crypto"
	common_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/crypto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	svc_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/crypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// OnloadInput represents the input for onloading a cloud-only file
type OnloadInput struct {
	FileID       string `json:"file_id"`
	UserPassword string `json:"user_password"`
}

// OnloadOutput represents the result of onloading a cloud-only file
type OnloadOutput struct {
	FileID         primitive.ObjectID  `json:"file_id"`
	PreviousStatus dom_file.SyncStatus `json:"previous_status"`
	NewStatus      dom_file.SyncStatus `json:"new_status"`
	DecryptedPath  string              `json:"decrypted_path"`
	DownloadedSize int64               `json:"downloaded_size"`
	Message        string              `json:"message"`
}

// OnloadService defines the interface for onloading cloud-only files
type OnloadService interface {
	Onload(ctx context.Context, input *OnloadInput) (*OnloadOutput, error)
}

// onloadService implements the OnloadService interface
type onloadService struct {
	logger                     *zap.Logger
	configService              config.ConfigService
	cryptoService              svc_crypto.CryptoService
	getFileUseCase             uc_file.GetFileUseCase
	updateFileUseCase          uc_file.UpdateFileUseCase
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase       uc_collection.GetCollectionUseCase
	fileDTORepo                filedto.FileDTORepository
	pathUtilsUseCase           localfile.PathUtilsUseCase
	createDirectoryUseCase     localfile.CreateDirectoryUseCase
	httpClient                 *http.Client
}

// NewOnloadService creates a new service for onloading cloud-only files
func NewOnloadService(
	logger *zap.Logger,
	configService config.ConfigService,
	cryptoService svc_crypto.CryptoService,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	fileDTORepo filedto.FileDTORepository,
	pathUtilsUseCase localfile.PathUtilsUseCase,
	createDirectoryUseCase localfile.CreateDirectoryUseCase,
) OnloadService {
	return &onloadService{
		logger:                     logger,
		configService:              configService,
		cryptoService:              cryptoService,
		getFileUseCase:             getFileUseCase,
		updateFileUseCase:          updateFileUseCase,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		getCollectionUseCase:       getCollectionUseCase,
		fileDTORepo:                fileDTORepo,
		pathUtilsUseCase:           pathUtilsUseCase,
		createDirectoryUseCase:     createDirectoryUseCase,
		httpClient:                 &http.Client{Timeout: 30 * time.Second},
	}
}

// Onload handles the onloading of a cloud-only file to local storage
func (s *onloadService) Onload(ctx context.Context, input *OnloadInput) (*OnloadOutput, error) {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.FileID == "" {
		s.logger.Error("file ID is required")
		return nil, errors.NewAppError("file ID is required", nil)
	}
	if input.UserPassword == "" {
		s.logger.Error("user password is required for E2EE operations")
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	//
	// STEP 2: Convert file ID string to ObjectID
	//
	fileObjectID, err := primitive.ObjectIDFromHex(input.FileID)
	if err != nil {
		s.logger.Error("invalid file ID format",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	//
	// STEP 3: Get the file and validate it's cloud-only
	//
	s.logger.Debug("Getting file for onload operation",
		zap.String("fileID", input.FileID))

	file, err := s.getFileUseCase.Execute(ctx, fileObjectID)
	if err != nil {
		s.logger.Error("failed to get file",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to get file", err)
	}

	if file == nil {
		s.logger.Error("file not found", zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("file not found", nil)
	}

	previousStatus := file.SyncStatus

	// Only work with cloud-only files
	if file.SyncStatus != dom_file.SyncStatusCloudOnly {
		s.logger.Error("file is not cloud-only",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError(
			fmt.Sprintf("file is not cloud-only (current status: %v)", file.SyncStatus),
			nil)
	}

	//
	// STEP 4: Get user and collection for decryption
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
	// STEP 5: Download encrypted file from cloud
	//
	encryptedData, err := s.downloadEncryptedFile(ctx, file)
	if err != nil {
		return nil, errors.NewAppError("failed to download encrypted file", err)
	}

	//
	// STEP 6: Decrypt the collection key chain
	//
	collectionKey, err := s.decryptCollectionKeyChain(user, collection, input.UserPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	//
	// STEP 7: Decrypt the file key
	//
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
	// STEP 8: Decrypt the file content
	//
	decryptedData, err := s.decryptFileContent(encryptedData, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file content", err)
	}

	//
	// STEP 9: Save decrypted file locally
	//
	decryptedPath, err := s.saveDecryptedFile(ctx, file, decryptedData)
	if err != nil {
		return nil, errors.NewAppError("failed to save decrypted file", err)
	}

	//
	// STEP 10: Update file record with new path and sync status
	//
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
	}

	newStatus := dom_file.SyncStatusSynced
	updateInput.SyncStatus = &newStatus
	updateInput.FilePath = &decryptedPath

	_, err = s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("failed to update file sync status during onload",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file sync status during onload", err)
	}

	s.logger.Info("Successfully onloaded file",
		zap.String("fileID", input.FileID),
		zap.String("decryptedPath", decryptedPath),
		zap.Any("previousStatus", previousStatus),
		zap.Any("newStatus", newStatus))

	return &OnloadOutput{
		FileID:         fileObjectID,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		DecryptedPath:  decryptedPath,
		DownloadedSize: int64(len(encryptedData)),
		Message:        "File successfully onloaded and decrypted",
	}, nil
}

// downloadEncryptedFile downloads the encrypted file content from cloud storage
func (s *onloadService) downloadEncryptedFile(ctx context.Context, file *dom_file.File) ([]byte, error) {
	s.logger.Debug("Downloading encrypted file from cloud",
		zap.String("fileID", file.ID.Hex()))

	// Get download URL (this would typically involve getting a presigned URL)
	// For now, we'll use a direct download approach
	// In a real implementation, you might need to get a presigned download URL first

	// Download the encrypted file content
	// This is a simplified approach - in practice you might need to:
	// 1. Get a presigned download URL from the cloud service
	// 2. Download using that URL
	// For now, we'll assume we can download directly using the file DTO repository

	fileDTO, err := s.fileDTORepo.DownloadByIDFromCloud(ctx, file.ID)
	if err != nil {
		return nil, err
	}

	// In a real implementation, you'd download the actual file content
	// This is a placeholder that demonstrates the concept
	// You would typically use the fileDTO to get a download URL and then download the content

	s.logger.Debug("Successfully downloaded encrypted file metadata",
		zap.String("fileID", file.ID.Hex()),
		zap.Int64("encryptedSize", fileDTO.EncryptedFileSizeInBytes))

	// For now, we'll read from the local encrypted file path if it exists
	// In a real cloud scenario, you'd download from the cloud storage URL
	if file.EncryptedFilePath != "" {
		data, err := os.ReadFile(file.EncryptedFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read encrypted file: %w", err)
		}
		return data, nil
	}

	return nil, errors.NewAppError("no encrypted file path available for download", nil)
}

// decryptFileContent decrypts the encrypted file content using the file key
func (s *onloadService) decryptFileContent(encryptedData, fileKey []byte) ([]byte, error) {
	s.logger.Debug("Decrypting file content")

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

// saveDecryptedFile saves the decrypted file content to local storage
func (s *onloadService) saveDecryptedFile(ctx context.Context, file *dom_file.File, decryptedData []byte) (string, error) {
	s.logger.Debug("Saving decrypted file locally", zap.String("fileID", file.ID.Hex()))

	// Get app data directory
	appDataDir, err := s.configService.GetAppDataDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get app data directory: %w", err)
	}

	// Create files storage directory structure
	filesDir := s.pathUtilsUseCase.Join(ctx, appDataDir, "files")
	binDir := s.pathUtilsUseCase.Join(ctx, filesDir, "bin")
	collectionDir := s.pathUtilsUseCase.Join(ctx, binDir, file.CollectionID.Hex())

	// Create directories if they don't exist
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, collectionDir); err != nil {
		return "", fmt.Errorf("failed to create collection directory: %w", err)
	}

	// Generate file path with original extension
	fileExtension := filepath.Ext(file.Name)
	if fileExtension == "" {
		// Try to determine extension from MIME type if available
		fileExtension = s.getExtensionFromMimeType(file.MimeType)
	}

	destFileName := file.ID.Hex() + fileExtension
	destFilePath := s.pathUtilsUseCase.Join(ctx, collectionDir, destFileName)

	// Write the decrypted file
	err = os.WriteFile(destFilePath, decryptedData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}

	s.logger.Debug("Successfully saved decrypted file",
		zap.String("fileID", file.ID.Hex()),
		zap.String("filePath", destFilePath),
		zap.Int("size", len(decryptedData)))

	return destFilePath, nil
}

// decryptCollectionKeyChain decrypts the complete E2EE chain to get the collection key
func (s *onloadService) decryptCollectionKeyChain(user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	// STEP 1: Derive keyEncryptionKey from password
	keyEncryptionKey, err := common_crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer common_crypto.ClearBytes(keyEncryptionKey)

	// STEP 2: Decrypt masterKey with keyEncryptionKey
	masterKey, err := common_crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key - incorrect password?: %w", err)
	}
	defer common_crypto.ClearBytes(masterKey)

	// STEP 3: Decrypt collectionKey with masterKey
	if collection.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	collectionKey, err := common_crypto.DecryptWithSecretBox(
		collection.EncryptedCollectionKey.Ciphertext,
		collection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key: %w", err)
	}

	return collectionKey, nil
}

// getExtensionFromMimeType returns a file extension based on MIME type
func (s *onloadService) getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "text/plain":
		return ".txt"
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "application/json":
		return ".json"
	case "text/html":
		return ".html"
	case "application/zip":
		return ".zip"
	default:
		return ".dat" // Generic data file extension
	}
}
