// internal/service/filesyncer/onload.go
package filesyncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/config"
	dom_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/file"
	svc_filedownload "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/service/filedownload"
	uc_file "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/file"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// OnloadInput represents the input for onloading a cloud-only file
type OnloadInput struct {
	FileID       string `json:"file_id"`
	UserPassword string `json:"user_password"`
}

// OnloadOutput represents the result of onloading a cloud-only file
type OnloadOutput struct {
	FileID         primitive.ObjectID  `json:"file_id"`
	PreviousStatus dom_file.SyncStatus `json:"previous_status"`
	NewStatus      dom_file.SyncStatus `json:"new_status"`
	DecryptedPath  string              `json:"decrypted_path"`
	DownloadedSize int64               `json:"downloaded_size"`
	Message        string              `json:"message"`
}

// OnloadService defines the interface for onloading cloud-only files
type OnloadService interface {
	Onload(ctx context.Context, input *OnloadInput) (*OnloadOutput, error)
}

// onloadService implements the OnloadService interface
type onloadService struct {
	logger                 *zap.Logger
	configService          config.ConfigService
	getFileUseCase         uc_file.GetFileUseCase
	updateFileUseCase      uc_file.UpdateFileUseCase
	downloadService        svc_filedownload.DownloadService
	pathUtilsUseCase       localfile.PathUtilsUseCase
	createDirectoryUseCase localfile.CreateDirectoryUseCase
}

// NewOnloadService creates a new service for onloading cloud-only files
func NewOnloadService(
	logger *zap.Logger,
	configService config.ConfigService,
	getFileUseCase uc_file.GetFileUseCase,
	updateFileUseCase uc_file.UpdateFileUseCase,
	downloadService svc_filedownload.DownloadService,
	pathUtilsUseCase localfile.PathUtilsUseCase,
	createDirectoryUseCase localfile.CreateDirectoryUseCase,
) OnloadService {
	return &onloadService{
		logger:                 logger,
		configService:          configService,
		getFileUseCase:         getFileUseCase,
		updateFileUseCase:      updateFileUseCase,
		downloadService:        downloadService,
		pathUtilsUseCase:       pathUtilsUseCase,
		createDirectoryUseCase: createDirectoryUseCase,
	}
}

// Onload handles the onloading of a cloud-only file to local storage
func (s *onloadService) Onload(ctx context.Context, input *OnloadInput) (*OnloadOutput, error) {
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
	// STEP 3: Get the file and validate it's cloud-only
	//
	s.logger.Debug("Getting file for onload operation",
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

	// Only work with cloud-only files
	if file.SyncStatus != dom_file.SyncStatusCloudOnly {
		s.logger.Error("file is not cloud-only",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError(
			fmt.Sprintf("file is not cloud-only (current status: %v)", file.SyncStatus),
			nil)
	}

	//
	// STEP 4: Download and decrypt file using the download service
	//
	s.logger.Info("Downloading and decrypting file from cloud",
		zap.String("fileID", input.FileID))

	urlDuration := 1 * time.Hour // Default duration for download URLs
	downloadResult, err := s.downloadService.DownloadAndDecryptFile(ctx, fileObjectID, input.UserPassword, urlDuration)
	if err != nil {
		s.logger.Error("failed to download and decrypt file",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to download and decrypt file", err)
	}

	s.logger.Info("Successfully downloaded and decrypted file",
		zap.String("fileID", input.FileID),
		zap.String("fileName", downloadResult.DecryptedMetadata.Name),
		zap.Int64("size", downloadResult.OriginalSize))

	//
	// STEP 5: Save decrypted file locally
	//
	decryptedPath, err := s.saveDecryptedFile(ctx, file, downloadResult.DecryptedData, downloadResult.DecryptedMetadata.Name)
	if err != nil {
		s.logger.Error("failed to save decrypted file",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to save decrypted file", err)
	}

	//
	// STEP 6: Save thumbnail if present
	//
	if downloadResult.ThumbnailData != nil && len(downloadResult.ThumbnailData) > 0 {
		thumbnailPath, err := s.saveThumbnail(ctx, file, downloadResult.ThumbnailData, downloadResult.DecryptedMetadata.Name)
		if err != nil {
			s.logger.Warn("Failed to save thumbnail, continuing without it",
				zap.String("fileID", input.FileID),
				zap.Error(err))
		} else {
			s.logger.Debug("Successfully saved thumbnail",
				zap.String("fileID", input.FileID),
				zap.String("thumbnailPath", thumbnailPath))
		}
	}

	//
	// STEP 7: Update file record with new path and sync status
	//
	updateInput := uc_file.UpdateFileInput{
		ID: file.ID,
		// Developers note: We don't need to update the state, this is a strict local feature that doesn't affect the distributed clients and doesn't affect the cloud state.
	}

	newStatus := dom_file.SyncStatusSynced
	updateInput.SyncStatus = &newStatus
	updateInput.FilePath = &decryptedPath

	// Update the file name and MIME type from decrypted metadata
	if downloadResult.DecryptedMetadata.Name != "" {
		updateInput.DecryptedName = &downloadResult.DecryptedMetadata.Name
	}
	if downloadResult.DecryptedMetadata.MimeType != "" {
		updateInput.DecryptedMimeType = &downloadResult.DecryptedMetadata.MimeType
	}

	_, err = s.updateFileUseCase.Execute(ctx, updateInput)
	if err != nil {
		s.logger.Error("failed to update file sync status during onload",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file sync status during onload", err)
	}

	s.logger.Info("Successfully onloaded file",
		zap.String("fileID", input.FileID),
		zap.String("decryptedPath", decryptedPath),
		zap.Any("previousStatus", previousStatus),
		zap.Any("newStatus", newStatus))

	return &OnloadOutput{
		FileID:         fileObjectID,
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		DecryptedPath:  decryptedPath,
		DownloadedSize: downloadResult.OriginalSize,
		Message:        "File successfully onloaded and decrypted",
	}, nil
}

// saveDecryptedFile saves the decrypted file content to local storage
func (s *onloadService) saveDecryptedFile(ctx context.Context, file *dom_file.File, decryptedData []byte, originalFileName string) (string, error) {
	s.logger.Debug("Saving decrypted file locally", zap.String("fileID", file.ID.Hex()))

	// Get app data directory
	appDataDir, err := s.configService.GetAppDataDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get app data directory: %w", err)
	}

	// Create files storage directory structure
	filesDir := s.pathUtilsUseCase.Join(ctx, appDataDir, "files")
	binDir := s.pathUtilsUseCase.Join(ctx, filesDir, "bin")
	collectionDir := s.pathUtilsUseCase.Join(ctx, binDir, file.CollectionID.Hex())

	// Create directories if they don't exist
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, collectionDir); err != nil {
		return "", fmt.Errorf("failed to create collection directory: %w", err)
	}

	// Generate file path with original extension
	fileExtension := filepath.Ext(originalFileName)
	if fileExtension == "" {
		// Try to determine extension from MIME type if available
		fileExtension = s.getExtensionFromMimeType(file.MimeType)
	}

	destFileName := file.ID.Hex() + fileExtension
	destFilePath := s.pathUtilsUseCase.Join(ctx, collectionDir, destFileName)

	// Write the decrypted file
	err = os.WriteFile(destFilePath, decryptedData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}

	s.logger.Debug("Successfully saved decrypted file",
		zap.String("fileID", file.ID.Hex()),
		zap.String("filePath", destFilePath),
		zap.Int("size", len(decryptedData)))

	return destFilePath, nil
}

// saveThumbnail saves the decrypted thumbnail to local storage
func (s *onloadService) saveThumbnail(ctx context.Context, file *dom_file.File, thumbnailData []byte, originalFileName string) (string, error) {
	s.logger.Debug("Saving thumbnail locally", zap.String("fileID", file.ID.Hex()))

	// Get app data directory
	appDataDir, err := s.configService.GetAppDataDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get app data directory: %w", err)
	}

	// Create files storage directory structure
	filesDir := s.pathUtilsUseCase.Join(ctx, appDataDir, "files")
	binDir := s.pathUtilsUseCase.Join(ctx, filesDir, "bin")
	collectionDir := s.pathUtilsUseCase.Join(ctx, binDir, file.CollectionID.Hex())

	// Create directories if they don't exist
	if err := s.createDirectoryUseCase.ExecuteAll(ctx, collectionDir); err != nil {
		return "", fmt.Errorf("failed to create collection directory: %w", err)
	}

	// Generate thumbnail file name
	thumbnailFileName := file.ID.Hex() + "_thumbnail.jpg"
	thumbnailPath := s.pathUtilsUseCase.Join(ctx, collectionDir, thumbnailFileName)

	// Write the thumbnail
	err = os.WriteFile(thumbnailPath, thumbnailData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write thumbnail: %w", err)
	}

	s.logger.Debug("Successfully saved thumbnail",
		zap.String("fileID", file.ID.Hex()),
		zap.String("thumbnailPath", thumbnailPath),
		zap.Int("size", len(thumbnailData)))

	return thumbnailPath, nil
}

// getExtensionFromMimeType returns a file extension based on MIME type
func (s *onloadService) getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "text/plain":
		return ".txt"
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "application/json":
		return ".json"
	case "text/html":
		return ".html"
	case "application/zip":
		return ".zip"
	default:
		return ".dat" // Generic data file extension
	}
}
