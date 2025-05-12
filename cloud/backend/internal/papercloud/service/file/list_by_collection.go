// cloud/backend/internal/papercloud/service/file/list_by_collection.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
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

func (svc *listFilesByCollectionServiceImpl) Execute(sessCtx context.Context, collectionID string) (*FilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if collectionID == "" {
		svc.logger.Warn("Empty collection ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required")
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
	collection, err := svc.collectionRepo.Get(collectionID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.String("collection_id", collectionID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.String("collection_id", collectionID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	// Check if user has access to this collection
	hasAccess, err := svc.collectionRepo.CheckAccess(
		collectionID,
		userID.Hex(),
		dom_collection.CollectionPermissionReadOnly,
	)
	if err != nil {
		svc.logger.Error("Failed checking collection access",
			zap.Any("error", err),
			zap.String("collection_id", collectionID),
			zap.String("user_id", userID.Hex()))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection access attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("collection_id", collectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this collection")
	}

	//
	// STEP 4: Get files from repository
	//
	files, err := svc.fileRepo.GetByCollection(collectionID)
	if err != nil {
		svc.logger.Error("Failed to get files by collection",
			zap.Any("error", err),
			zap.String("collection_id", collectionID))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	response := &FilesResponseDTO{
		Files: make([]*FileResponseDTO, len(files)),
	}

	for i, file := range files {
		fileDTO := &FileResponseDTO{
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
			EncryptedThumbnail:    file.EncryptedThumbnail,
			CreatedAt:             file.CreatedAt,
			ModifiedAt:            file.ModifiedAt,
		}

		response.Files[i] = fileDTO
	}

	svc.logger.Debug("Retrieved files by collection",
		zap.Int("count", len(files)),
		zap.String("collection_id", collectionID))

	return response, nil
}
