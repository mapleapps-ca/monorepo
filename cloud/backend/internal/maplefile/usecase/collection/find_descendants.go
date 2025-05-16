// cloud/backend/internal/maplefile/usecase/collection/find_descendants.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type FindDescendantsUseCase interface {
	Execute(ctx context.Context, collectionID primitive.ObjectID) ([]*dom_collection.Collection, error)
}

type findDescendantsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewFindDescendantsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) FindDescendantsUseCase {
	return &findDescendantsUseCaseImpl{config, logger, repo}
}

func (uc *findDescendantsUseCaseImpl) Execute(ctx context.Context, collectionID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating find descendants",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Find descendants.
	//

	return uc.repo.FindDescendants(ctx, collectionID)
}
