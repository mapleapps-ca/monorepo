// cloud/backend/internal/maplefile/usecase/collection/remove_member.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type RemoveCollectionMemberUseCase interface {
	Execute(ctx context.Context, collectionID string, recipientID string) error
}

type removeCollectionMemberUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewRemoveCollectionMemberUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) RemoveCollectionMemberUseCase {
	return &removeCollectionMemberUseCaseImpl{config, logger, repo}
}

func (uc *removeCollectionMemberUseCaseImpl) Execute(ctx context.Context, collectionID string, recipientID string) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID == "" {
		e["collection_id"] = "Collection ID is required"
	}
	if recipientID == "" {
		e["recipient_id"] = "Recipient ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating remove collection member",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Remove member from collection.
	//

	return uc.repo.RemoveMember(collectionID, recipientID)
}
