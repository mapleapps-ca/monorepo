// cloud/backend/internal/vault/usecase/encryptedfile/download.go
package encryptedfile

import (
	"context"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

// DownloadEncryptedFileUseCase defines operations for downloading encrypted file content
type DownloadEncryptedFileUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) (io.ReadCloser, error)
}

type downloadEncryptedFileUseCaseImpl struct {
	config     *config.Configuration
	logger     *zap.Logger
	repository domain.Repository
	s3Storage  s3.S3ObjectStorage
}

// NewDownloadEncryptedFileUseCase creates a new instance of the use case
func NewDownloadEncryptedFileUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repository domain.Repository,
	s3Storage s3.S3ObjectStorage,
) DownloadEncryptedFileUseCase {
	return &downloadEncryptedFileUseCaseImpl{
		config:     config,
		logger:     logger.With(zap.String("component", "download-encrypted-file-usecase")),
		repository: repository,
		s3Storage:  s3Storage,
	}
}

// Execute downloads the encrypted content of a file
func (uc *downloadEncryptedFileUseCaseImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
) (io.ReadCloser, error) {
	// Validate inputs
	if id.IsZero() {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID cannot be empty")
	}

	// Get the file metadata
	file, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get file metadata for download",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	if file == nil {
		return nil, httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Use the S3 storage to download the file
	content, err := uc.s3Storage.GetBinaryData(ctx, file.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to download encrypted file: %w", err)
	}

	uc.logger.Debug("Successfully downloaded encrypted file content",
		zap.String("id", id.Hex()),
		zap.String("userID", file.UserID.Hex()),
		zap.String("fileID", file.FileID),
	)

	return content, nil
}
