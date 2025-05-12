// cloud/backend/internal/papercloud/service/collection/delete.go
package collection

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

type DeleteCollectionRequestDTO struct {
	ID string `json:"id"`
}

type DeleteCollectionResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteCollectionService interface {
	Execute(sessCtx context.Context, req *DeleteCollectionRequestDTO) (*DeleteCollectionResponseDTO, error)
}

type deleteCollectionServiceImpl struct {
	config   *config.Configuration
	logger   *zap.Logger
	repo     dom_collection.CollectionRepository
	fileRepo dom_file.FileRepository
}

func NewDeleteCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
	fileRepo dom_file.FileRepository,
) DeleteCollectionService {
	return &deleteCollectionServiceImpl{
		config:   config,
		logger:   logger,
		repo:     repo,
		fileRepo: fileRepo,
	}
}

func (svc *deleteCollectionServiceImpl) Execute(sessCtx context.Context, req *DeleteCollectionRequestDTO) (*DeleteCollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection ID is required")
	}

	if req.ID == "" {
		svc.logger.Warn("Empty collection ID")
		return nil, httperror.NewForBadRequestWithSingleField("id", "Collection ID is required")
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
	// STEP 3: Retrieve existing collection
	//
	collection, err := svc.repo.Get(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get collection",
			zap.Any("error", err),
			zap.String("collection_id", req.ID))
		return nil, err
	}

	if collection == nil {
		svc.logger.Debug("Collection not found",
			zap.String("collection_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 4: Check if user has rights to delete this collection
	//
	if collection.OwnerID != userID.Hex() {
		svc.logger.Warn("Unauthorized collection deletion attempt",
			zap.String("user_id", userID.Hex()),
			zap.String("collection_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "Only the collection owner can delete a collection")
	}

	//
	// STEP 5: Check for associated files that need to be deleted
	//
	files, err := svc.fileRepo.GetByCollection(req.ID)
	if err != nil {
		svc.logger.Error("Failed to fetch files in collection",
			zap.Any("error", err),
			zap.String("collection_id", req.ID))
		return nil, err
	}

	// Delete all files within the collection
	for _, file := range files {
		err = svc.fileRepo.Delete(file.ID)
		if err != nil {
			svc.logger.Error("Failed to delete file during collection deletion",
				zap.Any("error", err),
				zap.String("file_id", file.ID),
				zap.String("collection_id", req.ID))
			// Continue deleting other files even if one fails
		}
	}

	//
	// STEP 6: Delete the collection
	//
	err = svc.repo.Delete(req.ID)
	if err != nil {
		svc.logger.Error("Failed to delete collection",
			zap.Any("error", err),
			zap.String("collection_id", req.ID))
		return nil, err
	}

	svc.logger.Info("Collection deleted successfully",
		zap.String("collection_id", req.ID),
		zap.Int("files_deleted", len(files)))

	return &DeleteCollectionResponseDTO{
		Success: true,
		Message: "Collection and all associated files deleted successfully",
	}, nil
}
