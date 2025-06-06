// cloud/backend/internal/maplefile/service/collection/list_by_user.go
package collection

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

type CollectionsResponseDTO struct {
	Collections []*CollectionResponseDTO `json:"collections"`
}

type ListUserCollectionsService interface {
	Execute(ctx context.Context) (*CollectionsResponseDTO, error)
}

type listUserCollectionsServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewListUserCollectionsService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ListUserCollectionsService {
	logger = logger.Named("ListUserCollectionsService")
	return &listUserCollectionsServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *listUserCollectionsServiceImpl) Execute(ctx context.Context) (*CollectionsResponseDTO, error) {
	//
	// STEP 1: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(primitive.ObjectID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, errors.New("user ID not found in context")
	}

	//
	// STEP 2: Get user's collections from repository
	//
	collections, err := svc.repo.GetAllByUserID(ctx, userID)
	if err != nil {
		svc.logger.Error("Failed to get user collections",
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

	svc.logger.Debug("Retrieved user collections",
		zap.Int("count", len(collections)),
		zap.Any("user_id", userID))

	return response, nil
}
