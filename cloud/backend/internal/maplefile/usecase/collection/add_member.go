// cloud/backend/internal/maplefile/usecase/collection/add_member.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type AddCollectionMemberUseCase interface {
	Execute(ctx context.Context, collectionID string, membership *dom_collection.CollectionMembership) error
}

type addCollectionMemberUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewAddCollectionMemberUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) AddCollectionMemberUseCase {
	return &addCollectionMemberUseCaseImpl{config, logger, repo}
}

func (uc *addCollectionMemberUseCaseImpl) Execute(ctx context.Context, collectionID string, membership *dom_collection.CollectionMembership) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if membership == nil {
		e["membership"] = "Membership details are required"
	} else {
		if membership.RecipientID == "" {
			e["recipient_id"] = "Recipient ID is required"
		}
		if membership.RecipientEmail == "" {
			e["recipient_email"] = "Recipient email is required"
		}
		if membership.GrantedByID == "" {
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
		uc.logger.Warn("Failed validating add collection member",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Add member to collection.
	//

	return uc.repo.AddMember(collectionID, membership)
}
