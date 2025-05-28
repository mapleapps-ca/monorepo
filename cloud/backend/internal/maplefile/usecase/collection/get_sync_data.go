// cloud/backend/internal/maplefile/usecase/collection/get_sync_data.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

type GetCollectionSyncDataUseCase interface {
	Execute(ctx context.Context, userID primitive.ObjectID, cursor *dom_collection.CollectionSyncCursor, limit int64) (*dom_collection.CollectionSyncResponse, error)
}

type getCollectionSyncDataUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetCollectionSyncDataUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetCollectionSyncDataUseCase {
	logger = logger.Named("GetCollectionSyncDataUseCase")
	return &getCollectionSyncDataUseCaseImpl{config, logger, repo}
}

func (uc *getCollectionSyncDataUseCaseImpl) Execute(ctx context.Context, userID primitive.ObjectID, cursor *dom_collection.CollectionSyncCursor, limit int64) (*dom_collection.CollectionSyncResponse, error) {
	//
	// STEP 1: Validation.
	//

	// (Skip)

	//
	// STEP 2: Get filtered collections from repository.
	//

	result, err := uc.repo.GetCollectionSyncData(ctx, userID, cursor, limit)
	if err != nil {
		uc.logger.Error("Failed to get filtered collections from repository",
			zap.Any("error", err),
			zap.Any("userID", userID),
			zap.Any("cursor", cursor),
			zap.Int64("limit", limit))
		return nil, err
	}

	return result, nil
}
