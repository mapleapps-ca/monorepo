// cloud/backend/internal/maplefile/usecase/collection/archive.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ArchiveCollectionUseCase interface {
	Execute(ctx context.Context, id gocql.UUID) error
}

type archiveCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewArchiveCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ArchiveCollectionUseCase {
	logger = logger.Named("ArchiveCollectionUseCase")
	return &archiveCollectionUseCaseImpl{config, logger, repo}
}

func (uc *archiveCollectionUseCaseImpl) Execute(ctx context.Context, id gocql.UUID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection archival",
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

	err = dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateArchived)
	if err != nil {
		uc.logger.Warn("Invalid state transition for collection archival",
			zap.Any("collection_id", id),
			zap.String("current_state", collection.State),
			zap.String("target_state", dom_collection.CollectionStateArchived),
			zap.Error(err))
		return httperror.NewForBadRequestWithSingleField("state", err.Error())
	}

	//
	// STEP 4: Update state to archived.
	//

	collection.State = dom_collection.CollectionStateArchived
	return uc.repo.Update(ctx, collection)
}
