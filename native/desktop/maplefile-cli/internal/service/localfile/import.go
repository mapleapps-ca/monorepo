// internal/service/localfile/import.go
package localfile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localfile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/localfile"
)

// ImportInput represents the input for importing a file
type ImportInput struct {
	FilePath          string                `json:"file_path"`
	CollectionID      string                `json:"collection_id"`
	EncryptedFileID   string                `json:"encrypted_file_id"`
	EncryptedMetadata string                `json:"encrypted_metadata"`
	DecryptedName     string                `json:"decrypted_name,omitempty"`
	DecryptedMimeType string                `json:"decrypted_mime_type,omitempty"`
	EncryptedFileKey  keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion string                `json:"encryption_version"`
	ThumbnailData     []byte                `json:"thumbnail_data,omitempty"`
	StorageMode       string                `json:"storage_mode,omitempty"`
}

// ImportOutput represents the result of importing a file
type ImportOutput struct {
	File *localfile.LocalFile `json:"file"`
}

// ImportService defines the interface for importing files
type ImportService interface {
	Import(ctx context.Context, input ImportInput) (*ImportOutput, error)
}

// importService implements the ImportService interface
type importService struct {
	logger            *zap.Logger
	importFileUseCase uc.ImportLocalFileUseCase
}

// NewImportService creates a new service for importing files
func NewImportService(
	logger *zap.Logger,
	importFileUseCase uc.ImportLocalFileUseCase,
) ImportService {
	return &importService{
		logger:            logger,
		importFileUseCase: importFileUseCase,
	}
}

// Import imports a file into the system
func (s *importService) Import(ctx context.Context, input ImportInput) (*ImportOutput, error) {
	// Validate inputs
	if input.FilePath == "" {
		return nil, errors.NewAppError("file path is required", nil)
	}

	if input.CollectionID == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.EncryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	if input.EncryptedMetadata == "" {
		return nil, errors.NewAppError("encrypted metadata is required", nil)
	}

	// Convert collection ID to ObjectID
	collectionID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Validate storage mode
	if input.StorageMode != localfile.StorageModeEncryptedOnly &&
		input.StorageMode != localfile.StorageModeDecryptedOnly &&
		input.StorageMode != localfile.StorageModeHybrid {
		return nil, errors.NewAppError("invalid storage mode", nil)
	}

	// Prepare use case input
	useCaseInput := uc.ImportFileInput{
		FilePath:          input.FilePath,
		CollectionID:      collectionID,
		EncryptedFileID:   input.EncryptedFileID,
		EncryptedMetadata: input.EncryptedMetadata,
		DecryptedName:     input.DecryptedName,
		DecryptedMimeType: input.DecryptedMimeType,
		EncryptedFileKey:  input.EncryptedFileKey,
		EncryptionVersion: input.EncryptionVersion,
		GenerateThumbnail: len(input.ThumbnailData) > 0,
		ThumbnailData:     input.ThumbnailData,
		StorageMode:       input.StorageMode,
	}

	// Call the use case
	file, err := s.importFileUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		return nil, err
	}

	return &ImportOutput{
		File: file,
	}, nil
}
