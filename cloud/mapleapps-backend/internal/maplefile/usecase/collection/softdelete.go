// cloud/backend/internal/maplefile/usecase/collection/softdelete.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type SoftDeleteCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

type softDeleteCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewSoftDeleteCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) SoftDeleteCollectionUseCase {
	logger = logger.Named("SoftDeleteCollectionUseCase")
	return &softDeleteCollectionUseCaseImpl{config, logger, repo}
}

func (uc *softDeleteCollectionUseCaseImpl) Execute(ctx context.Context, id primitive.ObjectID) error {
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

	return uc.repo.SoftDelete(ctx, id)
}
