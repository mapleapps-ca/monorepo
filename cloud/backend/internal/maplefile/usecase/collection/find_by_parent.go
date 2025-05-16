// cloud/backend/internal/maplefile/usecase/collection/find_by_parent.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type FindCollectionsByParentUseCase interface {
	Execute(ctx context.Context, parentID primitive.ObjectID) ([]*dom_collection.Collection, error)
}

type findCollectionsByParentUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewFindCollectionsByParentUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) FindCollectionsByParentUseCase {
	return &findCollectionsByParentUseCaseImpl{config, logger, repo}
}

func (uc *findCollectionsByParentUseCaseImpl) Execute(ctx context.Context, parentID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if parentID.IsZero() {
		e["parent_id"] = "Parent ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating find collections by parent",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Find collections by parent.
	//

	return uc.repo.FindByParent(ctx, parentID)
}
