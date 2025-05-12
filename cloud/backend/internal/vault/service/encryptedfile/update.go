// cloud/backend/internal/vault/service/encryptedfile/update.go
package encryptedfile

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/usecase/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

// UpdateEncryptedFileService defines operations for updating an encrypted file
type UpdateEncryptedFileService interface {
	Execute(ctx context.Context, id primitive.ObjectID, encryptedMetadata string, encryptedHash string, encryptedContent io.Reader) (*domain.EncryptedFile, error)
}

type updateEncryptedFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	s3Storage      s3.S3ObjectStorage
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase
	updateUseCase  encryptedfile.UpdateEncryptedFileUseCase
}

// NewUpdateEncryptedFileService creates a new instance of the service
func NewUpdateEncryptedFileService(
	config *config.Configuration,
	logger *zap.Logger,
	s3Storage s3.S3ObjectStorage,
	getByIDUseCase encryptedfile.GetEncryptedFileByIDUseCase,
	updateUseCase encryptedfile.UpdateEncryptedFileUseCase,
) UpdateEncryptedFileService {
	return &updateEncryptedFileServiceImpl{
		config:         config,
		logger:         logger.With(zap.String("component", "update-encrypted-file-service")),
		s3Storage:      s3Storage,
		getByIDUseCase: getByIDUseCase,
		updateUseCase:  updateUseCase,
	}
}

// Execute updates an encrypted file after verifying ownership
func (s *updateEncryptedFileServiceImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
	encryptedMetadata string,
	encryptedHash string,
	encryptedContent io.Reader,
) (*domain.EncryptedFile, error) {
	// First get the file to verify ownership
	file, err := s.getByIDUseCase.Execute(ctx, id)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Verify that the authenticated user has access to this file
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if ok && !userID.IsZero() && file.UserID != userID {
		s.logger.Warn("Unauthorized file update attempt",
			zap.String("file_id", id.Hex()),
			zap.String("file_owner", file.UserID.Hex()),
			zap.String("requester", userID.Hex()),
		)
		return nil, httperror.NewForForbiddenWithSingleField("message", "You do not have permission to update this file")
	}

	// If new content is provided, update it in S3 first
	if encryptedContent != nil {
		s.logger.Info("Updating file content in S3",
			zap.String("file_id", id.Hex()),
			zap.String("storage_path", file.StoragePath))

		// Read the content for uploading to S3
		content, err := ioutil.ReadAll(encryptedContent)
		if err != nil {
			s.logger.Error("Failed to read updated encrypted content",
				zap.Error(err),
				zap.String("file_id", id.Hex()))
			return nil, fmt.Errorf("failed to read updated encrypted content: %w", err)
		}

		// Upload the new content to S3 - always private for encrypted files
		err = s.s3Storage.UploadContentWithVisibility(ctx, file.StoragePath, content, false)
		if err != nil {
			s.logger.Error("Failed to upload updated encrypted content to S3",
				zap.Error(err),
				zap.String("file_id", id.Hex()),
				zap.String("storage_path", file.StoragePath))
			return nil, fmt.Errorf("failed to upload updated encrypted content: %w", err)
		}

		s.logger.Debug("Successfully uploaded updated encrypted content to S3",
			zap.String("file_id", id.Hex()),
			zap.String("storage_path", file.StoragePath),
			zap.Int("size", len(content)))
	}

	// Update the file metadata using the use case
	return s.updateUseCase.Execute(ctx, id, encryptedMetadata, encryptedHash, nil) // Pass nil for content since we already handled S3 upload
}
