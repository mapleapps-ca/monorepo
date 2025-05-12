// cloud/backend/internal/vault/service/encryptedfile/create.go

package encryptedfile

import (
	"context"
	"fmt"
	"io"
	"io/ioutil" // For reading the entire content from a reader
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

// CreateEncryptedFileService defines the service for creating encrypted files
type CreateEncryptedFileService interface {
	Execute(ctx context.Context, userID primitive.ObjectID, fileID, encryptedMetadata, encryptedHash, encryptionVersion string, encryptedContent io.Reader) (*encryptedfile.EncryptedFile, error)
}

// createEncryptedFileService implements the CreateEncryptedFileService interface
type createEncryptedFileService struct {
	repo      encryptedfile.Repository
	s3Storage s3.S3ObjectStorage
	logger    *zap.Logger
}

// NewCreateEncryptedFileService creates a new service instance
func NewCreateEncryptedFileService(
	repo encryptedfile.Repository,
	s3Storage s3.S3ObjectStorage,
	logger *zap.Logger,
) CreateEncryptedFileService {
	return &createEncryptedFileService{
		repo:      repo,
		s3Storage: s3Storage,
		logger:    logger.With(zap.String("service", "create-encrypted-file")),
	}
}

// Execute handles the creation of a new encrypted file
func (s *createEncryptedFileService) Execute(
	ctx context.Context,
	userID primitive.ObjectID,
	fileID string,
	encryptedMetadata string,
	encryptedHash string,
	encryptionVersion string,
	encryptedContent io.Reader,
) (*encryptedfile.EncryptedFile, error) {
	// Create a new file entry
	file := &encryptedfile.EncryptedFile{
		ID:                primitive.NewObjectID(),
		UserID:            userID,
		FileID:            fileID,
		EncryptedMetadata: encryptedMetadata,
		EncryptedHash:     encryptedHash,
		EncryptionVersion: encryptionVersion,
		CreatedAt:         time.Now(),
		ModifiedAt:        time.Now(),
	}

	// Log the operation
	s.logger.Info("Creating new encrypted file",
		zap.String("id", file.ID.Hex()),
		zap.String("user_id", userID.Hex()),
		zap.String("file_id", fileID))

	// First, save metadata to MongoDB
	if err := s.repo.Create(ctx, file, nil); err != nil {
		s.logger.Error("Failed to create file metadata",
			zap.Error(err),
			zap.String("user_id", userID.Hex()),
			zap.String("file_id", fileID))
		return nil, fmt.Errorf("failed to create file metadata: %w", err)
	}

	// Then, upload the file content to S3
	if encryptedContent != nil {
		// Read the content into memory to calculate size and upload
		content, err := ioutil.ReadAll(encryptedContent)
		if err != nil {
			s.logger.Error("Failed to read encrypted content",
				zap.Error(err),
				zap.String("user_id", userID.Hex()),
				zap.String("file_id", fileID))
			return nil, fmt.Errorf("failed to read encrypted content: %w", err)
		}

		// Set the encrypted size in the file metadata
		file.EncryptedSize = int64(len(content))

		// Create the S3 storage path using the same format as in the Repository
		storagePath := fmt.Sprintf("%s/%s", userID.Hex(), fileID)

		// Upload to S3 - always use private visibility for encrypted files
		err = s.s3Storage.UploadContentWithVisibility(ctx, storagePath, content, false)
		if err != nil {
			s.logger.Error("Failed to upload encrypted content to S3",
				zap.Error(err),
				zap.String("user_id", userID.Hex()),
				zap.String("file_id", fileID),
				zap.String("storage_path", storagePath))
			return nil, fmt.Errorf("failed to upload encrypted content: %w", err)
		}

		s.logger.Debug("Successfully uploaded encrypted content to S3",
			zap.String("user_id", userID.Hex()),
			zap.String("file_id", fileID),
			zap.String("storage_path", storagePath),
			zap.Int64("size", file.EncryptedSize))

		// Update the file metadata with the encrypted size
		// Note: We're not updating the MongoDB record here, assuming EncryptedSize isn't critical
		// If it needs to be saved, we'd need another repo call to update
	}

	return file, nil
}
