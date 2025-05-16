// cloud/backend/internal/maplefile/usecase/collection/get.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type GetCollectionUseCase interface {
	Execute(ctx context.Context, id string) (*dom_collection.Collection, error)
}

type getCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetCollectionUseCase {
	return &getCollectionUseCaseImpl{config, logger, repo}
}

func (uc *getCollectionUseCaseImpl) Execute(ctx context.Context, id string) (*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id == "" {
		e["id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection retrieval",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	collection, err := uc.repo.Get(id)
	if err != nil {
		return nil, err
	}

	if collection == nil {
		uc.logger.Debug("Collection not found",
			zap.String("id", id))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	return collection, nil
}
