// cloud/backend/internal/maplefile/usecase/collection/remove_member.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/backend/pkg/httperror"
)

type RemoveCollectionMemberUseCase interface {
	Execute(ctx context.Context, collectionID, recipientID primitive.ObjectID) error
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
	logger = logger.Named("RemoveCollectionMemberUseCase")
	return &removeCollectionMemberUseCaseImpl{config, logger, repo}
}

func (uc *removeCollectionMemberUseCaseImpl) Execute(ctx context.Context, collectionID, recipientID primitive.ObjectID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if collectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if recipientID.IsZero() {
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

	return uc.repo.RemoveMember(ctx, collectionID, recipientID)
}
