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
	CollectionID          string                `json:"collection_id"`
	FileID                string                `json:"file_id"`
	EncryptedSize         int64                 `json:"encrypted_size"`
	EncryptedOriginalSize string                `json:"encrypted_original_size"`
	EncryptedMetadata     string                `json:"encrypted_metadata"`
	EncryptedFileKey      keys.EncryptedFileKey `json:"encrypted_file_key"`
	EncryptionVersion     string                `json:"encryption_version"`
	EncryptedHash         string                `json:"encrypted_hash"`
	EncryptedThumbnail    string                `json:"encrypted_thumbnail,omitempty"`
}

type FileResponseDTO struct {
	ID                    string                `json:"id"`
	CollectionID          string                `json:"collection_id"`
	OwnerID               string                `json:"owner_id"`
	FileID                string                `json:"file_id"`
	StoragePath           string                `json:"storage_path"`
	EncryptedSize         int64                 `json:"encrypted_size"`
	EncryptedOriginalSize string                `json:"encrypted_original_size"`
	EncryptedMetadata     string                `json:"encrypted_metadata"`
	EncryptionVersion     string                `json:"encryption_version"`
	EncryptedHash         string                `json:"encrypted_hash"`
	EncryptedFileKey      keys.EncryptedFileKey `json:"encrypted_file_key,omitempty"`
	EncryptedThumbnail    string                `json:"encrypted_thumbnail,omitempty"`
	CreatedAt             time.Time             `json:"created_at"`
	ModifiedAt            time.Time             `json:"modified_at"`
}

type CreateFileService interface {
	Execute(sessCtx context.Context, req *CreateFileRequestDTO) (*FileResponseDTO, error)
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

func (svc *createFileServiceImpl) Execute(sessCtx context.Context, req *CreateFileRequestDTO) (*FileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File details are required")
	}

	e := make(map[string]string)
	if req.CollectionID == "" {
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
	userID, ok := sessCtx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Verify collection exists and user has access
	//
	collection, err := svc.collectionRepo.Get(req.CollectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.String("collection_id", req.CollectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.String("collection_id", req.CollectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	// Check if user has write access to this collection
	hasAccess, err := svc.collectionRepo.CheckAccess(
		req.CollectionID,
		userID.Hex(),
		dom_collection.CollectionPermissionReadWrite,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.String("collection_id", req.CollectionID),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized file upload attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have write access to this collection")
	}

	//
	// STEP 4: Create file object
	//
	now := time.Now()
	storagePath := generateStoragePath(req.CollectionID, req.FileID)

	file := &dom_file.File{
		ID:                    generateFileID(),
		CollectionID:          req.CollectionID,
		OwnerID:               userID.Hex(),
		FileID:                req.FileID,
		StoragePath:           storagePath,
		EncryptedSize:         req.EncryptedSize,
		EncryptedOriginalSize: req.EncryptedOriginalSize,
		EncryptedMetadata:     req.EncryptedMetadata,
		EncryptedFileKey:      req.EncryptedFileKey,
		EncryptionVersion:     req.EncryptionVersion,
		// EncryptedHash:         req.EncryptedHash,
		EncryptedThumbnail: req.EncryptedThumbnail,
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
			zap.String("collection_id", file.CollectionID),
			zap.String("file_id", file.FileID))
		return nil, err
	}

	//
	// STEP 6: Map domain model to response DTO
	//
	response := &FileResponseDTO{
		ID:                    file.ID,
		CollectionID:          file.CollectionID,
		OwnerID:               file.OwnerID,
		FileID:                file.FileID,
		StoragePath:           file.StoragePath,
		EncryptedSize:         file.EncryptedSize,
		EncryptedOriginalSize: file.EncryptedOriginalSize,
		EncryptedMetadata:     file.EncryptedMetadata,
		EncryptionVersion:     file.EncryptionVersion,
		EncryptedHash:         file.EncryptedHash,
		EncryptedFileKey:      file.EncryptedFileKey,
		EncryptedThumbnail:    file.EncryptedThumbnail,
		CreatedAt:             file.CreatedAt,
		ModifiedAt:            file.ModifiedAt,
	}

	svc.logger.Debug("File created successfully",
		zap.String("file_id", file.ID),
		zap.String("collection_id", file.CollectionID))

	return response, nil
}

// Helper function to generate a storage path for the file
func generateStoragePath(collectionID, fileID string) string {
	return "files/" + collectionID + "/" + fileID
}

// Helper function to generate a unique file ID
func generateFileID() string {
	return primitive.NewObjectID().Hex()
}
