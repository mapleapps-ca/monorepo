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
	logger = logger.Named("OnloadService")
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
		s.logger.Error("❌ input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.FileID == "" {
		s.logger.Error("❌ file ID is required")
		return nil, errors.NewAppError("file ID is required", nil)
	}
	if input.UserPassword == "" {
		s.logger.Error("❌ user password is required for E2EE operations")
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
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
	// STEP 3: Get the file and validate it's cloud-only
	//
	s.logger.Debug("🔍 Getting file for onload operation",
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

	// Only work with cloud-only files
	if file.SyncStatus != dom_file.SyncStatusCloudOnly {
		s.logger.Error("❌ file is not cloud-only",
			zap.String("fileID", input.FileID),
			zap.Any("syncStatus", file.SyncStatus))
		return nil, errors.NewAppError(
			fmt.Sprintf("file is not cloud-only (current status: %v)", file.SyncStatus),
			nil)
	}

	//
	// STEP 4: Download and decrypt file using the download service
	//
	s.logger.Info("⬇️ Downloading and decrypting file from cloud",
		zap.String("fileID", input.FileID))

	urlDuration := 1 * time.Hour // Default duration for download URLs
	downloadResult, err := s.downloadService.DownloadAndDecryptFile(ctx, fileObjectID, input.UserPassword, urlDuration)
	if err != nil {
		s.logger.Error("❌ failed to download and decrypt file",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to download and decrypt file", err)
	}

	s.logger.Info("✅ Successfully downloaded and decrypted file",
		zap.String("fileID", input.FileID),
		zap.String("fileName", downloadResult.DecryptedMetadata.Name),
		zap.Int64("size", downloadResult.OriginalSize))

	//
	// STEP 5: Save decrypted file locally
	//
	decryptedPath, err := s.saveDecryptedFile(ctx, file, downloadResult.DecryptedData, downloadResult.DecryptedMetadata)
	if err != nil {
		s.logger.Error("❌ failed to save decrypted file",
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
			s.logger.Warn("⚠️ Failed to save thumbnail, continuing without it",
				zap.String("fileID", input.FileID),
				zap.Error(err))
		} else {
			s.logger.Debug("✅ Successfully saved thumbnail",
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
		s.logger.Error("❌ failed to update file sync status during onload",
			zap.String("fileID", input.FileID),
			zap.Error(err))
		return nil, errors.NewAppError("failed to update file sync status during onload", err)
	}

	s.logger.Info("✨ Successfully onloaded file",
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
func (s *onloadService) saveDecryptedFile(ctx context.Context, file *dom_file.File, decryptedData []byte, metadata *svc_filedownload.DecryptedFileMetadata) (string, error) {
	s.logger.Debug("💾 Saving decrypted file locally", zap.String("fileID", file.ID.Hex()))

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

	// Enhanced file extension determination
	fileExtension := s.determineFileExtension(metadata, file.MimeType)

	destFileName := file.ID.Hex() + fileExtension
	destFilePath := s.pathUtilsUseCase.Join(ctx, collectionDir, destFileName)

	// Write the decrypted file
	err = os.WriteFile(destFilePath, decryptedData, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}

	s.logger.Debug("✅ Successfully saved decrypted file",
		zap.String("fileID", file.ID.Hex()),
		zap.String("filePath", destFilePath),
		zap.String("extension", fileExtension),
		zap.Int("size", len(decryptedData)))

	return destFilePath, nil
}

// saveThumbnail saves the decrypted thumbnail to local storage
func (s *onloadService) saveThumbnail(ctx context.Context, file *dom_file.File, thumbnailData []byte, originalFileName string) (string, error) {
	s.logger.Debug("🖼️ Saving thumbnail locally", zap.String("fileID", file.ID.Hex()))

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

	s.logger.Debug("✅ Successfully saved thumbnail",
		zap.String("fileID", file.ID.Hex()),
		zap.String("thumbnailPath", thumbnailPath),
		zap.Int("size", len(thumbnailData)))

	return thumbnailPath, nil
}

// Enhanced file extension determination with multiple fallback strategies
func (s *onloadService) determineFileExtension(metadata *svc_filedownload.DecryptedFileMetadata, mimeType string) string {
	// Strategy 1: Use explicit file extension from metadata (preferred)
	if metadata != nil && metadata.FileExtension != "" {
		s.logger.Debug("Using file extension from metadata", zap.String("extension", metadata.FileExtension))
		return metadata.FileExtension
	}

	// Strategy 2: Extract from metadata filename
	if metadata != nil && metadata.Name != "" {
		if ext := filepath.Ext(metadata.Name); ext != "" {
			s.logger.Debug("Using file extension from metadata name", zap.String("extension", ext))
			return ext
		}
	}

	// Strategy 3: Use enhanced MIME type mapping
	if mimeType != "" {
		if ext := s.getExtensionFromMimeType(mimeType); ext != ".dat" {
			s.logger.Debug("Using file extension from MIME type",
				zap.String("mimeType", mimeType),
				zap.String("extension", ext))
			return ext
		}
	}

	// Strategy 4: Final fallback
	s.logger.Warn("No file extension could be determined, using .dat fallback",
		zap.String("metadataName", func() string {
			if metadata != nil {
				return metadata.Name
			}
			return ""
		}()),
		zap.String("mimeType", mimeType))
	return ".dat"
}

// Enhanced MIME type to extension mapping with more comprehensive coverage
func (s *onloadService) getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	// Text files
	case "text/plain":
		return ".txt"
	case "text/html":
		return ".html"
	case "text/css":
		return ".css"
	case "text/javascript":
		return ".js"
	case "text/csv":
		return ".csv"
	case "text/xml":
		return ".xml"
	case "text/markdown":
		return ".md"

	// Documents
	case "application/pdf":
		return ".pdf"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	case "application/vnd.ms-powerpoint":
		return ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return ".pptx"
	case "application/rtf":
		return ".rtf"

	// Images
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/bmp":
		return ".bmp"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/tiff":
		return ".tiff"
	case "image/x-icon":
		return ".ico"

	// Audio
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "audio/mp4":
		return ".m4a"
	case "audio/x-flac":
		return ".flac"

	// Video
	case "video/mp4":
		return ".mp4"
	case "video/mpeg":
		return ".mpeg"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/webm":
		return ".webm"

	// Archives
	case "application/zip":
		return ".zip"
	case "application/x-rar-compressed":
		return ".rar"
	case "application/x-tar":
		return ".tar"
	case "application/gzip":
		return ".gz"
	case "application/x-7z-compressed":
		return ".7z"

	// Data formats
	case "application/json":
		return ".json"
	case "application/xml":
		return ".xml"
	case "application/yaml":
		return ".yaml"
	case "application/x-yaml":
		return ".yml"

	// Programming languages
	case "text/x-python":
		return ".py"
	case "text/x-java-source":
		return ".java"
	case "text/x-c":
		return ".c"
	case "text/x-c++src":
		return ".cpp"
	case "text/x-csharp":
		return ".cs"
	case "text/x-go":
		return ".go"
	case "text/x-ruby":
		return ".rb"
	case "text/x-php":
		return ".php"

	default:
		return ".dat" // Generic data file extension
	}
}
