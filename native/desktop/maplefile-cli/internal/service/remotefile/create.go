// internal/service/remotefile/create.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/remotefile"
)

// CreateInput represents the input for creating a remote file
type CreateInput struct {
	CollectionID          string                `json:"collection_id"`
	EncryptedFileID       string                `json:"encrypted_file_id"`
	EncryptedSize         int64                 `json:"encrypted_size"`
	EncryptedOriginalSize string                `json:"encrypted_original_size,omitempty"`
	EncryptedMetadata     string                `json:"encrypted_metadata"`
	EncryptedFileKey      keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion     string                `json:"encryption_version"`
	EncryptedHash         string                `json:"encrypted_hash,omitempty"`
	FileData              []byte                `json:"file_data,omitempty"`
}

// CreateOutput represents the result of creating a remote file
type CreateOutput struct {
	File *remotefile.RemoteFileResponse `json:"file"`
}

// CreateService defines the interface for creating remote files
type CreateService interface {
	Create(ctx context.Context, input CreateInput) (*CreateOutput, error)
}

// createService implements the CreateService interface
type createService struct {
	logger            *zap.Logger
	createFileUseCase uc.CreateRemoteFileUseCase
}

// NewCreateService creates a new service for creating remote files
func NewCreateService(
	logger *zap.Logger,
	createFileUseCase uc.CreateRemoteFileUseCase,
) CreateService {
	return &createService{
		logger:            logger,
		createFileUseCase: createFileUseCase,
	}
}

// Create creates a new remote file
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

	// Convert collection ID to ObjectID
	collectionID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Prepare use case input
	useCaseInput := uc.CreateRemoteFileInput{
		CollectionID:          collectionID,
		EncryptedFileID:       input.EncryptedFileID,
		EncryptedSize:         input.EncryptedSize,
		EncryptedOriginalSize: input.EncryptedOriginalSize,
		EncryptedMetadata:     input.EncryptedMetadata,
		EncryptedFileKey:      input.EncryptedFileKey,
		EncryptionVersion:     input.EncryptionVersion,
		EncryptedHash:         input.EncryptedHash,
		FileData:              input.FileData,
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
