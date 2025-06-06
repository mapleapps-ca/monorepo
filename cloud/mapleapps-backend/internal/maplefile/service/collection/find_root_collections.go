// cloud/backend/internal/maplefile/service/collection/find_root_collections.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

type FindRootCollectionsService interface {
	Execute(ctx context.Context) (*CollectionsResponseDTO, error)
}

type findRootCollectionsServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewFindRootCollectionsService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) FindRootCollectionsService {
	logger = logger.Named("FindRootCollectionsService")
	return &findRootCollectionsServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *findRootCollectionsServiceImpl) Execute(ctx context.Context) (*CollectionsResponseDTO, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, nil
	}

	//
	// STEP 2: Find root collections for the user
	//
	collections, err := svc.repo.FindRootCollections(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to find root collections",
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

	svc.logger.Debug("Found root collections",
		zap.Int("count", len(collections)),
		zap.Any("user_id", userID))

	return response, nil
}
