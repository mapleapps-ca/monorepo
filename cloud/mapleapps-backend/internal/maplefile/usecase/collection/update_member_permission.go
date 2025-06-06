// cloud/backend/internal/maplefile/usecase/collection/update_member_permission.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type UpdateMemberPermissionUseCase interface {
	Execute(ctx context.Context, collectionID, recipientID gocql.UUID, newPermission string) error
}

type updateMemberPermissionUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewUpdateMemberPermissionUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) UpdateMemberPermissionUseCase {
	logger = logger.Named("UpdateMemberPermissionUseCase")
	return &updateMemberPermissionUseCaseImpl{config, logger, repo}
}

func (uc *updateMemberPermissionUseCaseImpl) Execute(ctx context.Context, collectionID, recipientID gocql.UUID, newPermission string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID.String() == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if recipientID.String() == "" {
		e["recipient_id"] = "Recipient ID is required"
	}
	if newPermission == "" {
		// Default permission level will be set in the repository
	} else if newPermission != dom_collection.CollectionPermissionReadOnly &&
		newPermission != dom_collection.CollectionPermissionReadWrite &&
		newPermission != dom_collection.CollectionPermissionAdmin {
		e["permission_level"] = "Invalid permission level"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating update member permission",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Update member permission.
	//

	return uc.repo.UpdateMemberPermission(ctx, collectionID, recipientID, newPermission)
}
