// internal/service/localfile/lock.go
package localfile

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	svc_collectioncrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/collectioncrypto"
	svc_filecrypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filecrypto"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
)

// LockInput represents the input for locking a file (keeping only encrypted version)
type LockInput struct {
	FileID   string `json:"file_id"`
	Password string `json:"password"`
}

// LockOutput represents the result of locking a file
type LockOutput struct {
	FileID         gocql.UUID          `json:"file_id"`
	PreviousMode   string              `json:"previous_mode"`
	NewMode        string              `json:"new_mode"`
	PreviousStatus dom_file.SyncStatus `json:"previous_status"`
	DeletedPath    string              `json:"deleted_path"`
	RemainingPath  string              `json:"remaining_path"`
	Message        string              `json:"message"`
}

// LockService defines the interface for locking files (encrypted-only mode)
type LockService interface {
	Lock(ctx context.Context, input *LockInput) (*LockOutput, error)
}

// lockService implements the LockService interface
type lockService struct {
	logger                      *zap.Logger
	getFileUseCase              uc_file.GetFileUseCase
	updateFileUseCase           uc_file.UpdateFileUseCase
	deleteFileUseCase           localfile.DeleteFileUseCase
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase        uc_collection.GetCollectionUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
	fileDecryptionService       svc_filecrypto.FileDecryptionService
	fileEncryptionService       svc_filecrypto.FileEncryptionService
}

// NewLockService creates a new service for locking files
func NewLockService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	deleteFileUseCase localfile.DeleteFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	fileDecryptionService svc_filecrypto.FileDecryptionService,
	fileEncryptionService svc_filecrypto.FileEncryptionService,
) LockService {
	logger = logger.Named("LockService")
	return &lockService{
		logger:                      logger,
		getFileUseCase:              getFileUseCase,
		updateFileUseCase:           updateFileUseCase,
		deleteFileUseCase:           deleteFileUseCase,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		getCollectionUseCase:        getCollectionUseCase,
		collectionDecryptionService: collectionDecryptionService,
		fileDecryptionService:       fileDecryptionService,
		fileEncryptionService:       fileEncryptionService,
	}
}

// Lock handles locking a file using E2EE (keeping only encrypted version)
func (s *lockService) Lock(ctx context.Context, input *LockInput) (*LockOutput, error) {
	//
	// STEP 1: Validate inputs
	//
	if input == nil {
		s.logger.Error("❌ input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.FileID == "" {
		s.logger.Error("❌ file ID is required")
		return nil, errors.NewAppError("file ID is required", nil)
	}
	if input.Password == "" {
		s.logger.Error("❌ password is required for E2EE operations")
		return nil, errors.NewAppError("password is required for E2EE operations", nil)
	}

	//
	// STEP 2: Convert file ID string to ObjectID
	//
	fileObjectID, err := primitive.ObjectIDFromHex(input.FileID)
	if err != nil {
		s.logger.Error("❌ invalid file ID format",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("invalid file ID format", err)
	}

	//
	// STEP 3: Get file, user, and collection for E2EE operations
	//
	s.logger.Debug("🔍 Getting file for lock operation",
		zap.String("fileID", input.FileID))

	file, err := s.getFileUseCase.Execute(ctx, fileObjectID)
	if err != nil {
		s.logger.Error("❌ failed to get file",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to get file", err)
	}

	if file == nil {
		s.logger.Error("❌ file not found", zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("file not found", nil)
	}

	// Get user for E2EE key chain
	user, err := s.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		return nil, errors.NewAppError("failed to get logged in user", err)
	}
	if user == nil {
		return nil, errors.NewAppError("user not found", nil)
	}

	// Get collection for E2EE key chain
	collection, err := s.getCollectionUseCase.Execute(ctx, file.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if collection == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	previousMode := file.StorageMode
	previousStatus := file.SyncStatus

	//
	// STEP 4: Validate file status
	//
	if file.SyncStatus == dom_file.SyncStatusCloudOnly {
		s.logger.Error("❌ cannot lock cloud-only file",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError("cannot lock a cloud-only file. File must have local decrypted version.", nil)
	}

	if file.StorageMode == dom_file.StorageModeEncryptedOnly {
		s.logger.Info("✅ file is already locked (encrypted-only)",
			zap.String("fileID", input.FileID))
		return &LockOutput{
			FileID:         fileObjectID,
			PreviousMode:   previousMode,
			NewMode:        dom_file.StorageModeEncryptedOnly,
			PreviousStatus: previousStatus,
			RemainingPath:  file.EncryptedFilePath,
			Message:        "File is already locked (encrypted-only mode)",
		}, nil
	}

	//
	// STEP 5: Validate decrypted file exists (we need it to encrypt)
	//
	if file.FilePath == "" {
		s.logger.Error("❌ no decrypted file path available",
			zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("no decrypted file available to lock", nil)
	}

	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		s.logger.Error("❌ decrypted file does not exist",
			zap.String("fileID", input.FileID),
			zap.String("filePath", file.FilePath))
		return nil, errors.NewAppError("decrypted file does not exist on disk", nil)
	}

	//
	// STEP 6: Starting E2EE key chain decryption for lock operation using crypto service
	//
	s.logger.Debug("🔐 Starting E2EE key chain decryption for lock operation using crypto service")

	collectionKey, err := s.collectionDecryptionService.ExecuteDecryptCollectionKeyChain(ctx, user, collection, input.Password)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain using crypto service", err)
	}
	defer crypto.ClearBytes(collectionKey)

	fileKey, err := s.fileDecryptionService.DecryptFileKey(ctx, file.EncryptedFileKey, collectionKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file key using crypto service", err)
	}
	defer crypto.ClearBytes(fileKey)

	//
	// STEP 7: Encrypting file content for locking using crypto service
	//
	s.logger.Info("🔒 Encrypting file content for locking using crypto service",
		zap.String("fileID", input.FileID))

	// Read file content
	fileContent, err := os.ReadFile(file.FilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to read file for encryption", err)
	}

	encryptedData, err := s.fileEncryptionService.EncryptFileContent(ctx, fileContent, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to encrypt file content using crypto service", err)
	}

	//
	// STEP 8: Save encrypted version (or verify it exists and is correct)
	//
	encryptedPath := file.FilePath + ".encrypted"
	if file.EncryptedFilePath != "" {
		encryptedPath = file.EncryptedFilePath
	}

	if err := os.WriteFile(encryptedPath, encryptedData, 0600); err != nil {
		return nil, errors.NewAppError("failed to save encrypted file", err)
	}

	s.logger.Debug("✅ Successfully saved encrypted file using crypto service",
		zap.String("fileID", input.FileID),
		zap.String("encryptedPath", encryptedPath))

	//
	// STEP 9: Delete decrypted file
	//
	var deletedPath string
	s.logger.Info("🗑️ Deleting decrypted file for lock operation",
		zap.String("fileID", input.FileID),
		zap.String("filePath", file.FilePath))

	if err := s.deleteFileUseCase.Execute(ctx, file.FilePath); err != nil {
		s.logger.Warn("⚠️ Failed to delete decrypted file",
			zap.String("fileID", input.FileID),
			zap.String("filePath", file.FilePath),
			zap.Error(err))
		// Continue anyway, we'll still update the storage mode
	} else {
		deletedPath = file.FilePath
		s.logger.Debug("✅ Successfully deleted decrypted file",
			zap.String("fileID", input.FileID),
			zap.String("filePath", file.FilePath))
	}

	//
	// STEP 10: Update file record to encrypted-only mode
	//
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
		// Developers note: We don't need to update the state, this is a strict local feature that doesn't affect the distributed clients and doesn't affect the cloud state.
	}

	newMode := dom_file.StorageModeEncryptedOnly
	updateInput.StorageMode = &newMode
	updateInput.EncryptedFilePath = &encryptedPath

	// Clear the decrypted file path
	emptyPath := ""
	updateInput.FilePath = &emptyPath

	_, err = s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to update file storage mode during lock",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file storage mode during lock", err)
	}

	s.logger.Info("🎉 Successfully locked file using E2EE crypto services",
		zap.String("fileID", input.FileID),
		zap.String("previousMode", previousMode),
		zap.String("newMode", newMode))

	return &LockOutput{
		FileID:         fileObjectID,
		PreviousMode:   previousMode,
		NewMode:        newMode,
		PreviousStatus: previousStatus,
		DeletedPath:    deletedPath,
		RemainingPath:  encryptedPath,
		Message:        "File successfully locked to encrypted-only mode using E2EE crypto services",
	}, nil
}
