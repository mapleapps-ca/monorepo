// cloud/backend/internal/maplefile/service/file/get_upload_url.go
package file

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	s3storage "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

// FileUploadURLResponseDTO represents the response for a file upload URL request
type FileUploadURLResponseDTO struct {
	URL      string `json:"url"`
	FileID   string `json:"file_id"`
	ExpireAt string `json:"expire_at"`
}

// GetFileUploadURLService defines the interface for generating upload URLs
type GetFileUploadURLService interface {
	Execute(ctx context.Context, fileID, userID primitive.ObjectID) (*FileUploadURLResponseDTO, error)
}

type getFileUploadURLServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
	s3Storage      s3storage.S3ObjectStorage
}

// NewGetFileUploadURLService creates a new service for generating upload URLs
func NewGetFileUploadURLService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
	s3Storage s3storage.S3ObjectStorage,
) GetFileUploadURLService {
	return &getFileUploadURLServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
		s3Storage:      s3Storage,
	}
}

// Execute generates a pre-signed upload URL for a file
func (svc *getFileUploadURLServiceImpl) Execute(
	ctx context.Context,
	fileID, userID primitive.ObjectID,
) (*FileUploadURLResponseDTO, error) {
	// Get the file to ensure it exists and check access
	file, err := svc.fileRepo.Get(fileID)
	if err != nil {
		svc.logger.Error("Failed to get file", zap.Any("file_id", fileID), zap.Error(err))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found", zap.Any("file_id", fileID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	// Check if user has write access to the file's collection
	hasAccess, err := svc.collectionRepo.CheckAccess(
		ctx,
		file.CollectionID,
		userID,
		dom_collection.CollectionPermissionReadWrite,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file upload URL request attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", fileID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to upload to this file")
	}

	// Determine storage path for the file
	storagePath := fmt.Sprintf("users/%s/files/%s", file.OwnerID.Hex(), file.EncryptedFileID)

	// Generate a presigned URL valid for 15 minutes
	expiration := 15 * time.Minute
	url, err := svc.s3Storage.GetPresignedURL(ctx, storagePath, expiration)
	if err != nil {
		svc.logger.Error("Failed to generate presigned URL",
			zap.Any("error", err),
			zap.Any("file_id", fileID),
			zap.String("storage_path", storagePath))
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Failed to generate upload URL")
	}

	expireTime := time.Now().Add(expiration)

	return &FileUploadURLResponseDTO{
		URL:      url,
		FileID:   fileID.Hex(),
		ExpireAt: expireTime.Format(time.RFC3339),
	}, nil
}
