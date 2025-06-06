// cloud/backend/internal/maplefile/usecase/collection/list_shared_with_user.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListCollectionsSharedWithUserUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error)
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
	logger = logger.Named("ListCollectionsSharedWithUserUseCase")
	return &listCollectionsSharedWithUserUseCaseImpl{config, logger, repo}
}

func (uc *listCollectionsSharedWithUserUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
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

	return uc.repo.GetCollectionsSharedWithUser(ctx, userID)
}
