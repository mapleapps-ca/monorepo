// cloud/backend/internal/maplefile/service/file/delete_many.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteManyFilesRequestDTO struct {
	IDs []primitive.ObjectID `json:"ids"`
}

type DeleteManyFilesResponseDTO struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DeletedCount int    `json:"deleted_count"`
}

type DeleteManyFilesService interface {
	Execute(ctx context.Context, req *DeleteManyFilesRequestDTO) (*DeleteManyFilesResponseDTO, error)
}

type deleteManyFilesServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewDeleteManyFilesService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) DeleteManyFilesService {
	return &deleteManyFilesServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *deleteManyFilesServiceImpl) Execute(ctx context.Context, req *DeleteManyFilesRequestDTO) (*DeleteManyFilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil || len(req.IDs) == 0 {
		svc.logger.Warn("Failed validation with nil request or empty IDs list")
		return nil, httperror.NewForBadRequestWithSingleField("ids", "At least one file ID is required")
	}

	// Check for invalid ObjectIDs
	for _, id := range req.IDs {
		if id.IsZero() {
			svc.logger.Warn("Invalid file ID in batch deletion request")
			return nil, httperror.NewForBadRequestWithSingleField("ids", "All file IDs must be valid")
		}
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
	// STEP 3: Retrieve all files and check permissions
	//
	files, err := svc.fileRepo.GetMany(req.IDs)
	if err != nil {
		svc.logger.Error("Failed to get files",
			zap.Any("error", err),
			zap.Any("file_ids", req.IDs))
		return nil, err
	}

	if len(files) == 0 {
		svc.logger.Debug("No files found")
		return nil, httperror.NewForNotFoundWithSingleField("message", "No files found with the provided IDs")
	}

	// Check permissions for each file and collect authorized IDs for deletion
	authorizedIDs := make([]primitive.ObjectID, 0, len(files))
	collectionAccessCache := make(map[primitive.ObjectID]bool)

	for _, file := range files {
		// If user is the owner, they can delete the file
		if file.OwnerID == userID {
			authorizedIDs = append(authorizedIDs, file.ID)
			continue
		}

		// Check admin access to the collection (with caching for performance)
		hasAdminAccess, cached := collectionAccessCache[file.CollectionID]
		if !cached {
			hasAdminAccess, err = svc.collectionRepo.CheckAccess(
				ctx,
				file.CollectionID,
				userID,
				dom_collection.CollectionPermissionAdmin,
			)
			if err != nil {
				svc.logger.Error("Failed checking collection access",
					zap.Any("error", err),
					zap.Any("collection_id", file.CollectionID),
					zap.Any("user_id", userID))
				return nil, err
			}
			collectionAccessCache[file.CollectionID] = hasAdminAccess
		}

		if hasAdminAccess {
			authorizedIDs = append(authorizedIDs, file.ID)
		} else {
			svc.logger.Warn("Skipping unauthorized file deletion",
				zap.Any("user_id", userID),
				zap.Any("file_id", file.ID))
		}
	}

	// Check if we have any files authorized for deletion
	if len(authorizedIDs) == 0 {
		svc.logger.Warn("No authorized files to delete",
			zap.Any("user_id", userID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to delete any of the requested files")
	}

	//
	// STEP 4: Delete the authorized files
	//
	err = svc.fileRepo.DeleteMany(authorizedIDs)
	if err != nil {
		svc.logger.Error("Failed to delete files",
			zap.Any("error", err),
			zap.Any("file_ids", authorizedIDs))
		return nil, err
	}

	svc.logger.Info("Files deleted successfully",
		zap.Int("requested", len(req.IDs)),
		zap.Int("deleted", len(authorizedIDs)))

	return &DeleteManyFilesResponseDTO{
		Success:      true,
		Message:      "Files deleted successfully",
		DeletedCount: len(authorizedIDs),
	}, nil
}
