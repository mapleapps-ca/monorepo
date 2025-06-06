// cloud/backend/internal/maplefile/usecase/collection/find_root_collections.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type FindRootCollectionsUseCase interface {
	Execute(ctx context.Context, ownerID primitive.ObjectID) ([]*dom_collection.Collection, error)
}

type findRootCollectionsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewFindRootCollectionsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) FindRootCollectionsUseCase {
	logger = logger.Named("FindRootCollectionsUseCase")
	return &findRootCollectionsUseCaseImpl{config, logger, repo}
}

func (uc *findRootCollectionsUseCaseImpl) Execute(ctx context.Context, ownerID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if ownerID.IsZero() {
		e["owner_id"] = "Owner ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating find root collections",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Find root collections.
	//

	return uc.repo.FindRootCollections(ctx, ownerID)
}
