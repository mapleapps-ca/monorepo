// cloud/backend/internal/maplefile/usecase/collection/restore.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type RestoreCollectionUseCase interface {
	Execute(ctx context.Context, id primitive.ObjectID) error
}

type restoreCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewRestoreCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) RestoreCollectionUseCase {
	logger = logger.Named("RestoreCollectionUseCase")
	return &restoreCollectionUseCaseImpl{config, logger, repo}
}

func (uc *restoreCollectionUseCaseImpl) Execute(ctx context.Context, id primitive.ObjectID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.IsZero() {
		e["id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection restoration",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get current collection to validate state transition.
	//

	collection, err := uc.repo.GetWithAnyState(ctx, id)
	if err != nil {
		return err
	}
	if collection == nil {
		return httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	//
	// STEP 3: Validate state transition.
	//

	err = dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateActive)
	if err != nil {
		uc.logger.Warn("Invalid state transition for collection restoration",
			zap.Any("collection_id", id),
			zap.String("current_state", collection.State),
			zap.String("target_state", dom_collection.CollectionStateActive),
			zap.Error(err))
		return httperror.NewForBadRequestWithSingleField("state", err.Error())
	}

	//
	// STEP 4: Update state to active.
	//

	collection.State = dom_collection.CollectionStateActive
	return uc.repo.Update(ctx, collection)
}
