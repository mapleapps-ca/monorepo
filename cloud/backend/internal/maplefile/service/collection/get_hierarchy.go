// cloud/backend/internal/maplefile/service/collection/get_hierarchy.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetCollectionHierarchyService interface {
	Execute(ctx context.Context, rootID primitive.ObjectID) (*CollectionResponseDTO, error)
}

type getCollectionHierarchyServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetCollectionHierarchyService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetCollectionHierarchyService {
	logger = logger.Named("GetCollectionHierarchyService")
	return &getCollectionHierarchyServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *getCollectionHierarchyServiceImpl) Execute(ctx context.Context, rootID primitive.ObjectID) (*CollectionResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if rootID.IsZero() {
		svc.logger.Warn("Empty collection ID provided")
		return nil, httperror.NewForBadRequestWithSingleField("root_id", "Collection ID is required")
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
	// STEP 3: Check if user has access to the root collection
	//
	hasAccess, err := svc.repo.CheckAccess(ctx, rootID, userID, dom_collection.CollectionPermissionReadOnly)
	if err != nil {
		svc.logger.Error("Failed to check access",
			zap.Any("error", err),
			zap.Any("collection_id", rootID),
			zap.Any("user_id", userID))
		return nil, err
	}

	if !hasAccess {
		svc.logger.Warn("Unauthorized collection hierarchy access attempt",
			zap.Any("user_id", userID),
			zap.Any("collection_id", rootID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have access to this collection")
	}

	//
	// STEP 4: Get collection hierarchy
	//
	hierarchy, err := svc.repo.GetFullHierarchy(ctx, rootID)
	if err != nil {
		svc.logger.Error("Failed to get collection hierarchy",
			zap.Any("error", err),
			zap.Any("root_id", rootID))
		return nil, err
	}

	if hierarchy == nil {
		svc.logger.Debug("Collection not found",
			zap.Any("root_id", rootID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 5: Map domain model to response DTO
	//
	response := mapCollectionToDTO(hierarchy)

	svc.logger.Debug("Retrieved collection hierarchy",
		zap.Any("root_id", rootID))

	return response, nil
}
