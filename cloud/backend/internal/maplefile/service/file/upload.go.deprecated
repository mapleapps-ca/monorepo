// cloud/backend/internal/maplefile/service/file/upload.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type UploadFileRequestDTO struct {
	CollectionID           primitive.ObjectID    `json:"collection_id"`
	EncryptedMetadata      string                `json:"encrypted_metadata"`
	EncryptedFileKey       keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion      string                `json:"encryption_version"`
	EncryptedHash          string                `json:"encrypted_hash"`
	EncryptedData          []byte                `json:"encrypted_data"`
	EncryptedThumbnailData []byte                `json:"encrypted_thumbnail_data,omitempty"`
}

type FileResponseDTO struct {
	ID                            primitive.ObjectID    `json:"id"`
	CollectionID                  primitive.ObjectID    `json:"collection_id"`
	OwnerID                       primitive.ObjectID    `json:"owner_id"`
	EncryptedMetadata             string                `json:"encrypted_metadata"`
	EncryptedFileKey              keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion             string                `json:"encryption_version"`
	EncryptedHash                 string                `json:"encrypted_hash"`
	EncryptedFileSizeInBytes      int64                 `json:"encrypted_file_size_in_bytes"`
	EncryptedThumbnailSizeInBytes int64                 `json:"encrypted_thumbnail_size_in_bytes"`
	CreatedAt                     time.Time             `json:"created_at"`
	ModifiedAt                    time.Time             `json:"modified_at"`
}

type UploadFileResponseDTO struct {
	File    *FileResponseDTO `json:"file"`
	Success bool             `json:"success"`
	Message string           `json:"message"`
}

type UploadFileService interface {
	Execute(ctx context.Context, req *UploadFileRequestDTO) (*UploadFileResponseDTO, error)
}

type uploadFileServiceImpl struct {
	config                *config.Configuration
	logger                *zap.Logger
	collectionRepo        dom_collection.CollectionRepository
	storeDataUseCase      uc_fileobjectstorage.StoreEncryptedDataUseCase
	deleteDataUseCase     uc_fileobjectstorage.DeleteEncryptedDataUseCase
	createMetadataUseCase uc_filemetadata.CreateFileMetadataUseCase
}

func NewUploadFileService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	storeDataUseCase uc_fileobjectstorage.StoreEncryptedDataUseCase,
	deleteDataUseCase uc_fileobjectstorage.DeleteEncryptedDataUseCase,
	createMetadataUseCase uc_filemetadata.CreateFileMetadataUseCase,
) UploadFileService {
	return &uploadFileServiceImpl{
		config:                config,
		logger:                logger,
		collectionRepo:        collectionRepo,
		storeDataUseCase:      storeDataUseCase,
		deleteDataUseCase:     deleteDataUseCase,
		createMetadataUseCase: createMetadataUseCase,
	}
}

func (svc *uploadFileServiceImpl) Execute(ctx context.Context, req *UploadFileRequestDTO) (*UploadFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File upload details are required")
	}

	e := make(map[string]string)
	if req.CollectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if req.EncryptedMetadata == "" {
		e["encrypted_metadata"] = "Encrypted metadata is required"
	}
	if req.EncryptedFileKey.Ciphertext == nil || len(req.EncryptedFileKey.Ciphertext) == 0 {
		e["encrypted_file_key"] = "Encrypted file key is required"
	}
	if req.EncryptionVersion == "" {
		e["encryption_version"] = "Encryption version is required"
	}
	if req.EncryptedHash == "" {
		e["encrypted_hash"] = "Encrypted hash is required"
	}
	if req.EncryptedData == nil || len(req.EncryptedData) == 0 {
		e["encrypted_data"] = "Encrypted data is required"
	}

	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
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
	// STEP 3: Check if user has write access to the collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, req.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file upload attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to upload files to this collection")
	}

	//
	// STEP 4: Generate file ID and store encrypted data
	//
	fileID := primitive.NewObjectID()

	storagePath, err := svc.storeDataUseCase.Execute(userID.Hex(), fileID.Hex(), req.EncryptedData)
	if err != nil {
		svc.logger.Error("Failed to store encrypted file data",
			zap.Any("error", err),
			zap.Any("file_id", fileID))
		return nil, err
	}

	//
	// STEP 5: Store thumbnail if provided
	//
	var thumbnailStoragePath string
	var thumbnailSize int64

	if req.EncryptedThumbnailData != nil && len(req.EncryptedThumbnailData) > 0 {
		thumbnailStoragePath, err = svc.storeDataUseCase.Execute(userID.Hex(), fileID.Hex()+"_thumb", req.EncryptedThumbnailData)
		if err != nil {
			svc.logger.Warn("Failed to store thumbnail, continuing without it",
				zap.Any("error", err),
				zap.Any("file_id", fileID))
		} else {
			thumbnailSize = int64(len(req.EncryptedThumbnailData))
		}
	}

	//
	// STEP 6: Create file metadata
	//
	now := time.Now()
	file := &dom_file.File{
		ID:                            fileID,
		CollectionID:                  req.CollectionID,
		OwnerID:                       userID,
		EncryptedMetadata:             req.EncryptedMetadata,
		EncryptedFileKey:              req.EncryptedFileKey,
		EncryptionVersion:             req.EncryptionVersion,
		EncryptedHash:                 req.EncryptedHash,
		EncryptedFileObjectKey:        storagePath,
		EncryptedFileSizeInBytes:      int64(len(req.EncryptedData)),
		EncryptedThumbnailObjectKey:   thumbnailStoragePath,
		EncryptedThumbnailSizeInBytes: thumbnailSize,
		CreatedAt:                     now,
		ModifiedAt:                    now,
	}

	err = svc.createMetadataUseCase.Execute(file)
	if err != nil {
		// Clean up stored data on metadata creation failure
		cleanupErr := svc.deleteDataUseCase.Execute(storagePath)
		if cleanupErr != nil {
			svc.logger.Warn("Failed to cleanup stored data after metadata creation failure",
				zap.Any("cleanup_error", cleanupErr),
				zap.String("storage_path", storagePath))
		}
		if thumbnailStoragePath != "" {
			cleanupErr = svc.deleteDataUseCase.Execute(thumbnailStoragePath)
			if cleanupErr != nil {
				svc.logger.Warn("Failed to cleanup stored thumbnail after metadata creation failure",
					zap.Any("cleanup_error", cleanupErr),
					zap.String("thumbnail_storage_path", thumbnailStoragePath))
			}
		}

		svc.logger.Error("Failed to create file metadata",
			zap.Any("error", err),
			zap.Any("file_id", fileID))
		return nil, err
	}

	//
	// STEP 7: Map domain model to response DTO
	//
	response := &UploadFileResponseDTO{
		File:    mapFileToDTO(file),
		Success: true,
		Message: "File uploaded successfully",
	}

	svc.logger.Info("File uploaded successfully",
		zap.Any("file_id", fileID),
		zap.Any("collection_id", req.CollectionID),
		zap.Any("owner_id", userID),
		zap.Int64("file_size", file.EncryptedFileSizeInBytes),
		zap.Int64("thumbnail_size", thumbnailSize))

	return response, nil
}

// Helper function to map a File domain model to a FileResponseDTO
func mapFileToDTO(file *dom_file.File) *FileResponseDTO {
	return &FileResponseDTO{
		ID:                            file.ID,
		CollectionID:                  file.CollectionID,
		OwnerID:                       file.OwnerID,
		EncryptedMetadata:             file.EncryptedMetadata,
		EncryptedFileKey:              file.EncryptedFileKey,
		EncryptionVersion:             file.EncryptionVersion,
		EncryptedHash:                 file.EncryptedHash,
		EncryptedFileSizeInBytes:      file.EncryptedFileSizeInBytes,
		EncryptedThumbnailSizeInBytes: file.EncryptedThumbnailSizeInBytes,
		CreatedAt:                     file.CreatedAt,
		ModifiedAt:                    file.ModifiedAt,
	}
}
