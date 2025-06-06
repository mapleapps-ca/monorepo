// cloud/backend/internal/maplefile/usecase/collection/get_hierarchy.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type GetCollectionHierarchyUseCase interface {
	Execute(ctx context.Context, rootID gocql.UUID) (*dom_collection.Collection, error)
}

type getCollectionHierarchyUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewGetCollectionHierarchyUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) GetCollectionHierarchyUseCase {
	logger = logger.Named("GetCollectionHierarchyUseCase")
	return &getCollectionHierarchyUseCaseImpl{config, logger, repo}
}

func (uc *getCollectionHierarchyUseCaseImpl) Execute(ctx context.Context, rootID gocql.UUID) (*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if rootID.IsZero() {
		e["root_id"] = "Root collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating get collection hierarchy",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get collection hierarchy.
	//

	hierarchy, err := uc.repo.GetFullHierarchy(ctx, rootID)
	if err != nil {
		return nil, err
	}

	if hierarchy == nil {
		uc.logger.Debug("Collection not found",
			zap.Any("id", rootID))
		return nil, httperror.NewForNotFoundWithSingleField("message", "Collection not found")
	}

	return hierarchy, nil
}
