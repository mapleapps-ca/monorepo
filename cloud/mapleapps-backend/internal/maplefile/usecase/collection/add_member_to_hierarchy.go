// cloud/backend/internal/maplefile/usecase/collection/add_member_to_hierarchy.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type AddMemberToHierarchyUseCase interface {
	Execute(ctx context.Context, rootID gocql.UUID, membership *dom_collection.CollectionMembership) error
}

type addMemberToHierarchyUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewAddMemberToHierarchyUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) AddMemberToHierarchyUseCase {
	logger = logger.Named("AddMemberToHierarchyUseCase")
	return &addMemberToHierarchyUseCaseImpl{config, logger, repo}
}

func (uc *addMemberToHierarchyUseCaseImpl) Execute(ctx context.Context, rootID gocql.UUID, membership *dom_collection.CollectionMembership) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if rootID.IsZero() {
		e["root_id"] = "Root collection ID is required"
	}
	if membership == nil {
		e["membership"] = "Membership details are required"
	} else {
		if membership.RecipientID.IsZero() {
			e["recipient_id"] = "Recipient ID is required"
		}
		if membership.RecipientEmail == "" {
			e["recipient_email"] = "Recipient email is required"
		}
		if membership.GrantedByID.IsZero() {
			e["granted_by_id"] = "Granted by ID is required"
		}
		if len(membership.EncryptedCollectionKey) == 0 {
			e["encrypted_collection_key"] = "Encrypted collection key is required"
		}
		if membership.PermissionLevel == "" {
			// Default permission level will be set in the repository
		} else if membership.PermissionLevel != dom_collection.CollectionPermissionReadOnly &&
			membership.PermissionLevel != dom_collection.CollectionPermissionReadWrite &&
			membership.PermissionLevel != dom_collection.CollectionPermissionAdmin {
			e["permission_level"] = "Invalid permission level"
		}
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating add member to hierarchy",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Add member to collection hierarchy.
	//

	return uc.repo.AddMemberToHierarchy(ctx, rootID, membership)
}
