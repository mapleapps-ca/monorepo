// monorepo/cloud/backend/internal/maplefile/service/file/get_presigned_upload_url.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetPresignedUploadURLRequestDTO struct {
	FileID      gocql.UUID    `json:"file_id"`
	URLDuration time.Duration `json:"url_duration,omitempty"` // Optional, defaults to 1 hour
}

type GetPresignedUploadURLResponseDTO struct {
	File                    *FileResponseDTO `json:"file"`
	PresignedUploadURL      string           `json:"presigned_upload_url"`
	PresignedThumbnailURL   string           `json:"presigned_thumbnail_url,omitempty"`
	UploadURLExpirationTime time.Time        `json:"upload_url_expiration_time"`
	Success                 bool             `json:"success"`
	Message                 string           `json:"message"`
}

type GetPresignedUploadURLService interface {
	Execute(ctx context.Context, req *GetPresignedUploadURLRequestDTO) (*GetPresignedUploadURLResponseDTO, error)
}

type getPresignedUploadURLServiceImpl struct {
	config                            *config.Configuration
	logger                            *zap.Logger
	collectionRepo                    dom_collection.CollectionRepository
	getMetadataUseCase                uc_filemetadata.GetFileMetadataUseCase
	generatePresignedUploadURLUseCase uc_fileobjectstorage.GeneratePresignedUploadURLUseCase
}

func NewGetPresignedUploadURLService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getMetadataUseCase uc_filemetadata.GetFileMetadataUseCase,
	generatePresignedUploadURLUseCase uc_fileobjectstorage.GeneratePresignedUploadURLUseCase,
) GetPresignedUploadURLService {
	logger = logger.Named("GetPresignedUploadURLService")
	return &getPresignedUploadURLServiceImpl{
		config:                            config,
		logger:                            logger,
		collectionRepo:                    collectionRepo,
		getMetadataUseCase:                getMetadataUseCase,
		generatePresignedUploadURLUseCase: generatePresignedUploadURLUseCase,
	}
}

func (svc *getPresignedUploadURLServiceImpl) Execute(ctx context.Context, req *GetPresignedUploadURLRequestDTO) (*GetPresignedUploadURLResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Request details are required")
	}

	if req.FileID.String() == "" {
		svc.logger.Warn("Empty file ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("file_id", "File ID is required")
	}

	// Set default URL duration if not provided
	if req.URLDuration == 0 {
		req.URLDuration = 1 * time.Hour
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
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
	// STEP 4: Check if user has write access to the file's collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, file.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized presigned URL request",
			zap.Any("user_id", userID),
			zap.Any("file_id", req.FileID),
			zap.Any("collection_id", file.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to upload to this file")
	}

	//
	// STEP 5: Generate presigned upload URLs
	//
	expirationTime := time.Now().Add(req.URLDuration)

	presignedUploadURL, err := svc.generatePresignedUploadURLUseCase.Execute(ctx, file.EncryptedFileObjectKey, req.URLDuration)
	if err != nil {
		svc.logger.Error("Failed to generate presigned upload URL",
			zap.Any("error", err),
			zap.Any("file_id", req.FileID),
			zap.String("storage_path", file.EncryptedFileObjectKey))
		return nil, err
	}

	// Generate thumbnail upload URL if thumbnail path exists
	var presignedThumbnailURL string
	if file.EncryptedThumbnailObjectKey != "" {
		presignedThumbnailURL, err = svc.generatePresignedUploadURLUseCase.Execute(ctx, file.EncryptedThumbnailObjectKey, req.URLDuration)
		if err != nil {
			svc.logger.Warn("Failed to generate thumbnail presigned upload URL, continuing without it",
				zap.Any("error", err),
				zap.Any("file_id", req.FileID),
				zap.String("thumbnail_storage_path", file.EncryptedThumbnailObjectKey))
		}
	}

	//
	// STEP 6: Prepare response
	//
	response := &GetPresignedUploadURLResponseDTO{
		File:                    mapFileToDTO(file),
		PresignedUploadURL:      presignedUploadURL,
		PresignedThumbnailURL:   presignedThumbnailURL,
		UploadURLExpirationTime: expirationTime,
		Success:                 true,
		Message:                 "Presigned upload URLs generated successfully",
	}

	svc.logger.Info("Presigned upload URLs generated successfully",
		zap.Any("file_id", req.FileID),
		zap.Any("user_id", userID),
		zap.Time("url_expiration", expirationTime))

	return response, nil
}
