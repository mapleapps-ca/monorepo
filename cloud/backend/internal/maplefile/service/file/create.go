// cloud/backend/internal/maplefile/service/file/create.go
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
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CreateFileRequestDTO struct {
	CollectionID       primitive.ObjectID    `json:"collection_id"`
	EncryptedFileID    string                `json:"encrypted_file_id"`
	EncryptedFileSize  int64                 `json:"encrypted_file_size"`
	EncryptedMetadata  string                `json:"encrypted_metadata"`
	EncryptedFileKey   keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion  string                `json:"encryption_version"`
	EncryptedHash      string                `json:"encrypted_hash"`
	EncryptedThumbnail string                `json:"encrypted_thumbnail,omitempty"`
}

type FileResponseDTO struct {
	ID                 primitive.ObjectID    `json:"id"`
	CollectionID       primitive.ObjectID    `json:"collection_id"`
	OwnerID            primitive.ObjectID    `json:"owner_id"`
	EncryptedFileID    string                `json:"encrypted_file_id"`
	FileObjectKey      string                `json:"file_object_key"`
	EncryptedFileSize  int64                 `json:"encrypted_file_size"`
	EncryptedMetadata  string                `json:"encrypted_metadata"`
	EncryptionVersion  string                `json:"encryption_version"`
	EncryptedHash      string                `json:"encrypted_hash"`
	EncryptedFileKey   keys.EncryptedFileKey `json:"encrypted_file_key,omitempty"`
	ThumbnailObjectKey string                `json:"thumbnail_object_key,omitempty"`
	DownloadURL        string                `json:"download_url,omitempty"`
	DownloadExpiry     string                `json:"download_expiry,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	ModifiedAt         time.Time             `json:"modified_at"`
}

type CreateFileService interface {
	Execute(ctx context.Context, req *CreateFileRequestDTO) (*FileResponseDTO, error)
}

type createFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewCreateFileService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) CreateFileService {
	return &createFileServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *createFileServiceImpl) Execute(ctx context.Context, req *CreateFileRequestDTO) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File details are required")
	}

	e := make(map[string]string)
	if req.CollectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if req.EncryptedFileKey.Ciphertext == nil || len(req.EncryptedFileKey.Ciphertext) == 0 {
		e["encrypted_file_key"] = "Encrypted file key is required"
	}
	if req.EncryptedFileKey.Nonce == nil || len(req.EncryptedFileKey.Nonce) == 0 {
		e["encrypted_file_key"] = "Encrypted file key nonce is required"
	}
	if req.EncryptedMetadata == "" {
		e["encrypted_metadata"] = "Encrypted metadata is required"
	}
	if req.EncryptionVersion == "" {
		e["encryption_version"] = "Encryption version is required"
	}
	// if req.EncryptedHash == "" {
	// 	e["encrypted_hash"] = "Encrypted hash is required"
	// }

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
	// STEP 3: Verify collection exists and user has access
	//
	collection, err := svc.collectionRepo.Get(ctx, req.CollectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	// Check if user has write access to this collection
	hasAccess, err := svc.collectionRepo.CheckAccess(
		ctx,
		req.CollectionID,
		userID,
		dom_collection.CollectionPermissionReadWrite,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file upload attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have write access to this collection")
	}

	//
	// STEP 4: Create file object
	//
	now := time.Now()
	fileID := primitive.NewObjectID()

	// Use the provided EncryptedFileID or generate one
	encryptedFileID := req.EncryptedFileID
	if encryptedFileID == "" {
		encryptedFileID = primitive.NewObjectID().Hex()
	}

	file := &dom_file.File{
		ID:                 fileID,
		CollectionID:       req.CollectionID,
		OwnerID:            userID,
		EncryptedFileID:    encryptedFileID,
		EncryptedFileSize:  req.EncryptedFileSize,
		EncryptedMetadata:  req.EncryptedMetadata,
		EncryptedFileKey:   req.EncryptedFileKey,
		EncryptionVersion:  req.EncryptionVersion,
		EncryptedHash:      req.EncryptedHash,
		ThumbnailObjectKey: "", // Will be populated later if thumbnail is uploaded
		CreatedAt:          now,
		ModifiedAt:         now,
	}

	//
	// STEP 5: Create file in repository
	//
	err = svc.fileRepo.Create(file)
	if err != nil {
		svc.logger.Error("Failed to create file",
			zap.Any("error", err),
			zap.Any("collection_id", file.CollectionID),
			zap.String("encrypted_file_id", file.EncryptedFileID))
		return nil, err
	}

	// // Defensive code
	// if file.FileObjectKey == "" {
	// 	err := errors.New("file object key is empty")
	// 	svc.logger.Error("Failed to create file",
	// 		zap.Any("error", err),
	// 		zap.Any("collection_id", file.CollectionID),
	// 		zap.String("encrypted_file_id", file.EncryptedFileID))
	// 	return nil, err
	// }

	//
	// STEP 6: Map domain model to response DTO
	//
	response := &FileResponseDTO{
		ID:                 file.ID,
		CollectionID:       file.CollectionID,
		OwnerID:            file.OwnerID,
		EncryptedFileID:    file.EncryptedFileID,
		FileObjectKey:      file.FileObjectKey,
		EncryptedFileSize:  file.EncryptedFileSize,
		EncryptedMetadata:  file.EncryptedMetadata,
		EncryptionVersion:  file.EncryptionVersion,
		EncryptedHash:      file.EncryptedHash,
		EncryptedFileKey:   file.EncryptedFileKey,
		ThumbnailObjectKey: file.ThumbnailObjectKey,
		CreatedAt:          file.CreatedAt,
		ModifiedAt:         file.ModifiedAt,
	}

	svc.logger.Debug("File created successfully",
		zap.Any("file_id", file.ID),
		zap.Any("collection_id", file.CollectionID))

	return response, nil
}
