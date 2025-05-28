// cloud/backend/internal/maplefile/usecase/collection/delete.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type DeleteCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

type deleteCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewDeleteCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) DeleteCollectionUseCase {
	logger = logger.Named("DeleteCollectionUseCase")
	return &deleteCollectionUseCaseImpl{config, logger, repo}
}

func (uc *deleteCollectionUseCaseImpl) Execute(ctx context.Context, id primitive.ObjectID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.IsZero() {
		e["id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection deletion",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Delete from database.
	//

	return uc.repo.Delete(ctx, id)
}
