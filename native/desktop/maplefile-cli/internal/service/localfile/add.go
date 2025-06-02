// internal/service/localfile/add.go
package localfile

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_keys "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
	sprimitive "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/storage/mongodb"
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
	Add(ctx context.Context, input *AddInput, userPassword string) (*AddOutput, error)
}

// addService implements the AddService interface
type addService struct {
	logger                     *zap.Logger
	configService              config.ConfigService
	primitiveIDObjectGenerator sprimitive.SecurePrimitiveObjectIDGenerator
	readFileUseCase            localfile.ReadFileUseCase
	checkFileExistsUseCase     localfile.CheckFileExistsUseCase
	getFileInfoUseCase         localfile.GetFileInfoUseCase
	computeFileHashUseCase     localfile.ComputeFileHashUseCase
	pathUtilsUseCase           localfile.PathUtilsUseCase
	copyFileUseCase            localfile.CopyFileUseCase
	createDirectoryUseCase     localfile.CreateDirectoryUseCase
	transactionManager         dom_tx.Manager
	createFileUseCase          file.CreateFileUseCase
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase       uc_collection.GetCollectionUseCase
}

// NewAddService creates a new service for adding local files
func NewAddService(
	logger *zap.Logger,
	configService config.ConfigService,
	primitiveIDObjectGenerator sprimitive.SecurePrimitiveObjectIDGenerator,
	readFileUseCase localfile.ReadFileUseCase,
	checkFileExistsUseCase localfile.CheckFileExistsUseCase,
	getFileInfoUseCase localfile.GetFileInfoUseCase,
	computeFileHashUseCase localfile.ComputeFileHashUseCase,
	pathUtilsUseCase localfile.PathUtilsUseCase,
	copyFileUseCase localfile.CopyFileUseCase,
	createDirectoryUseCase localfile.CreateDirectoryUseCase,
	transactionManager dom_tx.Manager,
	createFileUseCase file.CreateFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) AddService {
	logger = logger.Named("AddService")
	return &addService{
		logger:                     logger,
		configService:              configService,
		primitiveIDObjectGenerator: primitiveIDObjectGenerator,
		readFileUseCase:            readFileUseCase,
		checkFileExistsUseCase:     checkFileExistsUseCase,
		getFileInfoUseCase:         getFileInfoUseCase,
		computeFileHashUseCase:     computeFileHashUseCase,
		pathUtilsUseCase:           pathUtilsUseCase,
		copyFileUseCase:            copyFileUseCase,
		createDirectoryUseCase:     createDirectoryUseCase,
		transactionManager:         transactionManager,
		createFileUseCase:          createFileUseCase,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		getCollectionUseCase:       getCollectionUseCase,
	}
}

// Add handles the addition of a local file to the MapleFile system
func (s *addService) Add(ctx context.Context, input *AddInput, userPassword string) (*AddOutput, error) {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.FilePath == "" {
		s.logger.Error("❌ File path is required")
		return nil, errors.NewAppError("file path is required", nil)
	}
	if input.CollectionID.IsZero() {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.OwnerID.IsZero() {
		s.logger.Error("❌ Owner ID is required")
		return nil, errors.NewAppError("owner ID is required", nil)
	}
	if input.StorageMode == "" {
		input.StorageMode = dom_file.StorageModeEncryptedOnly
	}
	if input.StorageMode != dom_file.StorageModeEncryptedOnly &&
		input.StorageMode != dom_file.StorageModeDecryptedOnly &&
		input.StorageMode != dom_file.StorageModeHybrid {
		s.logger.Error("❌ Invalid storage mode", zap.String("storageMode", input.StorageMode))
		return nil, errors.NewAppError("invalid storage mode. Must be 'encrypted_only', 'hybrid', or 'decrypted_only'", nil)
	}
	if userPassword == "" {
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	//
	// STEP 2: Clean and validate file path
	//
	cleanFilePath := s.pathUtilsUseCase.Clean(ctx, input.FilePath)
	s.logger.Debug("🔍 Cleaned file path",
		zap.String("original", input.FilePath),
		zap.String("cleaned", cleanFilePath))

	// Check if file exists
	exists, err := s.checkFileExistsUseCase.Execute(ctx, cleanFilePath)
	if err != nil {
		s.logger.Error("❌ Failed to check file existence", zap.String("filePath", cleanFilePath), zap.Error(err))
		return nil, errors.NewAppError("failed to check file existence", err)
	}
	if !exists {
		s.logger.Error("❌ File does not exist", zap.String("filePath", cleanFilePath))
		return nil, errors.NewAppError("file does not exist", nil)
	}

	//
	// STEP 3: Get related data.
	//
	// Get logged-in user
	user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, errors.NewAppError("logged in user does not exist", nil)
	}

	// Get collection
	collection, err := s.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collection == nil {
		return nil, errors.NewAppError("collection does not exist", nil)
	}

	// Get file information
	fileInfo, err := s.getFileInfoUseCase.Execute(ctx, cleanFilePath)
	if err != nil {
		s.logger.Error("❌ Failed to get file info", zap.String("filePath", cleanFilePath), zap.Error(err))
		return nil, errors.NewAppError("failed to get file info", err)
	}
	if fileInfo == nil {
		return nil, errors.NewAppError("fileInfo does not exist", nil)
	}
	if fileInfo.IsDirectory {
		s.logger.Error("❌ Path is a directory, not a file", zap.String("filePath", cleanFilePath))
		return nil, errors.NewAppError("the specified path is a directory, not a file", nil)
	}

	//
	// STEP 4: Get app directory and create file storage structure
	//
	appDataDir, err := s.configService.GetAppDataDirPath(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to get app data directory path", zap.Error(err))
		return nil, errors.NewAppError("failed to get app data directory path", err)
	}

	// Create files storage directory (cross-platform compatible)
	filesDir := s.pathUtilsUseCase.Join(ctx, appDataDir, "files")
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, filesDir); err != nil {
		s.logger.Error("❌ Failed to create files directory", zap.String("filesDir", filesDir), zap.Error(err))
		return nil, errors.NewAppError("failed to create files directory", err)
	}

	// Create bin subdirectory
	binDir := s.pathUtilsUseCase.Join(ctx, filesDir, "bin")
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, binDir); err != nil {
		s.logger.Error("❌ Failed to create bin directory", zap.String("binDir", binDir), zap.Error(err))
		return nil, errors.NewAppError("failed to create bin directory", err)
	}

	// Create collection-specific subdirectory
	collectionDir := s.pathUtilsUseCase.Join(ctx, binDir, input.CollectionID.Hex())
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, collectionDir); err != nil {
		s.logger.Error("❌ Failed to create collection directory", zap.String("collectionDir", collectionDir), zap.Error(err))
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
	fileID := s.primitiveIDObjectGenerator.GenerateValidObjectID()
	destFileName := fileID.Hex() + fileExtension
	destFilePath := s.pathUtilsUseCase.Join(ctx, collectionDir, destFileName)

	//
	// STEP 6: Copy file to app directory
	//
	s.logger.Info("📄 Copying file to app directory",
		zap.String("source", cleanFilePath),
		zap.String("destination", destFilePath))

	if err := s.copyFileUseCase.Execute(ctx, cleanFilePath, destFilePath); err != nil {
		s.logger.Error("❌ Failed to copy file",
			zap.String("source", cleanFilePath),
			zap.String("destination", destFilePath),
			zap.Error(err))
		return nil, errors.NewAppError("failed to copy file to app directory", err)
	}

	//
	// STEP 7:
	// E2EE DECRYPTION CHAIN: password → keyEncryptionKey → masterKey → collectionKey.
	// Generate encryption keys and encrypt file data
	//

	collectionKey, err := s.decryptCollectionKeyChain(user, collection, userPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	// STEP 1: Generate random fileKey (E2EE spec)
	fileKey, err := crypto.GenerateRandomBytes(crypto.FileKeySize)
	if err != nil {
		return nil, errors.NewAppError("failed to generate file key", err)
	}
	defer crypto.ClearBytes(fileKey)

	// STEP 2: Encrypt fileKey with collectionKey (E2EE spec)
	encryptedFileKeyData, err := crypto.EncryptWithSecretBox(fileKey, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file key", err)
	}

	// STEP 3: Encrypt file content with fileKey (E2EE spec)
	encryptedFileData, err := s.encryptFileContent(destFilePath, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file content", err)
	}

	// STEP 4: Encrypt file metadata with fileKey (E2EE spec)
	metadata := &dom_file.FileMetadata{
		Name:                   fileName,
		MimeType:               mimeType,
		Size:                   fileInfo.Size,
		Created:                fileInfo.ModifiedAt.Unix(),
		DecryptedFilePath:      destFilePath, // Decrypted file path (what we copied)
		DecryptedFileSize:      fileInfo.Size,
		EncryptedFilePath:      encryptedFileData.Path,
		EncryptedFileSize:      encryptedFileData.Size,
		EncryptedThumbnailPath: "", // Developer Note: Future feature
		EncryptedThumbnailSize: 0,  // Developer Note: Future feature
		DecryptedThumbnailPath: "", // Developer Note: Future feature
		DecryptedThumbnailSize: 0,  // Developer Note: Future feature
	}
	encryptedMetadataString, err := s.encryptFileMetadata(metadata, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file metadata", err)
	}

	// STEP 5: Encrypt file hash with fileKey (E2EE spec)
	encryptedHashString, err := s.encryptComputeFileHash(ctx, destFilePath, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file hash", err)
	}

	// Step 6
	currentTime := time.Now()
	historicalKey := dom_keys.EncryptedHistoricalKey{
		Ciphertext:    encryptedFileKeyData.Ciphertext,
		Nonce:         encryptedFileKeyData.Nonce,
		KeyVersion:    1,
		RotatedAt:     currentTime,
		RotatedReason: "Initial collection creation",
		Algorithm:     crypto.ChaCha20Poly1305Algorithm,
	}

	//
	// STEP 8: Create domain file object
	//
	domainFile := &dom_file.File{
		ID:                fileID,
		CollectionID:      input.CollectionID,
		OwnerID:           input.OwnerID,
		EncryptedMetadata: encryptedMetadataString,
		EncryptedFileKey: dom_keys.EncryptedFileKey{
			Ciphertext:   encryptedFileKeyData.Ciphertext,
			Nonce:        encryptedFileKeyData.Nonce,
			KeyVersion:   1,
			RotatedAt:    &currentTime,
			PreviousKeys: []dom_keys.EncryptedHistoricalKey{historicalKey},
		},
		EncryptionVersion: "1.0",
		EncryptedHash:     encryptedHashString,
		EncryptedFilePath: encryptedFileData.Path,
		EncryptedFileSize: encryptedFileData.Size,
		Name:              fileName, // Keep plaintext for local use
		MimeType:          mimeType,
		Metadata:          metadata,     // Decrypted metadata.
		FilePath:          destFilePath, // Decrypted file path (what we copied)
		FileSize:          fileInfo.Size,
		StorageMode:       input.StorageMode,
		CreatedAt:         currentTime,
		CreatedByUserID:   input.OwnerID,
		ModifiedAt:        currentTime,
		ModifiedByUserID:  input.OwnerID,
		Version:           1,                            // Always set `version=1` at creation of a collection
		SyncStatus:        dom_file.SyncStatusLocalOnly, // SET DEFAULT STATE
	}

	//
	// STEP 9: Save file record to database
	//
	if err := s.createFileUseCase.Execute(ctx, domainFile); err != nil {
		s.logger.Error("❌ Failed to create file record", zap.String("fileID", fileID.Hex()), zap.Error(err))
		return nil, errors.NewAppError("failed to create file record", err)
	}

	s.logger.Info("✅ Successfully added E2EE file",
		zap.String("fileID", fileID.Hex()),
		zap.String("fileName", fileName),
		zap.String("copiedPath", destFilePath))

	return &AddOutput{
		File:           domainFile,
		CopiedFilePath: destFilePath,
	}, nil
}

// Complete E2EE decryption chain
func (s *addService) decryptCollectionKeyChain(user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	// STEP 1: Derive keyEncryptionKey from password
	keyEncryptionKey, err := crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ClearBytes(keyEncryptionKey)

	// STEP 2: Decrypt masterKey with keyEncryptionKey
	masterKey, err := crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key - incorrect password?: %w", err)
	}
	defer crypto.ClearBytes(masterKey)

	// STEP 3: Decrypt collectionKey with masterKey
	if collection.EncryptedCollectionKey == nil {
		return nil, errors.NewAppError("collection has no encrypted key", nil)
	}

	collectionKey, err := crypto.DecryptWithSecretBox(
		collection.EncryptedCollectionKey.Ciphertext,
		collection.EncryptedCollectionKey.Nonce,
		masterKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt collection key: %w", err)
	}

	return collectionKey, nil
}

// Encrypt file metadata with fileKey
func (s *addService) encryptFileMetadata(metadata *dom_file.FileMetadata, fileKey []byte) (string, error) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	encryptedData, err := crypto.EncryptWithSecretBox(metadataBytes, fileKey)
	if err != nil {
		return "", err
	}

	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)
	return crypto.EncodeToBase64(combined), nil
}

type EncryptedFileData struct {
	Path string
	Size int64
	Hash []byte
}

// Encrypt file content with fileKey (E2EE: files encrypted with fileKey)
func (s *addService) encryptFileContent(filePath string, fileKey []byte) (*EncryptedFileData, error) {
	// Read original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Encrypt content with fileKey
	encryptedData, err := crypto.EncryptWithSecretBox(content, fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt file content: %w", err)
	}

	// Combine nonce and ciphertext
	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)

	// Write encrypted file
	encryptedPath := filePath + ".encrypted"
	if err := os.WriteFile(encryptedPath, combined, 0600); err != nil {
		return nil, fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return &EncryptedFileData{
		Path: encryptedPath,
		Size: int64(len(combined)),
	}, nil
}

// Helper: Compute file hash and return the hash in an encrypted format
func (s *addService) encryptComputeFileHash(ctx context.Context, filePath string, fileKey []byte) (string, error) {
	// Compute file hash - use buffered algorithm in case of large files.
	fileHashBytes, err := s.computeFileHashUseCase.ExecuteForBytes(ctx, filePath)
	if err != nil {
		s.logger.Error("❌ Failed to compute file hash",
			zap.String("file_path", filePath),
			zap.Error(err))
		return "", errors.NewAppError("failed to compute file hash", err)
	}

	encryptedData, err := crypto.EncryptWithSecretBox(fileHashBytes, fileKey)
	if err != nil {
		return "", err
	}

	combined := crypto.CombineNonceAndCiphertext(encryptedData.Nonce, encryptedData.Ciphertext)
	return crypto.EncodeToBase64(combined), nil
}
