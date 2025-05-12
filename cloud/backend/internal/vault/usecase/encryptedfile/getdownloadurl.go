// cloud/backend/internal/vault/usecase/encryptedfile/getdownloadurl.go
package encryptedfile

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

// GetEncryptedFileDownloadURLUseCase defines operations for generating a download URL
type GetEncryptedFileDownloadURLUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID, expiryDuration time.Duration) (string, error)
}

type getEncryptedFileDownloadURLUseCaseImpl struct {
	config     *config.Configuration
	logger     *zap.Logger
	repository domain.Repository
	s3Storage  s3.S3ObjectStorage
}

// NewGetEncryptedFileDownloadURLUseCase creates a new instance of the use case
func NewGetEncryptedFileDownloadURLUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repository domain.Repository,
	s3Storage s3.S3ObjectStorage,
) GetEncryptedFileDownloadURLUseCase {
	return &getEncryptedFileDownloadURLUseCaseImpl{
		config:     config,
		logger:     logger.With(zap.String("component", "get-encrypted-file-download-url-usecase")),
		repository: repository,
		s3Storage:  s3Storage,
	}
}

// Execute generates a pre-signed URL for direct download
func (uc *getEncryptedFileDownloadURLUseCaseImpl) Execute(
	ctx context.Context,
	id primitive.ObjectID,
	expiryDuration time.Duration,
) (string, error) {
	// Validate inputs
	if id.IsZero() {
		return "", httperror.NewForBadRequestWithSingleField("id", "File ID cannot be empty")
	}

	// Default expiry duration if not provided
	if expiryDuration <= 0 {
		expiryDuration = 15 * time.Minute // Default to 15 minutes
	}

	// Maximum expiry duration
	const maxExpiryDuration = 24 * time.Hour // 24 hours
	if expiryDuration > maxExpiryDuration {
		expiryDuration = maxExpiryDuration
	}

	// Get the file metadata
	file, err := uc.repository.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get file metadata for download URL",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get file metadata: %w", err)
	}

	if file == nil {
		return "", httperror.NewForBadRequestWithSingleField("id", "File not found")
	}

	// Generate the download URL
	url, err := uc.s3Storage.GetDownloadablePresignedURL(ctx, file.FileID, expiryDuration)
	if err != nil {
		uc.logger.Error("Failed to generate download URL",
			zap.String("id", id.Hex()),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	uc.logger.Debug("Successfully generated download URL",
		zap.String("id", id.Hex()),
		zap.String("userID", file.UserID.Hex()),
		zap.String("fileID", file.FileID),
		zap.Duration("expiry", expiryDuration),
	)

	return url, nil
}
