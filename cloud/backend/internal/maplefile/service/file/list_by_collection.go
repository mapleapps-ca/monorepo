// cloud/backend/internal/maplefile/service/file/list_by_collection.go
package file

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type FilesResponseDTO struct {
	Files []*FileResponseDTO `json:"files"`
}

type ListFilesByCollectionService interface {
	Execute(sessCtx context.Context, collectionID string) (*FilesResponseDTO, error)
}

type listFilesByCollectionServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewListFilesByCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) ListFilesByCollectionService {
	return &listFilesByCollectionServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *listFilesByCollectionServiceImpl) Execute(sessCtx context.Context, collectionIDStr string) (*FilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if collectionIDStr == "" {
		svc.logger.Warn("Empty collection ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required")
	}

	// Convert string ID to ObjectID
	collectionID, err := primitive.ObjectIDFromHex(collectionIDStr)
	if err != nil {
		svc.logger.Error("Invalid collection ID format",
			zap.String("collection_id", collectionIDStr),
			zap.Error(err))
		return nil, httperror.NewForBadRequestWithSingleField("collection_id",
			fmt.Sprintf("Invalid collection ID format: %v", err))
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
	// STEP 3: Check if the collection exists and user has access
	//
	collection, err := svc.collectionRepo.Get(sessCtx, collectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("collection_id", collectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	// Check if user has access to this collection
	hasAccess, err := svc.collectionRepo.CheckAccess(
		sessCtx,
		collectionID,
		userID,
		dom_collection.CollectionPermissionReadOnly,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection access attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", collectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this collection")
	}

	//
	// STEP 4: Get files from repository
	//
	files, err := svc.fileRepo.GetByCollection(collectionIDStr)
	if err != nil {
		svc.logger.Error("Failed to get files by collection",
			zap.Any("error", err),
			zap.Any("collection_id", collectionID))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	response := &FilesResponseDTO{
		Files: make([]*FileResponseDTO, len(files)),
	}

	for i, file := range files {
		response.Files[i] = &FileResponseDTO{
			ID:                    file.ID,
			CollectionID:          file.CollectionID,
			OwnerID:               file.OwnerID,
			EncryptedFileID:       file.EncryptedFileID,
			FileObjectKey:         file.FileObjectKey,
			FileSize:              file.FileSize,
			EncryptedOriginalSize: file.EncryptedOriginalSize,
			EncryptedMetadata:     file.EncryptedMetadata,
			EncryptionVersion:     file.EncryptionVersion,
			EncryptedHash:         file.EncryptedHash,
			ThumbnailObjectKey:    file.ThumbnailObjectKey,
			CreatedAt:             file.CreatedAt,
			ModifiedAt:            file.ModifiedAt,
		}
	}

	svc.logger.Debug("Retrieved files by collection",
		zap.Int("count", len(files)),
		zap.Any("collection_id", collectionID))

	return response, nil
}
