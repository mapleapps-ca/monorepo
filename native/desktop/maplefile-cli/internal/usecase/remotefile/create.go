// internal/usecase/remotefile/create.go
package remotefile

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/keys"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/remotefile"
)

// CreateRemoteFileInput defines the input for creating a remote file
type CreateRemoteFileInput struct {
	LocalID               primitive.ObjectID
	CollectionID          primitive.ObjectID
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

// Execute creates a new remote file with complete upload flow and rollback on failure
func (uc *createRemoteFileUseCase) Execute(
	ctx context.Context,
	input CreateRemoteFileInput,
) (*remotefile.RemoteFileResponse, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.EncryptedMetadata == "" {
		return nil, errors.NewAppError("encrypted metadata is required", nil)
	}

	if input.EncryptedFileKey.Ciphertext == nil || len(input.EncryptedFileKey.Ciphertext) == 0 {
		return nil, errors.NewAppError("encrypted file key is required", nil)
	}

	uc.logger.Info("Creating remote file with data upload",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.Int("fileDataSize", len(input.FileData)))

	// Step 1: Create file metadata on backend
	request := &remotefile.RemoteCreateFileRequest{
		LocalID:               input.LocalID,
		CollectionID:          input.CollectionID,
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
		zap.String("remoteFileID", response.ID.Hex()))

	// Step 2: Upload file data using ID if provided
	if input.FileData != nil && len(input.FileData) > 0 {
		uc.logger.Info("Uploading file data to backend/S3 using encrypted file ID",
			zap.String("remoteFileID", response.ID.Hex()),
			zap.Int("dataSize", len(input.FileData)))

		if err := uc.repository.UploadFileByLocalID(ctx, input.LocalID, input.FileData); err != nil {
			uc.logger.Error("Failed to upload file data, rolling back file creation",
				zap.String("remoteFileID", response.ID.Hex()),
				zap.Error(err))

			// Rollback: Delete the created file metadata
			deleteErr := uc.repository.Delete(ctx, response.ID)
			if deleteErr != nil {
				uc.logger.Error("Failed to rollback file creation after upload failure",
					zap.String("remoteFileID", response.ID.Hex()),
					zap.Error(deleteErr))
				// Return original upload error, but mention rollback failure
				return nil, errors.NewAppError(
					fmt.Sprintf("file upload failed and rollback also failed: upload error: %v, rollback error: %v",
						err, deleteErr), nil)
			}

			uc.logger.Info("Successfully rolled back file creation after upload failure",
				zap.String("remoteFileID", response.ID.Hex()))

			// Return the original upload error
			return nil, errors.NewAppError("failed to upload file data, file creation rolled back", err)
		}

		uc.logger.Info("File data uploaded successfully to backend/S3",
			zap.String("remoteFileID", response.ID.Hex()),
			zap.Int("uploadedBytes", len(input.FileData)))
	} else {
		uc.logger.Debug("No file data provided, skipping upload",
			zap.String("remoteFileID", response.ID.Hex()))
	}

	return response, nil
}
