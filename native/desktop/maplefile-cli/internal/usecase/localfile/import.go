// internal/usecase/localfile/import.go
package localfile

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
)

// ImportFileInput defines the input for importing a file
type ImportFileInput struct {
	FilePath          string
	CollectionID      primitive.ObjectID
	EncryptedFileID   string
	EncryptedMetadata string
	DecryptedName     string
	DecryptedMimeType string
	EncryptedFileKey  keys.EncryptedFileKey
	EncryptionVersion string
	GenerateThumbnail bool
	ThumbnailData     []byte
}

// ImportLocalFileUseCase defines the interface for importing files
type ImportLocalFileUseCase interface {
	Execute(ctx context.Context, input ImportFileInput) (*localfile.LocalFile, error)
}

// importLocalFileUseCase implements the ImportLocalFileUseCase interface
type importLocalFileUseCase struct {
	logger     *zap.Logger
	repository localfile.LocalFileRepository
}

// NewImportLocalFileUseCase creates a new use case for importing files
func NewImportLocalFileUseCase(
	logger *zap.Logger,
	repository localfile.LocalFileRepository,
) ImportLocalFileUseCase {
	return &importLocalFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute imports a file
func (uc *importLocalFileUseCase) Execute(
	ctx context.Context,
	input ImportFileInput,
) (*localfile.LocalFile, error) {
	// Validate inputs
	if input.FilePath == "" {
		return nil, errors.NewAppError("file path is required", nil)
	}

	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.EncryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	// Check if the file exists
	fileInfo, err := os.Stat(input.FilePath)
	if err != nil {
		return nil, errors.NewAppError("failed to access source file", err)
	}

	// Use filename from path if decrypted name is not provided
	decryptedName := input.DecryptedName
	if decryptedName == "" {
		decryptedName = filepath.Base(input.FilePath)
	}

	// Guess mime type from extension if not provided
	decryptedMimeType := input.DecryptedMimeType
	if decryptedMimeType == "" {
		// Simple extension-based mime type detection
		ext := filepath.Ext(input.FilePath)
		switch ext {
		case ".pdf":
			decryptedMimeType = "application/pdf"
		case ".jpg", ".jpeg":
			decryptedMimeType = "image/jpeg"
		case ".png":
			decryptedMimeType = "image/png"
		case ".txt":
			decryptedMimeType = "text/plain"
		case ".doc", ".docx":
			decryptedMimeType = "application/msword"
		case ".xls", ".xlsx":
			decryptedMimeType = "application/vnd.ms-excel"
		default:
			decryptedMimeType = "application/octet-stream"
		}
	}

	// Create a new local file
	file := &localfile.LocalFile{
		ID:                primitive.NewObjectID(),
		CollectionID:      input.CollectionID,
		EncryptedFileID:   input.EncryptedFileID,
		EncryptedMetadata: input.EncryptedMetadata,
		DecryptedName:     decryptedName,
		DecryptedMimeType: decryptedMimeType,
		OriginalSize:      fileInfo.Size(),
		EncryptedFileKey:  input.EncryptedFileKey,
		EncryptionVersion: input.EncryptionVersion,
		CreatedAt:         time.Now(),
		ModifiedAt:        time.Now(),
		IsModifiedLocally: true,
		SyncStatus:        localfile.SyncStatusLocalOnly,
	}

	// Save the file metadata
	if err := uc.repository.Create(ctx, file); err != nil {
		return nil, errors.NewAppError("failed to create file metadata", err)
	}

	// Import the file data
	if err := uc.repository.ImportFile(ctx, input.FilePath, file); err != nil {
		// If import fails, clean up the metadata
		_ = uc.repository.Delete(ctx, file.ID)
		return nil, errors.NewAppError("failed to import file data", err)
	}

	// Save thumbnail if provided
	if input.GenerateThumbnail && input.ThumbnailData != nil && len(input.ThumbnailData) > 0 {
		if err := uc.repository.SaveThumbnail(ctx, file, input.ThumbnailData); err != nil {
			uc.logger.Warn("Failed to save thumbnail for imported file",
				zap.String("fileID", file.ID.Hex()),
				zap.Error(err))
			// Continue even if thumbnail saving fails
		}
	}

	return file, nil
}
