// cloud/backend/internal/maplefile/service/file/create_many.go
package file

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CreateManyFilesRequestDTO struct {
	Files []*CreateFileRequestDTO `json:"files"`
}

type CreateManyFilesResponseDTO struct {
	Files []*FileResponseDTO `json:"files"`
}

type CreateManyFilesService interface {
	Execute(ctx context.Context, req *CreateManyFilesRequestDTO) (*CreateManyFilesResponseDTO, error)
}

type createManyFilesServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewCreateManyFilesService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) CreateManyFilesService {
	return &createManyFilesServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *createManyFilesServiceImpl) Execute(ctx context.Context, req *CreateManyFilesRequestDTO) (*CreateManyFilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil || len(req.Files) == 0 {
		svc.logger.Warn("Failed validation with nil request or empty files list")
		return nil, httperror.NewForBadRequestWithSingleField("files", "At least one file is required")
	}

	// Get user ID from context
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	// Validate all files and check for collection access rights
	checkedCollections := make(map[primitive.ObjectID]bool)
	domainFiles := make([]*dom_file.File, 0, len(req.Files))
	now := time.Now()

	for i, fileReq := range req.Files {
		// Basic validation
		e := make(map[string]string)
		if fileReq.CollectionID.IsZero() {
			e["collection_id"] = "Collection ID is required"
		}
		if fileReq.EncryptedFileKey.Ciphertext == nil || len(fileReq.EncryptedFileKey.Ciphertext) == 0 {
			e["encrypted_file_key"] = "Encrypted file key is required"
		}
		if fileReq.EncryptedFileKey.Nonce == nil || len(fileReq.EncryptedFileKey.Nonce) == 0 {
			e["encrypted_file_key"] = "Encrypted file key nonce is required"
		}
		if fileReq.EncryptedMetadata == "" {
			e["encrypted_metadata"] = "Encrypted metadata is required"
		}
		if fileReq.EncryptionVersion == "" {
			e["encryption_version"] = "Encryption version is required"
		}

		if len(e) != 0 {
			svc.logger.Warn("Failed validation for file in batch",
				zap.Int("index", i),
				zap.Any("error", e))
			return nil, httperror.NewForBadRequestWithSingleField("files", "Invalid file data at index "+string(i))
		}

		// Check collection access (only once per collection)
		if _, checked := checkedCollections[fileReq.CollectionID]; !checked {
			// Verify collection exists
			collection, err := svc.collectionRepo.Get(ctx, fileReq.CollectionID)
			if err != nil {
				svc.logger.Error("Failed to get collection",
					zap.Any("error", err),
					zap.Any("collection_id", fileReq.CollectionID))
				return nil, err
			}

			if collection == nil {
				svc.logger.Debug("Collection not found",
					zap.Any("collection_id", fileReq.CollectionID))
				return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found: "+fileReq.CollectionID.Hex())
			}

			// Check if user has write access to this collection
			hasAccess, err := svc.collectionRepo.CheckAccess(
				ctx,
				fileReq.CollectionID,
				userID,
				dom_collection.CollectionPermissionReadWrite,
			)
			if err != nil {
				svc.logger.Error("Failed checking collection access",
					zap.Any("error", err),
					zap.Any("collection_id", fileReq.CollectionID),
					zap.Any("user_id", userID))
				return nil, err
			}

			if !hasAccess {
				svc.logger.Warn("Unauthorized file upload attempt",
					zap.Any("user_id", userID),
					zap.Any("collection_id", fileReq.CollectionID))
				return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have write access to collection: "+fileReq.CollectionID.Hex())
			}

			// Mark collection as checked
			checkedCollections[fileReq.CollectionID] = true
		}

		// Create file domain object
		fileID := primitive.NewObjectID()

		file := &dom_file.File{
			ID:                 fileID,
			CollectionID:       fileReq.CollectionID,
			OwnerID:            userID,
			EncryptedFileSize:  fileReq.EncryptedFileSize,
			EncryptedMetadata:  fileReq.EncryptedMetadata,
			EncryptedFileKey:   fileReq.EncryptedFileKey,
			EncryptionVersion:  fileReq.EncryptionVersion,
			EncryptedHash:      fileReq.EncryptedHash,
			ThumbnailObjectKey: "", // Will be populated later if thumbnail is uploaded
			CreatedAt:          now,
			ModifiedAt:         now,
		}

		domainFiles = append(domainFiles, file)
	}

	//
	// STEP 3: Create files in repository
	//
	err := svc.fileRepo.CreateMany(domainFiles)
	if err != nil {
		svc.logger.Error("Failed to create files in batch",
			zap.Any("error", err),
			zap.Int("count", len(domainFiles)))
		return nil, err
	}

	//
	// STEP 4: Map domain models to response DTOs
	//
	response := &CreateManyFilesResponseDTO{
		Files: make([]*FileResponseDTO, len(domainFiles)),
	}

	for i, file := range domainFiles {
		response.Files[i] = &FileResponseDTO{
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
			CreatedAt:          file.CreatedAt,
			ModifiedAt:         file.ModifiedAt,
		}
	}

	svc.logger.Debug("Files created successfully",
		zap.Int("count", len(domainFiles)))

	return response, nil
}
