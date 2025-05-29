// internal/service/filesyncer/cloud_delete.go
package filesyncer

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/filedto"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
)

// CloudOnlyDeleteInput represents the input for deleting a file from cloud
type CloudOnlyDeleteInput struct {
	FileID       string `json:"file_id"`
	UserPassword string `json:"user_password"`
}

// CloudOnlyDeleteOutput represents the result of deleting a file from cloud
type CloudOnlyDeleteOutput struct {
	FileID         primitive.ObjectID  `json:"file_id"`
	PreviousStatus dom_file.SyncStatus `json:"previous_status"`
	NewStatus      dom_file.SyncStatus `json:"new_status"`
	Action         string              `json:"action"`
	Message        string              `json:"message"`
}

// CloudOnlyDeleteService defines the interface for deleting files from cloud
type CloudOnlyDeleteService interface {
	DeleteFromCloud(ctx context.Context, input *CloudOnlyDeleteInput) (*CloudOnlyDeleteOutput, error)
}

// cloudOnlyDeleteService implements the CloudOnlyDeleteService interface
type cloudOnlyDeleteService struct {
	logger            *zap.Logger
	getFileUseCase    uc_file.GetFileUseCase
	updateFileUseCase uc_file.UpdateFileUseCase
	fileDTORepo       filedto.FileDTORepository
}

// NewCloudOnlyDeleteService creates a new service for deleting files from cloud
func NewCloudOnlyDeleteService(
	logger *zap.Logger,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	fileDTORepo filedto.FileDTORepository,
) CloudOnlyDeleteService {
	logger = logger.Named("CloudOnlyDeleteService")
	return &cloudOnlyDeleteService{
		logger:            logger,
		getFileUseCase:    getFileUseCase,
		updateFileUseCase: updateFileUseCase,
		fileDTORepo:       fileDTORepo,
	}
}

// DeleteFromCloud handles the deletion of a file from cloud and updates local sync status
func (s *cloudOnlyDeleteService) DeleteFromCloud(ctx context.Context, input *CloudOnlyDeleteInput) (*CloudOnlyDeleteOutput, error) {
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
	if input.UserPassword == "" {
		s.logger.Error("❌ user password is required for authentication")
		return nil, errors.NewAppError("user password is required for authentication", nil)
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
	// STEP 3: Get the file to check its current sync status
	//
	s.logger.Debug("🔍 Getting file for cloud delete operation",
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

	previousStatus := file.SyncStatus

	//
	// STEP 4: Validate file sync status and provide detailed feedback
	//
	s.logger.Info("ℹ️ Checking file sync status for cloud deletion",
		zap.String("fileID", input.FileID),
		zap.Any("currentSyncStatus", file.SyncStatus),
		zap.String("fileName", file.Name))

	switch file.SyncStatus {
	case dom_file.SyncStatusLocalOnly:
		s.logger.Warn("⚠️ File is local-only, cannot delete from cloud",
			zap.String("fileID", input.FileID),
			zap.String("fileName", file.Name),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError(fmt.Sprintf("file '%s' is local-only and does not exist in cloud storage", file.Name), nil)

	case dom_file.SyncStatusModifiedLocally:
		s.logger.Warn("⚠️ File has local modifications, deletion from cloud may cause data loss",
			zap.String("fileID", input.FileID),
			zap.String("fileName", file.Name),
			zap.Any("syncStatus", file.SyncStatus))
		// Continue with deletion but log warning

	case dom_file.SyncStatusSynced, dom_file.SyncStatusCloudOnly:
		s.logger.Info("✅ File is eligible for cloud deletion",
			zap.String("fileID", input.FileID),
			zap.String("fileName", file.Name),
			zap.Any("syncStatus", file.SyncStatus))

	default:
		s.logger.Error("❌ unknown sync status",
			zap.String("fileID", input.FileID),
			zap.String("fileName", file.Name),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError(fmt.Sprintf("unknown sync status for file '%s': %v", file.Name, file.SyncStatus), nil)
	}

	//
	// STEP 5: Delete file from cloud backend
	//
	s.logger.Info("☁️ Deleting file from cloud backend",
		zap.String("fileID", input.FileID),
		zap.Any("previousStatus", previousStatus))

	err = s.fileDTORepo.DeleteByIDFromCloud(ctx, fileObjectID)
	if err != nil {
		s.logger.Error("❌ failed to delete file from cloud",
			zap.String("fileID", input.FileID),
			zap.String("fileObjectID", fileObjectID.Hex()),
			zap.Any("fileSyncStatus", file.SyncStatus),
			zap.Error(err))

		// Provide more specific error messages
		if err.Error() == "file not found in cloud" {
			return nil, errors.NewAppError("file does not exist in cloud storage", err)
		}

		return nil, errors.NewAppError("failed to delete file from cloud", err)
	}

	//
	// STEP 6: Update local file sync status to LocalOnly
	//
	newStatus := dom_file.SyncStatusLocalOnly
	newVersion := file.Version + 1
	updateInput := uc_file.UpdateFileInput{
		ID:         file.ID,
		SyncStatus: &newStatus,
		Version:    &newVersion,
	}

	_, err = s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("❌ failed to update file sync status after cloud deletion",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file sync status after cloud deletion", err)
	}

	s.logger.Info("🎉 Successfully deleted file from cloud and updated sync status",
		zap.String("fileID", input.FileID),
		zap.Any("previousStatus", previousStatus),
		zap.Any("newStatus", newStatus))

	return &CloudOnlyDeleteOutput{
		FileID:         fileObjectID,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		Action:         "cloud_deleted",
		Message:        "File successfully deleted from cloud and sync status updated to local-only",
	}, nil
}
