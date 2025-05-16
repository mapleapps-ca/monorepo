// cloud/backend/internal/maplefile/service/collection/list_shared_with_user.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

type ListSharedCollectionsService interface {
	Execute(ctx context.Context) (*CollectionsResponseDTO, error)
}

type listSharedCollectionsServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewListSharedCollectionsService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ListSharedCollectionsService {
	return &listSharedCollectionsServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *listSharedCollectionsServiceImpl) Execute(ctx context.Context) (*CollectionsResponseDTO, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, nil
	}

	//
	// STEP 2: Get collections shared with the user
	//
	collections, err := svc.repo.GetCollectionsSharedWithUser(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get shared collections",
			zap.Any("error", err),
			zap.Any("user_id", userID))
		return nil, err
	}

	//
	// STEP 3: Map domain models to response DTOs
	//
	response := &CollectionsResponseDTO{
		Collections: make([]*CollectionResponseDTO, len(collections)),
	}

	for i, collection := range collections {
		response.Collections[i] = mapCollectionToDTO(collection)
	}

	svc.logger.Debug("Retrieved shared collections",
		zap.Int("count", len(collections)),
		zap.Any("user_id", userID))

	return response, nil
}
