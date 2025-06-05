// internal/service/localfile/unlock.go
package localfile

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

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

// UnlockInput represents the input for unlocking a file
type UnlockInput struct {
	FileID      string `json:"file_id"`
	Password    string `json:"password"`
	StorageMode string `json:"storage_mode"` // "decrypted_only" or "hybrid"
}

// UnlockOutput represents the result of unlocking a file
type UnlockOutput struct {
	FileID         primitive.ObjectID  `json:"file_id"`
	PreviousMode   string              `json:"previous_mode"`
	NewMode        string              `json:"new_mode"`
	PreviousStatus dom_file.SyncStatus `json:"previous_status"`
	DeletedPath    string              `json:"deleted_path,omitempty"`
	RemainingPath  string              `json:"remaining_path"`
	Message        string              `json:"message"`
}

// UnlockService defines the interface for unlocking files
type UnlockService interface {
	Unlock(ctx context.Context, input *UnlockInput) (*UnlockOutput, error)
}

// unlockService implements the UnlockService interface
type unlockService struct {
	logger                      *zap.Logger
	getFileUseCase              uc_file.GetFileUseCase
	updateFileUseCase           uc_file.UpdateFileUseCase
	deleteFileUseCase           localfile.DeleteFileUseCase
	getUserByIsLoggedInUseCase  uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase        uc_collection.GetCollectionUseCase
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService
	fileDecryptionService       svc_filecrypto.FileDecryptionService
}

// NewUnlockService creates a new service for unlocking files
func NewUnlockService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	deleteFileUseCase localfile.DeleteFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
	collectionDecryptionService svc_collectioncrypto.CollectionDecryptionService,
	fileDecryptionService svc_filecrypto.FileDecryptionService,
) UnlockService {
	logger = logger.Named("UnlockService")
	return &unlockService{
		logger:                      logger,
		getFileUseCase:              getFileUseCase,
		updateFileUseCase:           updateFileUseCase,
		deleteFileUseCase:           deleteFileUseCase,
		getUserByIsLoggedInUseCase:  getUserByIsLoggedInUseCase,
		getCollectionUseCase:        getCollectionUseCase,
		collectionDecryptionService: collectionDecryptionService,
		fileDecryptionService:       fileDecryptionService,
	}
}

// Unlock handles unlocking a file using E2EE (accessing decrypted content)
func (s *unlockService) Unlock(ctx context.Context, input *UnlockInput) (*UnlockOutput, error) {
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
	if input.StorageMode == "" {
		input.StorageMode = dom_file.StorageModeHybrid // Safe default
	}
	if input.StorageMode != dom_file.StorageModeDecryptedOnly && input.StorageMode != dom_file.StorageModeHybrid {
		s.logger.Error("❌ invalid storage mode", zap.String("storageMode", input.StorageMode))
		return nil, errors.NewAppError("storage mode must be 'decrypted_only' or 'hybrid'", nil)
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
	s.logger.Debug("🔍 Getting file for unlock operation",
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
		s.logger.Error("❌ cannot unlock cloud-only file",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError("cannot unlock a cloud-only file. Use 'filesync onload' first.", nil)
	}

	// Check if already in desired mode
	if file.StorageMode == input.StorageMode {
		s.logger.Info("ℹ️ file is already in desired storage mode",
			zap.String("fileID", input.FileID),
			zap.String("storageMode", input.StorageMode))
		return &UnlockOutput{
			FileID:         fileObjectID,
			PreviousMode:   previousMode,
			NewMode:        input.StorageMode,
			PreviousStatus: previousStatus,
			RemainingPath:  file.FilePath,
			Message:        "File is already in the desired storage mode",
		}, nil
	}

	//
	// STEP 5: Validate encrypted file exists (we need it to decrypt)
	//
	if file.EncryptedFilePath == "" {
		s.logger.Error("❌ no encrypted file path available",
			zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("no encrypted file available to unlock", nil)
	}

	if _, err := os.Stat(file.EncryptedFilePath); os.IsNotExist(err) {
		s.logger.Error("❌ encrypted file does not exist",
			zap.String("fileID", input.FileID),
			zap.String("encryptedPath", file.EncryptedFilePath))
		return nil, errors.NewAppError("encrypted file does not exist on disk", nil)
	}

	//
	// STEP 6: Starting E2EE key chain decryption for unlock operation using crypto service
	//
	s.logger.Debug("🔐 Starting E2EE key chain decryption for unlock operation using crypto service")

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
	// STEP 7: Decrypting file content for unlocking using crypto service
	//
	s.logger.Info("🔐 Decrypting file content for unlocking using crypto service",
		zap.String("fileID", input.FileID))

	// Read encrypted file content
	encryptedData, err := os.ReadFile(file.EncryptedFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to read encrypted file", err)
	}

	decryptedData, err := s.fileDecryptionService.DecryptFileContent(ctx, encryptedData, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file content using crypto service", err)
	}

	//
	// STEP 8: Save decrypted version
	//
	decryptedPath := file.EncryptedFilePath
	if len(decryptedPath) > 10 && decryptedPath[len(decryptedPath)-10:] == ".encrypted" {
		decryptedPath = decryptedPath[:len(decryptedPath)-10] // Remove .encrypted extension
	}
	if file.FilePath != "" {
		decryptedPath = file.FilePath
	}

	if err := os.WriteFile(decryptedPath, decryptedData, 0644); err != nil {
		return nil, errors.NewAppError("failed to save decrypted file", err)
	}

	s.logger.Debug("💾 Successfully saved decrypted file using crypto service",
		zap.String("fileID", input.FileID),
		zap.String("decryptedPath", decryptedPath))

	//
	// STEP 9: Delete encrypted file if switching to decrypted-only mode
	//
	var deletedPath string
	if input.StorageMode == dom_file.StorageModeDecryptedOnly {
		s.logger.Info("🗑️ Deleting encrypted file for decrypted-only mode",
			zap.String("fileID", input.FileID),
			zap.String("encryptedPath", file.EncryptedFilePath))

		if err := s.deleteFileUseCase.Execute(ctx, file.EncryptedFilePath); err != nil {
			s.logger.Warn("⚠️ Failed to delete encrypted file",
				zap.String("fileID", input.FileID),
				zap.String("encryptedPath", file.EncryptedFilePath),
				zap.Error(err))
			// Continue anyway, we'll still update the storage mode
		} else {
			deletedPath = file.EncryptedFilePath
			s.logger.Debug("🗑️ Successfully deleted encrypted file",
				zap.String("fileID", input.FileID),
				zap.String("encryptedPath", file.EncryptedFilePath))
		}
	}

	//
	// STEP 10: Update file record to new storage mode
	//
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
		// Developers note: We don't need to update the state, this is a strict local feature that doesn't affect the distributed clients and doesn't affect the cloud state.
	}

	updateInput.StorageMode = &input.StorageMode
	updateInput.FilePath = &decryptedPath

	// Clear encrypted file path only for decrypted-only mode
	if input.StorageMode == dom_file.StorageModeDecryptedOnly {
		emptyPath := ""
		updateInput.EncryptedFilePath = &emptyPath
	}

	_, err = s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to update file storage mode during unlock",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file storage mode during unlock", err)
	}

	s.logger.Info("✅ Successfully unlocked file using E2EE crypto services",
		zap.String("fileID", input.FileID),
		zap.String("previousMode", previousMode),
		zap.String("newMode", input.StorageMode))

	message := "File successfully unlocked using E2EE crypto services"
	if input.StorageMode == dom_file.StorageModeHybrid {
		message = "File successfully unlocked to hybrid mode using E2EE crypto services (both encrypted and decrypted versions kept)"
	} else {
		message = "File successfully unlocked to decrypted-only mode using E2EE crypto services"
	}

	return &UnlockOutput{
		FileID:         fileObjectID,
		PreviousMode:   previousMode,
		NewMode:        input.StorageMode,
		PreviousStatus: previousStatus,
		DeletedPath:    deletedPath,
		RemainingPath:  decryptedPath,
		Message:        message,
	}, nil
}
