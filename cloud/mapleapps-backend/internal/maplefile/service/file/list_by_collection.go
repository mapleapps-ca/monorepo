// cloud/backend/internal/maplefile/service/file/list_by_collection.go
package file

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	uc_filemetadata "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/filemetadata"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListFilesByCollectionRequestDTO struct {
	CollectionID primitive.ObjectID `json:"collection_id"`
}

type FilesResponseDTO struct {
	Files []*FileResponseDTO `json:"files"`
}

type ListFilesByCollectionService interface {
	Execute(ctx context.Context, req *ListFilesByCollectionRequestDTO) (*FilesResponseDTO, error)
}

type listFilesByCollectionServiceImpl struct {
	config                      *config.Configuration
	logger                      *zap.Logger
	collectionRepo              dom_collection.CollectionRepository
	getFilesByCollectionUseCase uc_filemetadata.GetFileMetadataByCollectionUseCase
}

func NewListFilesByCollectionService(
	config *config.Configuration,
	logger *zap.Logger,
	collectionRepo dom_collection.CollectionRepository,
	getFilesByCollectionUseCase uc_filemetadata.GetFileMetadataByCollectionUseCase,
) ListFilesByCollectionService {
	logger = logger.Named("ListFilesByCollectionService")
	return &listFilesByCollectionServiceImpl{
		config:                      config,
		logger:                      logger,
		collectionRepo:              collectionRepo,
		getFilesByCollectionUseCase: getFilesByCollectionUseCase,
	}
}

func (svc *listFilesByCollectionServiceImpl) Execute(ctx context.Context, req *ListFilesByCollectionRequestDTO) (*FilesResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Collection ID is required")
	}

	if req.CollectionID.IsZero() {
		svc.logger.Warn("Empty collection ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("collection_id", "Collection ID is required")
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
	// STEP 3: Check if user has access to the collection
	//
	hasAccess, err := svc.collectionRepo.CheckAccess(ctx, req.CollectionID, userID, dom_collection.CollectionPermissionReadOnly)
	if err != nil {
		svc.logger.Error("Failed to check collection access",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection access attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", req.CollectionID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to access this collection")
	}

	//
	// STEP 4: Get files by collection
	//
	files, err := svc.getFilesByCollectionUseCase.Execute(req.CollectionID)
	if err != nil {
		svc.logger.Error("Failed to get files by collection",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID))
		return nil, err
	}

	//
	// STEP 5: Map domain models to response DTOs
	//
	response := &FilesResponseDTO{
		Files: make([]*FileResponseDTO, len(files)),
	}

	for i, file := range files {
		response.Files[i] = mapFileToDTO(file)
	}

	svc.logger.Debug("Found files by collection",
		zap.Int("count", len(files)),
		zap.Any("collection_id", req.CollectionID))

	return response, nil
}
