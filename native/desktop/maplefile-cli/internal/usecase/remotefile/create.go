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
	EncryptedFileSize     int64
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

// Execute creates a new remote file with complete upload flow
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

	uc.logger.Info("Creating remote file",
		zap.String("encryptedFileID", input.EncryptedFileID),
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.Int("fileDataSize", len(input.FileData)))

	// Step 1: Create file metadata on backend
	request := &remotefile.RemoteCreateFileRequest{
		CollectionID:          input.CollectionID,
		EncryptedFileID:       input.EncryptedFileID,
		EncryptedFileSize:     input.EncryptedFileSize,
		EncryptedOriginalSize: input.EncryptedOriginalSize,
		EncryptedMetadata:     input.EncryptedMetadata,
		EncryptedFileKey:      input.EncryptedFileKey,
		EncryptionVersion:     input.EncryptionVersion,
		EncryptedHash:         input.EncryptedHash,
	}

	// Create the remote file metadata
	response, err := uc.repository.Create(ctx, request)
	if err != nil {
		return nil, errors.NewAppError("failed to create remote file metadata", err)
	}

	uc.logger.Info("Remote file metadata created successfully",
		zap.String("remoteFileID", response.ID.Hex()),
		zap.String("encryptedFileID", input.EncryptedFileID))

	// Step 2: Upload file data to S3 if provided
	if input.FileData != nil && len(input.FileData) > 0 {
		uc.logger.Info("Uploading file data to S3",
			zap.String("remoteFileID", response.ID.Hex()),
			zap.Int("dataSize", len(input.FileData)))

		if err := uc.repository.UploadFile(ctx, response.ID, input.FileData); err != nil {
			uc.logger.Error("Failed to upload file data to S3",
				zap.String("remoteFileID", response.ID.Hex()),
				zap.Error(err))

			// Don't return error immediately - the file metadata was created successfully
			// But log the upload failure and continue
			uc.logger.Warn("File metadata created but data upload failed - file will need to be uploaded separately",
				zap.String("remoteFileID", response.ID.Hex()))

			// Return the response but indicate upload failure in logs
			// The caller can retry the upload later using the upload endpoint
		} else {
			uc.logger.Info("File data uploaded successfully to S3",
				zap.String("remoteFileID", response.ID.Hex()),
				zap.Int("uploadedBytes", len(input.FileData)))
		}
	} else {
		uc.logger.Debug("No file data provided, skipping upload",
			zap.String("remoteFileID", response.ID.Hex()))
	}

	return response, nil
}
