// cloud/backend/internal/maplefile/service/file/get_presigned_download_url.go
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

type GetPresignedDownloadURLRequestDTO struct {
	FileID      primitive.ObjectID `json:"file_id"`
	URLDuration time.Duration      `json:"url_duration,omitempty"` // Optional, defaults to 1 hour
}

type GetPresignedDownloadURLResponseDTO struct {
	File                      *FileResponseDTO `json:"file"`
	PresignedDownloadURL      string           `json:"presigned_download_url"`
	PresignedThumbnailURL     string           `json:"presigned_thumbnail_url,omitempty"`
	DownloadURLExpirationTime time.Time        `json:"download_url_expiration_time"`
	Success                   bool             `json:"success"`
	Message                   string           `json:"message"`
}

type GetPresignedDownloadURLService interface {
	Execute(ctx context.Context, req *GetPresignedDownloadURLRequestDTO) (*GetPresignedDownloadURLResponseDTO, error)
}

type getPresignedDownloadURLServiceImpl struct {
	config                              *config.Configuration
	logger                              *zap.Logger
	collectionRepo                      dom_collection.CollectionRepository
	getMetadataUseCase                  uc_filemetadata.GetFileMetadataUseCase
	generatePresignedDownloadURLUseCase uc_fileobjectstorage.GeneratePresignedDownloadURLUseCase
}

func NewGetPresignedDownloadURLService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	generatePresignedDownloadURLUseCase uc_fileobjectstorage.GeneratePresignedDownloadURLUseCase,
) GetPresignedDownloadURLService {
	logger = logger.Named("GetPresignedDownloadURLService")
	return &getPresignedDownloadURLServiceImpl{
		config:                              config,
		logger:                              logger,
		collectionRepo:                      collectionRepo,
		getMetadataUseCase:                  getMetadataUseCase,
		generatePresignedDownloadURLUseCase: generatePresignedDownloadURLUseCase,
	}
}

func (svc *getPresignedDownloadURLServiceImpl) Execute(ctx context.Context, req *GetPresignedDownloadURLRequestDTO) (*GetPresignedDownloadURLResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("⚠️ Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Request details are required")
	}

	if req.FileID.IsZero() {
		svc.logger.Warn("⚠️ Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	// Set default URL duration if not provided
	if req.URLDuration == 0 {
		req.URLDuration = 1 * time.Hour
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("🔴 Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Get file metadata
	//
	file, err := svc.getMetadataUseCase.Execute(req.FileID)
	if err != nil {
		svc.logger.Error("🔴 Failed to get file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID))
		return nil, err
	}

	//
	// STEP 4: Check if user has read access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadOnly)
	if err != nil {
		svc.logger.Error("🔴 Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("⚠️ Unauthorized presigned download URL request",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to download this file")
	}

	//
	// STEP 5: Generate presigned download URLs
	//
	expirationTime := time.Now().Add(req.URLDuration)

	presignedDownloadURL, err := svc.generatePresignedDownloadURLUseCase.Execute(ctx, file.EncryptedFileObjectKey, req.URLDuration)
	if err != nil {
		svc.logger.Error("🔴 Failed to generate presigned download URL",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		return nil, err
	}

	// Generate thumbnail download URL if thumbnail path exists
	var presignedThumbnailURL string
	if file.EncryptedThumbnailObjectKey != "" {
		presignedThumbnailURL, err = svc.generatePresignedDownloadURLUseCase.Execute(ctx, file.EncryptedThumbnailObjectKey, req.URLDuration)
		if err != nil {
			svc.logger.Warn("⚠️ Failed to generate thumbnail presigned download URL, continuing without it",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID),
				zap.String("thumbnail_storage_path", file.EncryptedThumbnailObjectKey))
		}
	}

	//
	// STEP 6: Prepare response
	//
	response := &GetPresignedDownloadURLResponseDTO{
		File:                      mapFileToDTO(file),
		PresignedDownloadURL:      presignedDownloadURL,
		PresignedThumbnailURL:     presignedThumbnailURL,
		DownloadURLExpirationTime: expirationTime,
		Success:                   true,
		Message:                   "Presigned download URLs generated successfully",
	}

	svc.logger.Info("✅ Presigned download URLs generated successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("user_id", userID),
		zap.Time("url_expiration", expirationTime))

	return response, nil
}
