// cloud/backend/internal/maplefile/service/file/create_pending_file.go
package file

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/iam/domain/keys"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
	uc_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	uc_fileobjectstorage "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/fileobjectstorage"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type CreatePendingFileRequestDTO struct {
	ID                gocql.UUID            `json:"id"`
	CollectionID      gocql.UUID            `json:"collection_id"`
	EncryptedMetadata string                `json:"encrypted_metadata"`
	EncryptedFileKey  keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion string                `json:"encryption_version"`
	EncryptedHash     string                `json:"encrypted_hash"`
	// Optional: expected file size for validation (in bytes)
	ExpectedFileSizeInBytes int64 `json:"expected_file_size_in_bytes,omitempty"`
	// Optional: expected thumbnail size for validation (in bytes)
	ExpectedThumbnailSizeInBytes int64 `json:"expected_thumbnail_size_in_bytes,omitempty"`
}

type FileResponseDTO struct {
	ID                            gocql.UUID            `json:"id"`
	CollectionID                  gocql.UUID            `json:"collection_id"`
	OwnerID                       gocql.UUID            `json:"owner_id"`
	EncryptedMetadata             string                `json:"encrypted_metadata"`
	EncryptedFileKey              keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion             string                `json:"encryption_version"`
	EncryptedHash                 string                `json:"encrypted_hash"`
	EncryptedFileSizeInBytes      int64                 `json:"encrypted_file_size_in_bytes"`
	EncryptedThumbnailSizeInBytes int64                 `json:"encrypted_thumbnail_size_in_bytes"`
	CreatedAt                     time.Time             `json:"created_at"`
	ModifiedAt                    time.Time             `json:"modified_at"`
	Version                       uint64                `json:"version"`
	State                         string                `json:"state"`
	TombstoneVersion              uint64                `json:"tombstone_version"`
	TombstoneExpiry               time.Time             `json:"tombstone_expiry"`
}

type CreatePendingFileResponseDTO struct {
	File                    *FileResponseDTO `json:"file"`
	PresignedUploadURL      string           `json:"presigned_upload_url"`
	PresignedThumbnailURL   string           `json:"presigned_thumbnail_url,omitempty"`
	UploadURLExpirationTime time.Time        `json:"upload_url_expiration_time"`
	Success                 bool             `json:"success"`
	Message                 string           `json:"message"`
}

type CreatePendingFileService interface {
	Execute(ctx context.Context, req *CreatePendingFileRequestDTO) (*CreatePendingFileResponseDTO, error)
}

type createPendingFileServiceImpl struct {
	config                            *config.Configuration
	logger                            *zap.Logger
	checkCollectionAccessUseCase      uc_collection.CheckCollectionAccessUseCase
	checkFileExistsUseCase            uc_filemetadata.CheckFileExistsUseCase
	createMetadataUseCase             uc_filemetadata.CreateFileMetadataUseCase
	generatePresignedUploadURLUseCase uc_fileobjectstorage.GeneratePresignedUploadURLUseCase
}

func NewCreatePendingFileService(
	config *config.Configuration,
	logger *zap.Logger,
	checkCollectionAccessUseCase uc_collection.CheckCollectionAccessUseCase,
	checkFileExistsUseCase uc_filemetadata.CheckFileExistsUseCase,
	createMetadataUseCase uc_filemetadata.CreateFileMetadataUseCase,
	generatePresignedUploadURLUseCase uc_fileobjectstorage.GeneratePresignedUploadURLUseCase,
) CreatePendingFileService {
	logger = logger.Named("CreatePendingFileService")
	return &createPendingFileServiceImpl{
		config:                            config,
		logger:                            logger,
		checkCollectionAccessUseCase:      checkCollectionAccessUseCase,
		checkFileExistsUseCase:            checkFileExistsUseCase,
		createMetadataUseCase:             createMetadataUseCase,
		generatePresignedUploadURLUseCase: generatePresignedUploadURLUseCase,
	}
}

func (svc *createPendingFileServiceImpl) Execute(ctx context.Context, req *CreatePendingFileRequestDTO) (*CreatePendingFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("⚠️ Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File creation details are required")
	}

	e := make(map[string]string)
	if req.ID.String() == "" {
		e["id"] = "Client-side generated ID is required"
	}
	doesExist, err := svc.checkFileExistsUseCase.Execute(req.ID)
	if err != nil {
		e["id"] = fmt.Sprintf("Client-side generated ID causes error: %v", req.ID)
	}
	if doesExist {
		e["id"] = "Client-side generated ID already exists"
	}
	if req.CollectionID.String() == "" {
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

	if len(e) != 0 {
		svc.logger.Warn("⚠️ Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("❌ Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Check if user has write access to the collection
	//
	hasAccess, err := svc.checkCollectionAccessUseCase.Execute(ctx, req.CollectionID, userID, dom_collection.CollectionPermissionReadWrite)
	if err != nil {
		svc.logger.Error("❌ Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("⚠️ Unauthorized file creation attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to create files in this collection")
	}

	//
	// STEP 4: Generate storage paths.
	//
	storagePath := generateStoragePath(userID.String(), req.ID.String())
	thumbnailStoragePath := generateThumbnailStoragePath(userID.String(), req.ID.String())

	//
	// STEP 5: Generate presigned upload URLs
	//
	uploadURLDuration := 1 * time.Hour // URLs valid for 1 hour
	expirationTime := time.Now().Add(uploadURLDuration)

	presignedUploadURL, err := svc.generatePresignedUploadURLUseCase.Execute(ctx, storagePath, uploadURLDuration)
	if err != nil {
		svc.logger.Error("❌ Failed to generate presigned upload URL",
			zap.Any("error", err),
			zap.Any("file_id", req.ID),
			zap.String("storage_path", storagePath))
		return nil, err
	}

	// Generate thumbnail upload URL (optional)
	var presignedThumbnailURL string
	if req.ExpectedThumbnailSizeInBytes > 0 {
		presignedThumbnailURL, err = svc.generatePresignedUploadURLUseCase.Execute(ctx, thumbnailStoragePath, uploadURLDuration)
		if err != nil {
			svc.logger.Warn("⚠️ Failed to generate thumbnail presigned upload URL, continuing without it",
				zap.Any("error", err),
				zap.Any("file_id", req.ID),
				zap.String("thumbnail_storage_path", thumbnailStoragePath))
		}
	}

	//
	// STEP 6: Create pending file metadata record
	//
	now := time.Now()
	file := &dom_file.File{
		ID:                            req.ID,
		CollectionID:                  req.CollectionID,
		OwnerID:                       userID,
		EncryptedMetadata:             req.EncryptedMetadata,
		EncryptedFileKey:              req.EncryptedFileKey,
		EncryptionVersion:             req.EncryptionVersion,
		EncryptedHash:                 req.EncryptedHash,
		EncryptedFileObjectKey:        storagePath,
		EncryptedFileSizeInBytes:      req.ExpectedFileSizeInBytes, // Will be updated when upload completes
		EncryptedThumbnailObjectKey:   thumbnailStoragePath,
		EncryptedThumbnailSizeInBytes: req.ExpectedThumbnailSizeInBytes, // Will be updated when upload completes
		CreatedAt:                     now,
		CreatedByUserID:               userID,
		ModifiedAt:                    now,
		ModifiedByUserID:              userID,
		Version:                       1,                         // File creation always starts mutation version at 1.
		State:                         dom_file.FileStatePending, // File creation always starts state in a pending upload.
	}

	err = svc.createMetadataUseCase.Execute(file)
	if err != nil {
		svc.logger.Error("❌ Failed to create pending file metadata",
			zap.Any("error", err),
			zap.Any("file_id", req.ID))
		return nil, err
	}

	//
	// STEP 7: Prepare response
	//
	response := &CreatePendingFileResponseDTO{
		File:                    mapFileToDTO(file),
		PresignedUploadURL:      presignedUploadURL,
		PresignedThumbnailURL:   presignedThumbnailURL,
		UploadURLExpirationTime: expirationTime,
		Success:                 true,
		Message:                 "Pending file created successfully. Use the presigned URL to upload your file.",
	}

	svc.logger.Info("✅ Pending file created successfully",
		zap.Any("file_id", req.ID),
		zap.Any("collection_id", req.CollectionID),
		zap.Any("owner_id", userID),
		zap.String("storage_path", storagePath),
		zap.Time("url_expiration", expirationTime))

	return response, nil
}

// Helper function to generate consistent storage path
func generateStoragePath(ownerID, fileID string) string {
	return fmt.Sprintf("users/%s/files/%s", ownerID, fileID)
}

// Helper function to generate consistent thumbnail storage path
func generateThumbnailStoragePath(ownerID, fileID string) string {
	return fmt.Sprintf("users/%s/files/%s_thumb", ownerID, fileID)
}
