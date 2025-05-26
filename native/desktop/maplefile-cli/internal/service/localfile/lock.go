// internal/service/localfile/lock.go
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

// LockInput represents the input for locking a file (keeping only encrypted version)
type LockInput struct {
	FileID string `json:"file_id"`
}

// LockOutput represents the result of locking a file
type LockOutput struct {
	FileID         primitive.ObjectID  `json:"file_id"`
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
	logger            *zap.Logger
	getFileUseCase    uc_file.GetFileUseCase
	updateFileUseCase uc_file.UpdateFileUseCase
	deleteFileUseCase localfile.DeleteFileUseCase
}

// NewLockService creates a new service for locking files
func NewLockService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	deleteFileUseCase localfile.DeleteFileUseCase,
) LockService {
	return &lockService{
		logger:            logger,
		getFileUseCase:    getFileUseCase,
		updateFileUseCase: updateFileUseCase,
		deleteFileUseCase: deleteFileUseCase,
	}
}

// Lock handles locking a file (keeping only encrypted version)
func (s *lockService) Lock(ctx context.Context, input *LockInput) (*LockOutput, error) {
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
	s.logger.Debug("Getting file for lock operation",
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
		s.logger.Error("cannot lock cloud-only file",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError("cannot lock a cloud-only file. File must have local encrypted version.", nil)
	}

	if file.StorageMode == dom_file.StorageModeEncryptedOnly {
		s.logger.Info("file is already locked (encrypted-only)",
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
	// STEP 5: Validate encrypted file exists
	//
	if file.EncryptedFilePath == "" {
		s.logger.Error("no encrypted file path available",
			zap.String("fileID", input.FileID))
		return nil, errors.NewAppError("no encrypted file available to lock to", nil)
	}

	if _, err := os.Stat(file.EncryptedFilePath); os.IsNotExist(err) {
		s.logger.Error("encrypted file does not exist",
			zap.String("fileID", input.FileID),
			zap.String("encryptedPath", file.EncryptedFilePath))
		return nil, errors.NewAppError("encrypted file does not exist on disk", nil)
	}

	//
	// STEP 6: Delete decrypted file if it exists
	//
	var deletedPath string
	if file.FilePath != "" {
		if _, err := os.Stat(file.FilePath); err == nil {
			s.logger.Info("Deleting decrypted file",
				zap.String("fileID", input.FileID),
				zap.String("filePath", file.FilePath))

			if err := s.deleteFileUseCase.Execute(ctx, file.FilePath); err != nil {
				s.logger.Warn("Failed to delete decrypted file",
					zap.String("fileID", input.FileID),
					zap.String("filePath", file.FilePath),
					zap.Error(err))
				// Continue anyway, we'll still update the storage mode
			} else {
				deletedPath = file.FilePath
				s.logger.Debug("Successfully deleted decrypted file",
					zap.String("fileID", input.FileID),
					zap.String("filePath", file.FilePath))
			}
		}
	}

	//
	// STEP 7: Update file record to encrypted-only mode
	//
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
	}

	newMode := dom_file.StorageModeEncryptedOnly
	updateInput.StorageMode = &newMode

	// Clear the decrypted file path
	emptyPath := ""
	updateInput.FilePath = &emptyPath

	_, err = s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("failed to update file storage mode during lock",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file storage mode during lock", err)
	}

	s.logger.Info("Successfully locked file",
		zap.String("fileID", input.FileID),
		zap.String("previousMode", previousMode),
		zap.String("newMode", newMode))

	return &LockOutput{
		FileID:         fileObjectID,
		PreviousMode:   previousMode,
		NewMode:        newMode,
		PreviousStatus: previousStatus,
		DeletedPath:    deletedPath,
		RemainingPath:  file.EncryptedFilePath,
		Message:        "File successfully locked to encrypted-only mode",
	}, nil
}
