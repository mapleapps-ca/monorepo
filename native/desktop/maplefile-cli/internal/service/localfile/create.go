// internal/service/localfile/create.go
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

// CreateInput represents the input for creating a local file
type CreateInput struct {
	CollectionID      string                `json:"collection_id"`
	EncryptedFileID   string                `json:"encrypted_file_id"`
	EncryptedMetadata string                `json:"encrypted_metadata"`
	DecryptedName     string                `json:"decrypted_name"`
	DecryptedMimeType string                `json:"decrypted_mime_type"`
	OriginalSize      int64                 `json:"original_size"`
	EncryptedFileKey  keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion string                `json:"encryption_version"`
	FileData          []byte                `json:"file_data"`
	ThumbnailData     []byte                `json:"thumbnail_data,omitempty"`
}

// CreateOutput represents the result of creating a local file
type CreateOutput struct {
	File *localfile.LocalFile `json:"file"`
}

// CreateService defines the interface for creating local files
type CreateService interface {
	Create(ctx context.Context, input CreateInput) (*CreateOutput, error)
}

// createService implements the CreateService interface
type createService struct {
	logger            *zap.Logger
	createFileUseCase uc.CreateLocalFileUseCase
}

// NewCreateService creates a new service for creating local files
func NewCreateService(
	logger *zap.Logger,
	createFileUseCase uc.CreateLocalFileUseCase,
) CreateService {
	return &createService{
		logger:            logger,
		createFileUseCase: createFileUseCase,
	}
}

// Create handles file creation
func (s *createService) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	// Validate inputs
	if input.CollectionID == "" {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.EncryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	if input.EncryptedMetadata == "" {
		return nil, errors.NewAppError("encrypted metadata is required", nil)
	}

	if len(input.FileData) == 0 {
		return nil, errors.NewAppError("file data is required", nil)
	}

	// Convert collection ID to ObjectID
	collectionID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Prepare use case input
	useCaseInput := uc.CreateLocalFileInput{
		EncryptedFileID:   input.EncryptedFileID,
		CollectionID:      collectionID,
		EncryptedMetadata: input.EncryptedMetadata,
		DecryptedName:     input.DecryptedName,
		DecryptedMimeType: input.DecryptedMimeType,
		EncryptedFileKey:  input.EncryptedFileKey,
		EncryptionVersion: input.EncryptionVersion,
		FileData:          input.FileData,
		CreateThumbnail:   len(input.ThumbnailData) > 0,
		ThumbnailData:     input.ThumbnailData,
	}

	// Call the use case
	file, err := s.createFileUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		return nil, err
	}

	return &CreateOutput{
		File: file,
	}, nil
}
