// cloud/mapleapps-backend/internal/maplefile/usecase/collection/count_collections.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// CountCollectionsResponse contains the collection counts for a user
type CountCollectionsResponse struct {
	OwnedCollections  int `json:"owned_collections"`
	SharedCollections int `json:"shared_collections"`
	TotalCollections  int `json:"total_collections"`
}

type CountUserCollectionsUseCase interface {
	Execute(ctx context.Context, userID gocql.UUID) (*CountCollectionsResponse, error)
}

type countUserCollectionsUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewCountUserCollectionsUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) CountUserCollectionsUseCase {
	logger = logger.Named("CountUserCollectionsUseCase")
	return &countUserCollectionsUseCaseImpl{config, logger, repo}
}

func (uc *countUserCollectionsUseCaseImpl) Execute(ctx context.Context, userID gocql.UUID) (*CountCollectionsResponse, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if userID.String() == "" {
		e["user_id"] = "User ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating count user collections",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Count collections.
	//

	ownedCollections, err := uc.repo.CountOwnedCollections(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count owned collections",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	sharedCollections, err := uc.repo.CountSharedCollections(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to count shared collections",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	response := &CountCollectionsResponse{
		OwnedCollections:  ownedCollections,
		SharedCollections: sharedCollections,
		TotalCollections:  ownedCollections + sharedCollections,
	}

	uc.logger.Debug("Successfully counted user collections",
		zap.String("user_id", userID.String()),
		zap.Int("owned_collections", ownedCollections),
		zap.Int("shared_collections", sharedCollections),
		zap.Int("total_collections", response.TotalCollections))

	return response, nil
}
