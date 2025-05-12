// cloud/backend/internal/papercloud/usecase/collection/check_access.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type CheckCollectionAccessUseCase interface {
	Execute(ctx context.Context, collectionID string, userID string, requiredPermission string) (bool, error)
}

type checkCollectionAccessUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewCheckCollectionAccessUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) CheckCollectionAccessUseCase {
	return &checkCollectionAccessUseCaseImpl{config, logger, repo}
}

func (uc *checkCollectionAccessUseCaseImpl) Execute(ctx context.Context, collectionID string, userID string, requiredPermission string) (bool, error) {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if userID == "" {
		e["user_id"] = "User ID is required"
	}
	if requiredPermission == "" {
		// Default to read-only if not specified
		requiredPermission = dom_collection.CollectionPermissionReadOnly
	} else if requiredPermission != dom_collection.CollectionPermissionReadOnly &&
		requiredPermission != dom_collection.CollectionPermissionReadWrite &&
		requiredPermission != dom_collection.CollectionPermissionAdmin {
		e["required_permission"] = "Invalid permission level"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating check collection access",
			zap.Any("error", e))
		return false, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Check access.
	//

	return uc.repo.CheckAccess(collectionID, userID, requiredPermission)
}
