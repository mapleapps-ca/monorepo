// monorepo/cloud/backend/internal/maplefile/usecase/collection/move_collection.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

// MoveCollectionRequest contains data needed to move a collection
type MoveCollectionRequest struct {
	CollectionID        gocql.UUID   `json:"collection_id"`
	NewParentID         gocql.UUID   `json:"new_parent_id"`
	UpdatedAncestors    []gocql.UUID `json:"updated_ancestors"`
	UpdatedPathSegments []string     `json:"updated_path_segments"`
}

type MoveCollectionUseCase interface {
	Execute(ctx context.Context, request MoveCollectionRequest) error
}

type moveCollectionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewMoveCollectionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) MoveCollectionUseCase {
	logger = logger.Named("MoveCollectionUseCase")
	return &moveCollectionUseCaseImpl{config, logger, repo}
}

func (uc *moveCollectionUseCaseImpl) Execute(ctx context.Context, request MoveCollectionRequest) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if request.CollectionID.String() == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if request.NewParentID.String() == "" {
		e["new_parent_id"] = "New parent ID is required"
	}
	if len(request.UpdatedAncestors) == 0 {
		e["updated_ancestors"] = "Updated ancestors are required"
	}
	if len(request.UpdatedPathSegments) == 0 {
		e["updated_path_segments"] = "Updated path segments are required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating move collection",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Move collection.
	//

	return uc.repo.MoveCollection(
		ctx,
		request.CollectionID,
		request.NewParentID,
		request.UpdatedAncestors,
		request.UpdatedPathSegments,
	)
}
