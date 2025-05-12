// cloud/backend/internal/papercloud/usecase/collection/list_shared_with_user.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type ListCollectionsSharedWithUserUseCase interface {
	Execute(ctx context.Context, userID string) ([]*dom_collection.Collection, error)
}

type listCollectionsSharedWithUserUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewListCollectionsSharedWithUserUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ListCollectionsSharedWithUserUseCase {
	return &listCollectionsSharedWithUserUseCaseImpl{config, logger, repo}
}

func (uc *listCollectionsSharedWithUserUseCaseImpl) Execute(ctx context.Context, userID string) ([]*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating list shared collections",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetCollectionsSharedWithUser(userID)
}
