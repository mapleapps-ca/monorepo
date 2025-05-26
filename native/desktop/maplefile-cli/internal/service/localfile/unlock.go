// internal/service/localfile/unlock.go
package localfile

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// UnlockInput represents the input for unlocking a file
type UnlockInput struct {
	FileID      string `json:"file_id"`
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
	logger            *zap.Logger
	getFileUseCase    uc_file.GetFileUseCase
	updateFileUseCase uc_file.UpdateFileUseCase
	deleteFileUseCase localfile.DeleteFileUseCase
}

// NewUnlockService creates a new service for unlocking files
func NewUnlockService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	deleteFileUseCase localfile.DeleteFileUseCase,
) UnlockService {
	return &unlockService{
		logger:            logger,
		getFileUseCase:    getFileUseCase,
		updateFileUseCase: updateFileUseCase,
		deleteFileUseCase: deleteFileUseCase,
	}
}

// Unlock handles unlocking a file (keeping decrypted version)
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
	if input.StorageMode == "" {
		input.StorageMode = dom_file.StorageModeDecryptedOnly
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
	// STEP 3: Get the file
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

	previousMode := file.StorageMode
	previousStatus := file.SyncStatus

	//
	// STEP 4: Validate file status
	//
	if file.SyncStatus == dom_file.SyncStatusCloudOnly {
		s.logger.Error("cannot unlock cloud-only file",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError("cannot unlock a cloud-only file. File must have local decrypted version.", nil)
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
	// STEP 5: Validate decrypted file exists
	//
	if file.FilePath == "" {
		s.logger.Error("no decrypted file path available",
			zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("no decrypted file available to unlock to", nil)
	}

	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		s.logger.Error("decrypted file does not exist",
			zap.String("fileID", input.FileID),
			zap.String("filePath", file.FilePath))
		return nil, errors.NewAppError("decrypted file does not exist on disk", nil)
	}

	//
	// STEP 6: Delete encrypted file if switching to decrypted-only mode
	//
	var deletedPath string
	if input.StorageMode == dom_file.StorageModeDecryptedOnly {
		if file.EncryptedFilePath != "" {
			if _, err := os.Stat(file.EncryptedFilePath); err == nil {
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
		}
	}

	//
	// STEP 7: Update file record to new storage mode
	//
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
	}

	updateInput.StorageMode = &input.StorageMode

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

	s.logger.Info("Successfully unlocked file",
		zap.String("fileID", input.FileID),
		zap.String("previousMode", previousMode),
		zap.String("newMode", input.StorageMode))

	message := "File successfully unlocked"
	if input.StorageMode == dom_file.StorageModeHybrid {
		message = "File successfully unlocked to hybrid mode (both encrypted and decrypted versions kept)"
	} else {
		message = "File successfully unlocked to decrypted-only mode"
	}

	return &UnlockOutput{
		FileID:         fileObjectID,
		PreviousMode:   previousMode,
		NewMode:        input.StorageMode,
		PreviousStatus: previousStatus,
		DeletedPath:    deletedPath,
		RemainingPath:  file.FilePath,
		Message:        message,
	}, nil
}
