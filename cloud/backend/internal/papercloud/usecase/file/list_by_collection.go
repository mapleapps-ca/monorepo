// cloud/backend/internal/papercloud/usecase/file/list_by_collection.go
package file

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type ListFilesByCollectionUseCase interface {
	Execute(ctx context.Context, collectionID string) ([]*dom_file.File, error)
}

type listFilesByCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_file.FileRepository
}

func NewListFilesByCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_file.FileRepository,
) ListFilesByCollectionUseCase {
	return &listFilesByCollectionUseCaseImpl{config, logger, repo}
}

func (uc *listFilesByCollectionUseCaseImpl) Execute(ctx context.Context, collectionID string) ([]*dom_file.File, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating list files by collection request",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetByCollection(collectionID)
}
