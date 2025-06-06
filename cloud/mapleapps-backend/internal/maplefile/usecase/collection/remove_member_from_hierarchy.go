// cloud/backend/internal/maplefile/usecase/collection/remove_member_from_hierarchy.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type RemoveMemberFromHierarchyUseCase interface {
	Execute(ctx context.Context, rootID, recipientID gocql.UUID) error
}

type removeMemberFromHierarchyUseCaseImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewRemoveMemberFromHierarchyUseCase(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) RemoveMemberFromHierarchyUseCase {
	logger = logger.Named("RemoveMemberFromHierarchyUseCase")
	return &removeMemberFromHierarchyUseCaseImpl{config, logger, repo}
}

func (uc *removeMemberFromHierarchyUseCaseImpl) Execute(ctx context.Context, rootID, recipientID gocql.UUID) error {
	//
	// STEP 1: Validation.
	//

	e := make(map[string]string)
	if rootID.String() == "" {
		e["root_id"] = "Root collection ID is required"
	}
	if recipientID.String() == "" {
		e["recipient_id"] = "Recipient ID is required"
	}
	if len(e) != 0 {
		uc.logger.Warn("Failed validating remove member from hierarchy",
			zap.Any("error", e))
		return httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Remove member from collection hierarchy.
	//

	return uc.repo.RemoveMemberFromHierarchy(ctx, rootID, recipientID)
}
