// internal/service/localfile/unlock.go
package localfile

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	dom_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/user"
	uc_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collection"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
	pkg_crypto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/pkg/crypto"
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
	logger                     *zap.Logger
	getFileUseCase             uc_file.GetFileUseCase
	updateFileUseCase          uc_file.UpdateFileUseCase
	deleteFileUseCase          localfile.DeleteFileUseCase
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getCollectionUseCase       uc_collection.GetCollectionUseCase
}

// NewUnlockService creates a new service for unlocking files
func NewUnlockService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	deleteFileUseCase localfile.DeleteFileUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getCollectionUseCase uc_collection.GetCollectionUseCase,
) UnlockService {
	logger = logger.Named("UnlockService")
	return &unlockService{
		logger:                     logger,
		getFileUseCase:             getFileUseCase,
		updateFileUseCase:          updateFileUseCase,
		deleteFileUseCase:          deleteFileUseCase,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		getCollectionUseCase:       getCollectionUseCase,
	}
}

// Unlock handles unlocking a file using E2EE (accessing decrypted content)
func (s *unlockService) Unlock(ctx context.Context, input *UnlockInput) (*UnlockOutput, error) {
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
	if input.Password == "" {
		s.logger.Error("password is required for E2EE operations")
		return nil, errors.NewAppError("password is required for E2EE operations", nil)
	}
	if input.StorageMode == "" {
		input.StorageMode = dom_file.StorageModeHybrid // Safe default
	}
	if input.StorageMode != dom_file.StorageModeDecryptedOnly && input.StorageMode != dom_file.StorageModeHybrid {
		s.logger.Error("invalid storage mode", zap.String("storageMode", input.StorageMode))
		return nil, errors.NewAppError("storage mode must be 'decrypted_only' or 'hybrid'", nil)
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
	// STEP 3: Get file, user, and collection for E2EE operations
	//
	s.logger.Debug("Getting file for unlock operation",
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
		s.logger.Error("cannot unlock cloud-only file",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError("cannot unlock a cloud-only file. Use 'filesync onload' first.", nil)
	}

	// Check if already in desired mode
	if file.StorageMode == input.StorageMode {
		s.logger.Info("file is already in desired storage mode",
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
		s.logger.Error("no encrypted file path available",
			zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("no encrypted file available to unlock", nil)
	}

	if _, err := os.Stat(file.EncryptedFilePath); os.IsNotExist(err) {
		s.logger.Error("encrypted file does not exist",
			zap.String("fileID", input.FileID),
			zap.String("encryptedPath", file.EncryptedFilePath))
		return nil, errors.NewAppError("encrypted file does not exist on disk", nil)
	}

	//
	// STEP 6: E2EE DECRYPTION CHAIN: password → keyEncryptionKey → masterKey → collectionKey → fileKey
	//
	s.logger.Debug("Starting E2EE key chain decryption for unlock operation")

	collectionKey, err := s.decryptCollectionKeyChain(user, collection, input.Password)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt collection key chain", err)
	}
	defer pkg_crypto.ClearBytes(collectionKey)

	// Decrypt file key using collection key
	fileKey, err := pkg_crypto.DecryptWithSecretBox(
		file.EncryptedFileKey.Ciphertext,
		file.EncryptedFileKey.Nonce,
		collectionKey,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file key", err)
	}
	defer pkg_crypto.ClearBytes(fileKey)

	//
	// STEP 7: Decrypt the encrypted file content using the file key
	//
	s.logger.Info("Decrypting file content for unlocking",
		zap.String("fileID", input.FileID))

	decryptedData, err := s.decryptFileContent(file.EncryptedFilePath, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file content", err)
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

	s.logger.Debug("Successfully saved decrypted file",
		zap.String("fileID", input.FileID),
		zap.String("decryptedPath", decryptedPath))

	//
	// STEP 9: Delete encrypted file if switching to decrypted-only mode
	//
	var deletedPath string
	if input.StorageMode == dom_file.StorageModeDecryptedOnly {
		s.logger.Info("Deleting encrypted file for decrypted-only mode",
			zap.String("fileID", input.FileID),
			zap.String("encryptedPath", file.EncryptedFilePath))

		if err := s.deleteFileUseCase.Execute(ctx, file.EncryptedFilePath); err != nil {
			s.logger.Warn("Failed to delete encrypted file",
				zap.String("fileID", input.FileID),
				zap.String("encryptedPath", file.EncryptedFilePath),
				zap.Error(err))
			// Continue anyway, we'll still update the storage mode
		} else {
			deletedPath = file.EncryptedFilePath
			s.logger.Debug("Successfully deleted encrypted file",
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
		s.logger.Error("failed to update file storage mode during unlock",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file storage mode during unlock", err)
	}

	s.logger.Info("Successfully unlocked file using E2EE",
		zap.String("fileID", input.FileID),
		zap.String("previousMode", previousMode),
		zap.String("newMode", input.StorageMode))

	message := "File successfully unlocked using E2EE"
	if input.StorageMode == dom_file.StorageModeHybrid {
		message = "File successfully unlocked to hybrid mode using E2EE (both encrypted and decrypted versions kept)"
	} else {
		message = "File successfully unlocked to decrypted-only mode using E2EE"
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

// decryptCollectionKeyChain decrypts the complete E2EE chain to get the collection key
func (s *unlockService) decryptCollectionKeyChain(user *dom_user.User, collection *dom_collection.Collection, password string) ([]byte, error) {
	s.logger.Debug("Starting E2EE key chain decryption",
		zap.String("userID", user.ID.Hex()),
		zap.String("collectionID", collection.ID.Hex()))

	// STEP 1: Derive keyEncryptionKey from password
	keyEncryptionKey, err := pkg_crypto.DeriveKeyFromPassword(password, user.PasswordSalt)
	if err != nil {
		return nil, errors.NewAppError("failed to derive key encryption key", err)
	}
	defer pkg_crypto.ClearBytes(keyEncryptionKey)

	// STEP 2: Decrypt masterKey with keyEncryptionKey
	masterKey, err := pkg_crypto.DecryptWithSecretBox(
		user.EncryptedMasterKey.Ciphertext,
		user.EncryptedMasterKey.Nonce,
		keyEncryptionKey,
	)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt master key - incorrect password?", err)
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
		return nil, errors.NewAppError("failed to decrypt collection key", err)
	}

	return collectionKey, nil
}

// decryptFileContent decrypts encrypted file content using the file key
func (s *unlockService) decryptFileContent(encryptedFilePath string, fileKey []byte) ([]byte, error) {
	s.logger.Debug("Decrypting file content", zap.String("encryptedFilePath", encryptedFilePath))

	// Read encrypted file content
	encryptedData, err := os.ReadFile(encryptedFilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to read encrypted file", err)
	}

	// The encrypted data should be in the format: nonce (24 bytes) + ciphertext
	const nonceSize = 24
	if len(encryptedData) < nonceSize {
		return nil, errors.NewAppError("encrypted file too short", nil)
	}

	// Extract nonce and ciphertext from combined data
	nonce := make([]byte, nonceSize)
	copy(nonce, encryptedData[:nonceSize])

	ciphertext := make([]byte, len(encryptedData)-nonceSize)
	copy(ciphertext, encryptedData[nonceSize:])

	// Decrypt the content
	decryptedData, err := pkg_crypto.DecryptWithSecretBox(ciphertext, nonce, fileKey)
	if err != nil {
		return nil, errors.NewAppError("failed to decrypt file content", err)
	}

	s.logger.Debug("Successfully decrypted file content",
		zap.Int("encryptedSize", len(encryptedData)),
		zap.Int("decryptedSize", len(decryptedData)))

	return decryptedData, nil
}
