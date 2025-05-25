// internal/service/filesyncer/offload.go
package filesyncer

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/fileupload"
	svc_fileupload "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/fileupload"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// OffloadInput represents the input for offloading a local file
type OffloadInput struct {
	FileID       string `json:"file_id"`
	UserPassword string `json:"user_password"`
}

// OffloadOutput represents the result of offloading a local file
type OffloadOutput struct {
	FileID         primitive.ObjectID           `json:"file_id"`
	Action         string                       `json:"action"` // "uploaded" or "offloaded"
	PreviousStatus dom_file.SyncStatus          `json:"previous_status"`
	NewStatus      dom_file.SyncStatus          `json:"new_status"`
	UploadResult   *fileupload.FileUploadResult `json:"upload_result,omitempty"`
	Message        string                       `json:"message"`
}

// OffloadService defines the interface for offloading local files
type OffloadService interface {
	Offload(ctx context.Context, input *OffloadInput) (*OffloadOutput, error)
}

// offloadService implements the OffloadService interface
type offloadService struct {
	logger            *zap.Logger
	getFileUseCase    uc_file.GetFileUseCase
	updateFileUseCase uc_file.UpdateFileUseCase
	uploadService     svc_fileupload.UploadService
	deleteFileUseCase localfile.DeleteFileUseCase
}

// NewOffloadService creates a new service for offloading local files
func NewOffloadService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	uploadService svc_fileupload.UploadService,
	deleteFileUseCase localfile.DeleteFileUseCase,
) OffloadService {
	return &offloadService{
		logger:            logger,
		getFileUseCase:    getFileUseCase,
		updateFileUseCase: updateFileUseCase,
		uploadService:     uploadService,
		deleteFileUseCase: deleteFileUseCase,
	}
}

// Offload handles the offloading of a local file to the cloud
func (s *offloadService) Offload(ctx context.Context, input *OffloadInput) (*OffloadOutput, error) {
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
	// STEP 3: Get the file to check its current sync status
	//
	s.logger.Debug("Getting file for offload operation",
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

	//
	// STEP 4: Handle different sync statuses
	//
	switch file.SyncStatus {
	case dom_file.SyncStatusLocalOnly, dom_file.SyncStatusModifiedLocally:
		// File needs to be uploaded first
		return s.handleUploadAndOffload(ctx, file, input.UserPassword, previousStatus)

	case dom_file.SyncStatusSynced:
		// File is already synced, just offload (delete local decrypted copy)
		return s.handleOffloadOnly(ctx, file, previousStatus)

	case dom_file.SyncStatusCloudOnly:
		// File is already offloaded
		s.logger.Info("File is already offloaded", zap.String("fileID", input.FileID))
		return &OffloadOutput{
			FileID:         fileObjectID,
			Action:         "no_action",
			PreviousStatus: previousStatus,
			NewStatus:      dom_file.SyncStatusCloudOnly,
			Message:        "File is already offloaded to cloud only",
		}, nil

	default:
		s.logger.Error("unknown sync status",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError("unknown sync status", nil)
	}
}

// handleUploadAndOffload uploads the file first, then offloads it
func (s *offloadService) handleUploadAndOffload(
	ctx context.Context,
	file *dom_file.File,
	userPassword string,
	previousStatus dom_file.SyncStatus,
) (*OffloadOutput, error) {
	s.logger.Info("Uploading file before offload", zap.String("fileID", file.ID.Hex()))

	// Upload the file first
	uploadResult, err := s.uploadService.UploadFile(ctx, file.ID, userPassword)
	if err != nil {
		s.logger.Error("failed to upload file during offload",
			zap.String("fileID", file.ID.Hex()),
			zap.Error(err))
		return nil, errors.NewAppError("failed to upload file during offload", err)
	}

	if !uploadResult.Success {
		s.logger.Error("file upload was not successful",
			zap.String("fileID", file.ID.Hex()))
		return nil, errors.NewAppError("file upload was not successful", uploadResult.Error)
	}

	// Refresh file data after upload
	refreshedFile, err := s.getFileUseCase.Execute(ctx, file.ID)
	if err != nil {
		s.logger.Error("failed to refresh file after upload",
			zap.String("fileID", file.ID.Hex()),
			zap.Error(err))
		return nil, errors.NewAppError("failed to refresh file after upload", err)
	}

	// Now offload the uploaded file
	return s.handleOffloadOnly(ctx, refreshedFile, previousStatus)
}

// handleOffloadOnly removes both encrypted and decrypted local files and updates sync status
func (s *offloadService) handleOffloadOnly(
	ctx context.Context,
	file *dom_file.File,
	previousStatus dom_file.SyncStatus,
) (*OffloadOutput, error) {
	s.logger.Info("Offloading file (removing local encrypted and decrypted copies)",
		zap.String("fileID", file.ID.Hex()))

	// Delete the decrypted local file if it exists
	if file.FilePath != "" {
		if _, err := os.Stat(file.FilePath); err == nil {
			// File exists, delete it
			if err := s.deleteFileUseCase.Execute(ctx, file.FilePath); err != nil {
				s.logger.Warn("Failed to delete decrypted local file",
					zap.String("fileID", file.ID.Hex()),
					zap.String("filePath", file.FilePath),
					zap.Error(err))
				// Continue anyway, we'll still update the sync status
			} else {
				s.logger.Debug("Successfully deleted decrypted local file",
					zap.String("fileID", file.ID.Hex()),
					zap.String("filePath", file.FilePath))
			}
		}
	}

	// Delete the encrypted local file if it exists
	if file.EncryptedFilePath != "" {
		if _, err := os.Stat(file.EncryptedFilePath); err == nil {
			// File exists, delete it
			if err := s.deleteFileUseCase.Execute(ctx, file.EncryptedFilePath); err != nil {
				s.logger.Warn("Failed to delete encrypted local file",
					zap.String("fileID", file.ID.Hex()),
					zap.String("encryptedFilePath", file.EncryptedFilePath),
					zap.Error(err))
				// Continue anyway, we'll still update the sync status
			} else {
				s.logger.Debug("Successfully deleted encrypted local file",
					zap.String("fileID", file.ID.Hex()),
					zap.String("encryptedFilePath", file.EncryptedFilePath))
			}
		}
	}

	// Update sync status to CloudOnly and clear both file paths
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
	}

	newStatus := dom_file.SyncStatusCloudOnly
	updateInput.SyncStatus = &newStatus

	// Clear both encrypted and decrypted file paths
	emptyPath := ""
	updateInput.FilePath = &emptyPath
	updateInput.EncryptedFilePath = &emptyPath

	_, err := s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("failed to update file sync status during offload",
			zap.String("fileID", file.ID.Hex()),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file sync status during offload", err)
	}

	s.logger.Info("Successfully offloaded file",
		zap.String("fileID", file.ID.Hex()),
		zap.Any("previousStatus", previousStatus),
		zap.Any("newStatus", newStatus))

	return &OffloadOutput{
		FileID:         file.ID,
		Action:         "offloaded",
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Message:        "File successfully offloaded to cloud only",
	}, nil
}
