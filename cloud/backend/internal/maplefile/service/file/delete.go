// cloud/backend/internal/maplefile/service/file/delete.go
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

type DeleteFileRequestDTO struct {
	ID string `json:"id"`
}

type DeleteFileResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type DeleteFileService interface {
	Execute(ctx context.Context, req *DeleteFileRequestDTO) (*DeleteFileResponseDTO, error)
}

type deleteFileServiceImpl struct {
	config         *config.Configuration
	logger         *zap.Logger
	fileRepo       dom_file.FileRepository
	collectionRepo dom_collection.CollectionRepository
}

func NewDeleteFileService(
	config *config.Configuration,
	logger *zap.Logger,
	fileRepo dom_file.FileRepository,
	collectionRepo dom_collection.CollectionRepository,
) DeleteFileService {
	return &deleteFileServiceImpl{
		config:         config,
		logger:         logger,
		fileRepo:       fileRepo,
		collectionRepo: collectionRepo,
	}
}

func (svc *deleteFileServiceImpl) Execute(ctx context.Context, req *DeleteFileRequestDTO) (*DeleteFileResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "File ID is required")
	}

	if req.ID == "" {
		svc.logger.Warn("Empty file ID")
		return nil, httperror.NewForBadRequestWithSingleField("id", "File ID is required")
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
	// STEP 3: Retrieve existing file
	//
	file, err := svc.fileRepo.Get(req.ID)
	if err != nil {
		svc.logger.Error("Failed to get file",
			zap.Any("error", err),
			zap.String("file_id", req.ID))
		return nil, err
	}

	if file == nil {
		svc.logger.Debug("File not found",
			zap.String("file_id", req.ID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "File not found")
	}

	//
	// STEP 4: Check if user has rights to delete this file
	//
	// Check if user is the owner or has admin rights on the collection
	isOwner := file.OwnerID == userID.Hex()

	hasAdminAccess := false
	if !isOwner {
		// Convert collection ID string to ObjectID
		collectionID, err := primitive.ObjectIDFromHex(file.CollectionID)
		if err != nil {
			svc.logger.Error("Failed to convert collection ID to ObjectID",
				zap.Any("error", err),
				zap.String("collection_id", file.CollectionID))
			return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Invalid collection ID format")
		}

		// Check if user has admin access to the collection
		hasAdminAccess, err = svc.collectionRepo.CheckAccess(
			ctx,
			collectionID,
			userID,
			dom_collection.CollectionPermissionAdmin,
		)
		if err != nil {
			svc.logger.Error("Failed checking collection access",
				zap.Any("error", err),
				zap.String("collection_id", file.CollectionID),
				zap.Any("user_id", userID))
			return nil, err
		}
	}

	if !isOwner && !hasAdminAccess {
		svc.logger.Warn("Unauthorized file deletion attempt",
			zap.Any("user_id", userID),
			zap.String("file_id", req.ID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to delete this file")
	}

	//
	// STEP 5: Delete the file
	//
	err = svc.fileRepo.Delete(req.ID)
	if err != nil {
		svc.logger.Error("Failed to delete file",
			zap.Any("error", err),
			zap.String("file_id", req.ID))
		return nil, err
	}

	svc.logger.Info("File deleted successfully",
		zap.String("file_id", req.ID),
		zap.String("collection_id", file.CollectionID))

	return &DeleteFileResponseDTO{
		Success: true,
		Message: "File deleted successfully",
	}, nil
}
