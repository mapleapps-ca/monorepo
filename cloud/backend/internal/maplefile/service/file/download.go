// Deprecated
// cloud/backend/internal/maplefile/service/file/download.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DownloadFileRequestDTO struct {
	FileID          primitive.ObjectID `json:"file_id"`
	UsePresignedURL bool               `json:"use_presigned_url"`
	URLDuration     time.Duration      `json:"url_duration,omitempty"` // Optional, defaults to 1 hour
}

type DownloadFileResponseDTO struct {
	File          *FileResponseDTO `json:"file"`
	PresignedURL  string           `json:"presigned_url,omitempty"`
	EncryptedData []byte           `json:"encrypted_data,omitempty"`
	Success       bool             `json:"success"`
	Message       string           `json:"message"`
}

type DownloadFileService interface {
	Execute(ctx context.Context, req *DownloadFileRequestDTO) (*DownloadFileResponseDTO, error)
}

type downloadFileServiceImpl struct {
	config                              *config.Configuration
	logger                              *zap.Logger
	collectionRepo                      dom_collection.CollectionRepository
	getMetadataUseCase                  uc_filemetadata.GetFileMetadataUseCase
	getDataUseCase                      uc_fileobjectstorage.GetEncryptedDataUseCase
	generatePresignedDownloadURLUseCase uc_fileobjectstorage.GeneratePresignedDownloadURLUseCase
}

func NewDownloadFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	getDataUseCase uc_fileobjectstorage.GetEncryptedDataUseCase,
	generatePresignedDownloadURLUseCase uc_fileobjectstorage.GeneratePresignedDownloadURLUseCase,
) DownloadFileService {
	return &downloadFileServiceImpl{
		config:                              config,
		logger:                              logger,
		collectionRepo:                      collectionRepo,
		getMetadataUseCase:                  getMetadataUseCase,
		getDataUseCase:                      getDataUseCase,
		generatePresignedDownloadURLUseCase: generatePresignedDownloadURLUseCase,
	}
}

func (svc *downloadFileServiceImpl) Execute(ctx context.Context, req *DownloadFileRequestDTO) (*DownloadFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File download details are required")
	}

	if req.FileID.IsZero() {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	// Set default URL duration if not provided
	if req.UsePresignedURL && req.URLDuration == 0 {
		req.URLDuration = 1 * time.Hour
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
	// STEP 3: Get file metadata
	//
	file, err := svc.getMetadataUseCase.Execute(req.FileID)
	if err != nil {
		svc.logger.Error("Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 4: Check if user has access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadOnly)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file download attempt",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to download this file")
	}

	//
	// STEP 5: Get file data or presigned URL
	//
	response := &DownloadFileResponseDTO{
		File:    mapFileToDTO(file),
		Success: true,
		Message: "File retrieved successfully",
	}

	if req.UsePresignedURL {
		// Generate presigned URL
		url, err := svc.generatePresignedDownloadURLUseCase.Execute(ctx, file.EncryptedFileObjectKey, req.URLDuration)
		if err != nil {
			svc.logger.Error("Failed to generate presigned URL",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID))
			return nil, err
		}
		response.PresignedURL = url
	} else {
		// Get encrypted data directly
		data, err := svc.getDataUseCase.Execute(file.EncryptedFileObjectKey)
		if err != nil {
			svc.logger.Error("Failed to get encrypted file data",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID))
			return nil, err
		}
		response.EncryptedData = data
	}

	svc.logger.Debug("File download prepared successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("user_id", userID),
		zap.Bool("presigned_url", req.UsePresignedURL))

	return response, nil
}
