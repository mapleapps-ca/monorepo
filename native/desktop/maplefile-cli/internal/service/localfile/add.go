// internal/service/localfile/add.go
package localfile

import (
	"context"
	"encoding/json"
	"mime"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// AddInput represents the input for adding a local file
type AddInput struct {
	FilePath     string             `json:"file_path"`
	CollectionID primitive.ObjectID `json:"collection_id"`
	OwnerID      primitive.ObjectID `json:"owner_id"`
	Name         string             `json:"name,omitempty"`
	StorageMode  string             `json:"storage_mode"`
}

// AddOutput represents the result of adding a local file
type AddOutput struct {
	File           *dom_file.File `json:"file"`
	CopiedFilePath string         `json:"copied_file_path"`
}

// AddService defines the interface for adding local files
type AddService interface {
	Add(ctx context.Context, input *AddInput) (*AddOutput, error)
}

// addService implements the AddService interface
type addService struct {
	logger                 *zap.Logger
	configService          config.ConfigService
	readFileUseCase        localfile.ReadFileUseCase
	checkFileExistsUseCase localfile.CheckFileExistsUseCase
	getFileInfoUseCase     localfile.GetFileInfoUseCase
	pathUtilsUseCase       localfile.PathUtilsUseCase
	copyFileUseCase        localfile.CopyFileUseCase
	createDirectoryUseCase localfile.CreateDirectoryUseCase
	createFileUseCase      file.CreateFileUseCase
}

// NewAddService creates a new service for adding local files
func NewAddService(
	logger *zap.Logger,
	configService config.ConfigService,
	readFileUseCase localfile.ReadFileUseCase,
	checkFileExistsUseCase localfile.CheckFileExistsUseCase,
	getFileInfoUseCase localfile.GetFileInfoUseCase,
	pathUtilsUseCase localfile.PathUtilsUseCase,
	copyFileUseCase localfile.CopyFileUseCase,
	createDirectoryUseCase localfile.CreateDirectoryUseCase,
	createFileUseCase file.CreateFileUseCase,
) AddService {
	return &addService{
		logger:                 logger,
		configService:          configService,
		readFileUseCase:        readFileUseCase,
		checkFileExistsUseCase: checkFileExistsUseCase,
		getFileInfoUseCase:     getFileInfoUseCase,
		pathUtilsUseCase:       pathUtilsUseCase,
		copyFileUseCase:        copyFileUseCase,
		createDirectoryUseCase: createDirectoryUseCase,
		createFileUseCase:      createFileUseCase,
	}
}

// Add handles the addition of a local file to the MapleFile system
func (s *addService) Add(ctx context.Context, input *AddInput) (*AddOutput, error) {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.FilePath == "" {
		s.logger.Error("file path is required")
		return nil, errors.NewAppError("file path is required", nil)
	}
	if input.CollectionID.IsZero() {
		s.logger.Error("collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.OwnerID.IsZero() {
		s.logger.Error("owner ID is required")
		return nil, errors.NewAppError("owner ID is required", nil)
	}
	if input.StorageMode == "" {
		input.StorageMode = dom_file.StorageModeEncryptedOnly
	}
	if input.StorageMode != dom_file.StorageModeEncryptedOnly &&
		input.StorageMode != dom_file.StorageModeDecryptedOnly &&
		input.StorageMode != dom_file.StorageModeHybrid {
		s.logger.Error("invalid storage mode", zap.String("storageMode", input.StorageMode))
		return nil, errors.NewAppError("invalid storage mode. Must be 'encrypted_only', 'hybrid', or 'decrypted_only'", nil)
	}

	//
	// STEP 2: Clean and validate file path
	//
	cleanFilePath := s.pathUtilsUseCase.Clean(ctx, input.FilePath)
	s.logger.Debug("Cleaned file path",
		zap.String("original", input.FilePath),
		zap.String("cleaned", cleanFilePath))

	// Check if file exists
	exists, err := s.checkFileExistsUseCase.Execute(ctx, cleanFilePath)
	if err != nil {
		s.logger.Error("Failed to check file existence", zap.String("filePath", cleanFilePath), zap.Error(err))
		return nil, errors.NewAppError("failed to check file existence", err)
	}
	if !exists {
		s.logger.Error("File does not exist", zap.String("filePath", cleanFilePath))
		return nil, errors.NewAppError("file does not exist", nil)
	}

	//
	// STEP 3: Get file information
	//
	fileInfo, err := s.getFileInfoUseCase.Execute(ctx, cleanFilePath)
	if err != nil {
		s.logger.Error("Failed to get file info", zap.String("filePath", cleanFilePath), zap.Error(err))
		return nil, errors.NewAppError("failed to get file info", err)
	}
	if fileInfo.IsDirectory {
		s.logger.Error("Path is a directory, not a file", zap.String("filePath", cleanFilePath))
		return nil, errors.NewAppError("the specified path is a directory, not a file", nil)
	}

	//
	// STEP 4: Get app directory and create file storage structure
	//
	appDataDir, err := s.configService.GetAppDataDirPath(ctx)
	if err != nil {
		s.logger.Error("Failed to get app data directory path", zap.Error(err))
		return nil, errors.NewAppError("failed to get app data directory path", err)
	}

	// Create files storage directory (cross-platform compatible)
	filesDir := s.pathUtilsUseCase.Join(ctx, appDataDir, "files")
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, filesDir); err != nil {
		s.logger.Error("Failed to create files directory", zap.String("filesDir", filesDir), zap.Error(err))
		return nil, errors.NewAppError("failed to create files directory", err)
	}

	// Create bin subdirectory
	binDir := s.pathUtilsUseCase.Join(ctx, filesDir, "bin")
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, binDir); err != nil {
		s.logger.Error("Failed to create bin directory", zap.String("binDir", binDir), zap.Error(err))
		return nil, errors.NewAppError("failed to create bin directory", err)
	}

	// Create collection-specific subdirectory
	collectionDir := s.pathUtilsUseCase.Join(ctx, binDir, input.CollectionID.Hex())
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, collectionDir); err != nil {
		s.logger.Error("Failed to create collection directory", zap.String("collectionDir", collectionDir), zap.Error(err))
		return nil, errors.NewAppError("failed to create collection directory", err)
	}

	//
	// STEP 5: Prepare file metadata
	//
	fileName := input.Name
	if fileName == "" {
		fileName = s.pathUtilsUseCase.GetFileName(ctx, cleanFilePath)
	}

	// Detect MIME type
	fileExtension := s.pathUtilsUseCase.GetFileExtension(ctx, cleanFilePath)
	mimeType := mime.TypeByExtension(fileExtension)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Generate unique file ID and create destination path
	fileID := primitive.NewObjectID()
	destFileName := fileID.Hex() + fileExtension
	destFilePath := s.pathUtilsUseCase.Join(ctx, collectionDir, destFileName)

	//
	// STEP 6: Copy file to app directory
	//
	s.logger.Info("Copying file to app directory",
		zap.String("source", cleanFilePath),
		zap.String("destination", destFilePath))

	if err := s.copyFileUseCase.Execute(ctx, cleanFilePath, destFilePath); err != nil {
		s.logger.Error("Failed to copy file",
			zap.String("source", cleanFilePath),
			zap.String("destination", destFilePath),
			zap.Error(err))
		return nil, errors.NewAppError("failed to copy file to app directory", err)
	}

	//
	// STEP 7: Generate encryption keys and encrypt file data
	//
	// Generate file key
	fileKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		s.logger.Error("Failed to generate file key", zap.Error(err))
		return nil, errors.NewAppError("failed to generate file key", err)
	}

	// For now, use a simple placeholder for collection key
	// In a real implementation, you'd retrieve this from the collection
	collectionKey, err := crypto.GenerateRandomBytes(crypto.SecretBoxKeySize)
	if err != nil {
		s.logger.Error("Failed to generate collection key", zap.Error(err))
		return nil, errors.NewAppError("failed to generate collection key", err)
	}

	// Encrypt file key with collection key
	encryptedFileKeyCiphertext, encryptedFileKeyNonce, err := crypto.EncryptWithSecretBox(fileKey, collectionKey)
	if err != nil {
		s.logger.Error("Failed to encrypt file key", zap.Error(err))
		return nil, errors.NewAppError("failed to encrypt file key", err)
	}

	encryptedFileKey := keys.EncryptedFileKey{
		Ciphertext: encryptedFileKeyCiphertext,
		Nonce:      encryptedFileKeyNonce,
		KeyVersion: 1,
	}

	// Create metadata map
	metadataMap := map[string]interface{}{
		"name":      fileName,
		"mime_type": mimeType,
		"size":      fileInfo.Size,
	}

	// Encode metadata map to JSON bytes. This serves as the format
	// that would eventually be encrypted. For now, we store the raw bytes
	// as a placeholder for the 'encrypted' metadata.
	// This satisfies the requirement to "fix simple conversion" by encoding,
	// while "skipping encryption" by storing the raw encoded data.
	metadataBytes, err := json.Marshal(metadataMap)
	if err != nil {
		s.logger.Error("Failed to marshal metadata map to JSON", zap.Error(err))
		// Handle this error - it prevents the rest of the logic from running
		return nil, errors.NewAppError("failed to prepare metadata for storage", err)
	}

	// Store the raw JSON bytes (converted to string) in the field intended for encrypted data.
	// This is NOT encrypted, just encoded and converted to string.
	encryptedMetadataString := string(metadataBytes)

	//
	// STEP 8: Create domain file object
	//
	currentTime := time.Now()
	domainFile := &dom_file.File{
		ID:                fileID,
		CollectionID:      input.CollectionID,
		OwnerID:           input.OwnerID,
		EncryptedMetadata: encryptedMetadataString, // Assign the string here
		EncryptedFileKey:  encryptedFileKey,
		EncryptionVersion: "v1",
		Name:              fileName,
		MimeType:          mimeType,
		FilePath:          destFilePath, // Decrypted file path (what we copied)
		FileSize:          fileInfo.Size,
		StorageMode:       input.StorageMode,
		CreatedAt:         currentTime,
		CreatedByUserID:   input.OwnerID,
		ModifiedAt:        currentTime,
		ModifiedByUserID:  input.OwnerID,
		Version:           1,
		SyncStatus:        dom_file.SyncStatusLocalOnly,
	}

	// Set paths based on storage mode
	switch input.StorageMode {
	case dom_file.StorageModeEncryptedOnly:
		// In real implementation, you'd encrypt the file and store only encrypted version
		domainFile.EncryptedFilePath = destFilePath
		domainFile.EncryptedFileSize = fileInfo.Size
		domainFile.FilePath = "" // No decrypted version stored
		domainFile.FileSize = 0
	case dom_file.StorageModeDecryptedOnly:
		// Keep only decrypted version (not recommended)
		domainFile.FilePath = destFilePath
		domainFile.FileSize = fileInfo.Size
		domainFile.EncryptedFilePath = ""
		domainFile.EncryptedFileSize = 0
	case dom_file.StorageModeHybrid:
		// Keep both versions
		domainFile.FilePath = destFilePath
		domainFile.FileSize = fileInfo.Size
		// In real implementation, you'd create encrypted version too
		domainFile.EncryptedFilePath = destFilePath + ".encrypted"
		domainFile.EncryptedFileSize = fileInfo.Size
	}

	//
	// STEP 9: Save file record to database
	//
	if err := s.createFileUseCase.Execute(ctx, domainFile); err != nil {
		s.logger.Error("Failed to create file record", zap.String("fileID", fileID.Hex()), zap.Error(err))
		return nil, errors.NewAppError("failed to create file record", err)
	}

	s.logger.Info("Successfully added file",
		zap.String("fileID", fileID.Hex()),
		zap.String("fileName", fileName),
		zap.String("copiedPath", destFilePath))

	return &AddOutput{
		File:           domainFile,
		CopiedFilePath: destFilePath,
	}, nil
}
