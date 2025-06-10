// cloud/mapleapps-backend/internal/maplefile/usecase/collection/restore.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type RestoreCollectionUseCase interface {
	Execute(ctx context.Context, id gocql.UUID) error
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

func (uc *restoreCollectionUseCaseImpl) Execute(ctx context.Context, id gocql.UUID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if id.String() == "" {
		e["id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating collection restoration",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Restore collection using repository method.
	//

	return uc.repo.Restore(ctx, id)
}
