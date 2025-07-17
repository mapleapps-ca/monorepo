// monorepo/cloud/mapleapps-backend/internal/maplefile/usecase/collection/archive.go
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
	// STEP 2: Archive collection using repository method.
	//

	return uc.repo.Archive(ctx, id)
}
