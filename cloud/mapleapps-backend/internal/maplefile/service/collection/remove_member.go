// cloud/backend/internal/maplefile/service/collection/remove_member.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/pkg/httperror"
)

type RemoveMemberRequestDTO struct {
	CollectionID          gocql.UUID `json:"collection_id"`
	RecipientID           gocql.UUID `json:"recipient_id"`
	RemoveFromDescendants bool       `json:"remove_from_descendants"`
}

type RemoveMemberResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RemoveMemberService interface {
	Execute(ctx context.Context, req *RemoveMemberRequestDTO) (*RemoveMemberResponseDTO, error)
}

type removeMemberServiceImpl struct {
	config *config.Configuration
	logger *zap.Logger
	repo   dom_collection.CollectionRepository
}

func NewRemoveMemberService(
	config *config.Configuration,
	logger *zap.Logger,
	repo dom_collection.CollectionRepository,
) RemoveMemberService {
	logger = logger.Named("RemoveMemberService")
	return &removeMemberServiceImpl{
		config: config,
		logger: logger,
		repo:   repo,
	}
}

func (svc *removeMemberServiceImpl) Execute(ctx context.Context, req *RemoveMemberRequestDTO) (*RemoveMemberResponseDTO, error) {
	//
	// STEP 1: Validation
	//
	if req == nil {
		svc.logger.Warn("Failed validation with nil request")
		return nil, httperror.NewForBadRequestWithSingleField("non_field_error", "Remove member details are required")
	}

	e := make(map[string]string)
	if req.CollectionID.IsZero() {
		e["collection_id"] = "Collection ID is required"
	}
	if req.RecipientID.IsZero() {
		e["recipient_id"] = "Recipient ID is required"
	}

	if len(e) != 0 {
		svc.logger.Warn("Failed validation",
			zap.Any("error", e))
		return nil, httperror.NewForBadRequest(&e)
	}

	//
	// STEP 2: Get user ID from context
	//
	userID, ok := ctx.Value(constants.SessionFederatedUserID).(gocql.UUID)
	if !ok {
		svc.logger.Error("Failed getting user ID from context")
		return nil, httperror.NewForInternalServerErrorWithSingleField("message", "Authentication context error")
	}

	//
	// STEP 3: Check if user has admin access to the collection
	//
	hasAccess, err := svc.repo.CheckAccess(ctx, req.CollectionID, userID, dom_collection.CollectionPermissionAdmin)
	if err != nil {
		svc.logger.Error("Failed to check access",
			zap.Any("error", err),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("user_id", userID))
		return nil, err
	}

	// Collection owners and admin members can remove members
	if !hasAccess {
		isOwner, _ := svc.repo.IsCollectionOwner(ctx, req.CollectionID, userID)

		if !isOwner {
			svc.logger.Warn("Unauthorized member removal attempt",
				zap.Any("user_id", userID),
				zap.Any("collection_id", req.CollectionID))
			return nil, httperror.NewForForbiddenWithSingleField("message", "You don't have permission to remove members from this collection")
		}
	}

	//
	// STEP 4: Remove the member
	//
	var err2 error

	if req.RemoveFromDescendants {
		err2 = svc.repo.RemoveMemberFromHierarchy(ctx, req.CollectionID, req.RecipientID)
	} else {
		err2 = svc.repo.RemoveMember(ctx, req.CollectionID, req.RecipientID)
	}

	if err2 != nil {
		svc.logger.Error("Failed to remove member",
			zap.Any("error", err2),
			zap.Any("collection_id", req.CollectionID),
			zap.Any("recipient_id", req.RecipientID),
			zap.Bool("remove_from_descendants", req.RemoveFromDescendants))
		return nil, err2
	}

	svc.logger.Info("Member removed successfully",
		zap.Any("collection_id", req.CollectionID),
		zap.Any("recipient_id", req.RecipientID),
		zap.Bool("removed_from_descendants", req.RemoveFromDescendants))

	return &RemoveMemberResponseDTO{
		Success: true,
		Message: "Member removed successfully",
	}, nil
}
