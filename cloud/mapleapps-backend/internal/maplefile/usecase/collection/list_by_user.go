// monorepo/cloud/backend/internal/maplefile/usecase/collection/list_by_user.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type ListCollectionsByUserUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error)
}

type listCollectionsByUserUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewListCollectionsByUserUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) ListCollectionsByUserUseCase {
	logger = logger.Named("ListCollectionsByUserUseCase")
	return &listCollectionsByUserUseCaseImpl{config, logger, repo}
}

func (uc *listCollectionsByUserUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating list collections by user",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get from database.
	//

	return uc.repo.GetAllByUserID(ctx, userID)
}
