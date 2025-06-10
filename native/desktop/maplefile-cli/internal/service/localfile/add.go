// internal/service/localfile/add.go
package localfile

import (
	"context"
	"mime"
	"os"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_tx "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/transaction"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	svc_filecrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filecrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// LocalFileAddInput represents the input for adding a local file
type LocalFileAddInput struct {
	FilePath     string     `json:"file_path"`
	CollectionID gocql.UUID `json:"collection_id"`
	OwnerID      gocql.UUID `json:"owner_id"`
	Name         string     `json:"name,omitempty"`
	StorageMode  string     `json:"storage_mode"`
}

// LocalFileAddOutput represents the result of adding a local file
type LocalFileAddOutput struct {
	File           *dom_file.File `json:"file"`
	CopiedFilePath string         `json:"copied_file_path"`
}

// LocalFileAddService defines the interface for adding local files
type LocalFileAddService interface {
	Add(ctx context.Context, input *LocalFileAddInput, userPassword string) (*LocalFileAddOutput, error)
}

// localFileAddService implements the LocalFileAddService interface
type localFileAddService struct {
	logger                      *zap.Logger
	configService               config.ConfigService
	readFileUseCase             localfile.ReadFileUseCase
	checkFileExistsUseCase      localfile.CheckFileExistsUseCase
	getFileInfoUseCase          localfile.GetFileInfoUseCase
	computeFileHashUseCase      localfile.ComputeFileHashUseCase
	pathUtilsUseCase            localfile.PathUtilsUseCase
	copyFileUseCase             localfile.CopyFileUseCase
	createDirectoryUseCase      localfile.CreateDirectoryUseCase
	transactionManager          dom_tx.Manager
	createFileUseCase           file.CreateFileUseCase
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase        uc_collection.GetCollectionUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
	fileEncryptionService       svc_filecrypto.FileEncryptionService
}

// NewLocalFileAddService creates a new service for adding local files
func NewLocalFileAddService(
	logger *zap.Logger,
	configService config.ConfigService,
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
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	fileEncryptionService svc_filecrypto.FileEncryptionService,
) LocalFileAddService {
	logger = logger.Named("LocalFileAddService")
	return &localFileAddService{
		logger:                      logger,
		configService:               configService,
		readFileUseCase:             readFileUseCase,
		checkFileExistsUseCase:      checkFileExistsUseCase,
		getFileInfoUseCase:          getFileInfoUseCase,
		computeFileHashUseCase:      computeFileHashUseCase,
		pathUtilsUseCase:            pathUtilsUseCase,
		copyFileUseCase:             copyFileUseCase,
		createDirectoryUseCase:      createDirectoryUseCase,
		transactionManager:          transactionManager,
		createFileUseCase:           createFileUseCase,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		getCollectionUseCase:        getCollectionUseCase,
		collectionDecryptionService: collectionDecryptionService,
		fileEncryptionService:       fileEncryptionService,
	}
}

// Add handles the addition of a local file to the MapleFile system
func (s *localFileAddService) Add(ctx context.Context, input *LocalFileAddInput, userPassword string) (*LocalFileAddOutput, error) {
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
	if input.CollectionID.String() == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.OwnerID.String() == "" {
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
	collectionDir := s.pathUtilsUseCase.Join(ctx, binDir, input.CollectionID.String())
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

	// Extract file extension
	fileExtension := s.pathUtilsUseCase.GetFileExtension(ctx, cleanFilePath)

	// Detect MIME type
	mimeType := mime.TypeByExtension(fileExtension)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Generate unique file ID and create destination path
	fileID := gocql.TimeUUID()
	destFileName := fileID.String() + fileExtension
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
	// STEP 7: E2EE ENCRYPTION CHAIN using crypto services
	//

	s.logger.Debug("🔐 Starting E2EE key chain decryption using crypto service")
	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, userPassword)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer crypto.ClearBytes(collectionKey)

	s.logger.Debug("🔐 Generating file key and encrypting with collection key using crypto service")
	encryptedFileKey, fileKey, err := s.fileEncryptionService.GenerateFileKeyAndEncryptWithCollectionKey(ctx, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to generate and encrypt file key", err)
	}
	defer crypto.ClearBytes(fileKey)

	s.logger.Debug("🔐 Encrypting file content using crypto service")
	fileContent, err := os.ReadFile(destFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to read file for encryption", err)
	}

	encryptedFileData, err := s.fileEncryptionService.EncryptFileContent(ctx, fileContent, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file content", err)
	}

	// Write encrypted file
	encryptedPath := destFilePath + ".encrypted"
	if err := os.WriteFile(encryptedPath, encryptedFileData, 0600); err != nil {
		return nil, errors.NewAppError("failed to write encrypted file", err)
	}

	s.logger.Debug("🔐 Encrypting file metadata using crypto service")
	metadata := &dom_file.FileMetadata{
		Name:                   fileName,
		MimeType:               mimeType,
		Size:                   fileInfo.Size,
		Created:                fileInfo.ModifiedAt.Unix(),
		FileExtension:          fileExtension, // Explicitly store file extension
		DecryptedFilePath:      destFilePath,
		DecryptedFileSize:      fileInfo.Size,
		EncryptedFilePath:      encryptedPath,
		EncryptedFileSize:      int64(len(encryptedFileData)),
		EncryptedThumbnailPath: "", // Developer Note: Future feature
		EncryptedThumbnailSize: 0,  // Developer Note: Future feature
		DecryptedThumbnailPath: "", // Developer Note: Future feature
		DecryptedThumbnailSize: 0,  // Developer Note: Future feature
	}

	encryptedMetadataString, err := s.fileEncryptionService.EncryptFileMetadata(ctx, metadata, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file metadata", err)
	}

	s.logger.Debug("🔐 Computing and encrypting file hash")
	fileHashBytes, err := s.computeFileHashUseCase.ExecuteForBytes(ctx, destFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to compute file hash", err)
	}

	// Encrypt the hash using the same pattern as other file encryption service methods
	encryptedHashData, err := s.fileEncryptionService.EncryptFileContent(ctx, fileHashBytes, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file hash", err)
	}
	encryptedHashString := crypto.EncodeToBase64(encryptedHashData)

	//
	// STEP 8: Create domain file object
	//
	currentTime := time.Now()
	domainFile := &dom_file.File{
		ID:                fileID,
		CollectionID:      input.CollectionID,
		OwnerID:           input.OwnerID,
		EncryptedMetadata: encryptedMetadataString,
		EncryptedFileKey:  *encryptedFileKey, // Use the struct from crypto service
		EncryptionVersion: "1.0",
		EncryptedHash:     encryptedHashString,
		EncryptedFilePath: encryptedPath,
		EncryptedFileSize: int64(len(encryptedFileData)),
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
		s.logger.Error("❌ Failed to create file record", zap.String("fileID", fileID.String()), zap.Error(err))
		return nil, errors.NewAppError("failed to create file record", err)
	}

	s.logger.Info("✅ Successfully added E2EE file using crypto services",
		zap.String("fileID", fileID.String()),
		zap.String("fileName", fileName),
		zap.String("copiedPath", destFilePath))

	return &LocalFileAddOutput{
		File:           domainFile,
		CopiedFilePath: destFilePath,
	}, nil
}
