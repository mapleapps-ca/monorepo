// cloud/backend/internal/maplefile/service/file/get.go
package file

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
	s3storage "github.com/mapleapps-ca/monorepo/cloud/backend/pkg/storage/object/s3"
)

type GetFileService interface {
	Execute(ctx context.Context, fileID primitive.ObjectID) (*FileResponseDTO, error)
}

type getFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
	s3Storage      s3storage.S3ObjectStorage
}

func NewGetFileService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
	s3Storage s3storage.S3ObjectStorage,
) GetFileService {
	return &getFileServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
		s3Storage:      s3Storage,
	}
}

func (svc *getFileServiceImpl) Execute(ctx context.Context, fileID primitive.ObjectID) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if fileID.IsZero() {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file from repository
	//
	file, err := svc.fileRepo.Get(fileID)
	if err != nil {
		svc.logger.Error("Failed to get file",
			zap.Any("error", err),
			zap.Any("file_id", fileID))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found",
			zap.Any("file_id", fileID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	//
	// STEP 4: Check if the user has access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(
		ctx,
		file.CollectionID,
		userID,
		dom_collection.CollectionPermissionReadOnly,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file access attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", fileID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this file")
	}

	//
	// STEP 5: Generate pre-signed download URL if file data exists
	//
	var downloadURL string
	var downloadExpiry string

	// Only generate download URL if the file has been uploaded and stored
	if file.FileObjectKey != "" {
		// Determine storage path for the file
		storagePath := fmt.Sprintf("users/%s/files/%s", file.OwnerID.Hex(), file.ID.Hex())

		// Generate a presigned URL valid for 5 minutes
		expiration := 5 * time.Minute
		url, err := svc.s3Storage.GetPresignedURL(ctx, storagePath, expiration)
		if err != nil {
			svc.logger.Error("Failed to generate presigned download URL",
				zap.Any("error", err),
				zap.Any("file_id", fileID),
				zap.String("storage_path", storagePath))
			// Don't fail the entire request, just log the error and continue without download URL
			svc.logger.Warn("Continuing without download URL due to presigned URL generation failure")
		} else {
			downloadURL = url
			downloadExpiry = time.Now().Add(expiration).Format(time.RFC3339)
		}
	}

	//
	// STEP 6: Map domain model to response DTO
	//
	response := &FileResponseDTO{
		ID:                 file.ID,
		CollectionID:       file.CollectionID,
		OwnerID:            file.OwnerID,
		FileObjectKey:      file.FileObjectKey,
		EncryptedFileSize:  file.EncryptedFileSize,
		EncryptedMetadata:  file.EncryptedMetadata,
		EncryptionVersion:  file.EncryptionVersion,
		EncryptedHash:      file.EncryptedHash,
		EncryptedFileKey:   file.EncryptedFileKey,
		ThumbnailObjectKey: file.ThumbnailObjectKey,
		DownloadURL:        downloadURL,
		DownloadExpiry:     downloadExpiry,
		CreatedAt:          file.CreatedAt,
		ModifiedAt:         file.ModifiedAt,
	}

	svc.logger.Debug("File retrieved successfully with download URL",
		zap.Any("file_id", fileID),
		zap.Bool("has_download_url", downloadURL != ""))

	return response, nil
}
