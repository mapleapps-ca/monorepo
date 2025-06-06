// cloud/backend/internal/maplefile/usecase/collection/update.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UpdateCollectionUseCase interface {
	Execute(ctx context.Context, collection *dom_collection.Collection) error
}

type updateCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewUpdateCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) UpdateCollectionUseCase {
	logger = logger.Named("UpdateCollectionUseCase")
	return &updateCollectionUseCaseImpl{config, logger, repo}
}

func (uc *updateCollectionUseCaseImpl) Execute(ctx context.Context, collection *dom_collection.Collection) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collection == nil {
		e["collection"] = "Collection is required"
	} else {
		if collection.ID.String() == "" {
			e["id"] = "Collection ID is required"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection update",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Update in database.
	//

	return uc.repo.Update(ctx, collection)
}
