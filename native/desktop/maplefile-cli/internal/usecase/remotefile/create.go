// internal/usecase/remotefile/create.go
package remotefile

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// CreateRemoteFileInput defines the input for creating a remote file
type CreateRemoteFileInput struct {
	CollectionID          primitive.ObjectID
	EncryptedFileID       string
	EncryptedSize         int64
	EncryptedOriginalSize string
	EncryptedMetadata     string
	EncryptedFileKey      keys.EncryptedFileKey
	EncryptionVersion     string
	EncryptedHash         string
	FileData              []byte
}

// CreateRemoteFileUseCase defines the interface for creating a remote file
type CreateRemoteFileUseCase interface {
	Execute(ctx context.Context, input CreateRemoteFileInput) (*remotefile.RemoteFileResponse, error)
}

// createRemoteFileUseCase implements the CreateRemoteFileUseCase interface
type createRemoteFileUseCase struct {
	logger     *zap.Logger
	repository remotefile.RemoteFileRepository
}

// NewCreateRemoteFileUseCase creates a new use case for creating remote files
func NewCreateRemoteFileUseCase(
	logger *zap.Logger,
	repository remotefile.RemoteFileRepository,
) CreateRemoteFileUseCase {
	return &createRemoteFileUseCase{
		logger:     logger,
		repository: repository,
	}
}

// Execute creates a new remote file
func (uc *createRemoteFileUseCase) Execute(
	ctx context.Context,
	input CreateRemoteFileInput,
) (*remotefile.RemoteFileResponse, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	if input.EncryptedFileID == "" {
		return nil, errors.NewAppError("encrypted file ID is required", nil)
	}

	if input.EncryptedMetadata == "" {
		return nil, errors.NewAppError("encrypted metadata is required", nil)
	}

	if input.EncryptedFileKey.Ciphertext == nil || len(input.EncryptedFileKey.Ciphertext) == 0 {
		return nil, errors.NewAppError("encrypted file key is required", nil)
	}

	// Create a request for the repository
	request := &remotefile.RemoteCreateFileRequest{
		CollectionID:          input.CollectionID,
		EncryptedFileID:       input.EncryptedFileID,
		EncryptedSize:         input.EncryptedSize,
		EncryptedOriginalSize: input.EncryptedOriginalSize,
		EncryptedMetadata:     input.EncryptedMetadata,
		EncryptedFileKey:      input.EncryptedFileKey,
		EncryptionVersion:     input.EncryptionVersion,
		EncryptedHash:         input.EncryptedHash,
	}

	// Create the remote file
	response, err := uc.repository.Create(ctx, request)
	if err != nil {
		return nil, errors.NewAppError("failed to create remote file", err)
	}

	// If file data is provided, upload it
	if input.FileData != nil && len(input.FileData) > 0 {
		if err := uc.repository.UploadFile(ctx, response.ID, input.FileData); err != nil {
			// Log error but still return the created file response
			uc.logger.Error("Failed to upload file data",
				zap.String("fileID", response.ID.Hex()),
				zap.Error(err))
			// Don't return early, as the file was created successfully
		}
	}

	return response, nil
}
