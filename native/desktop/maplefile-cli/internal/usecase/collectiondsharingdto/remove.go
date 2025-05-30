// native/desktop/maplefile-cli/internal/usecase/collectiondsharingdto/remove.go
package collectiondsharingdto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// RemoveMemberInput represents input for removing a member from a collection
type RemoveMemberInput struct {
	CollectionID          primitive.ObjectID `json:"collection_id"`
	RecipientEmail        string             `json:"recipient_email"`
	RemoveFromDescendants bool               `json:"remove_from_descendants"`
}

// RemoveMemberUseCase defines the interface for removing members from collections
type RemoveMemberUseCase interface {
	Execute(ctx context.Context, input *RemoveMemberInput) (*collectionsharingdto.RemoveMemberResponseDTO, error)
}

// removeMemberUseCase implements the RemoveMemberUseCase interface
type removeMemberUseCase struct {
	logger                     *zap.Logger
	sharingRepo                collectionsharingdto.CollectionSharingDTORepository
	getCollectionUseCase       uc_collectiondto.GetCollectionFromCloudUseCase
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase
	getUserByEmailUseCase      uc_user.GetByEmailUseCase
}

// NewRemoveMemberUseCase creates a new use case for removing members from collections
func NewRemoveMemberUseCase(
	logger *zap.Logger,
	sharingRepo collectionsharingdto.CollectionSharingDTORepository,
	getCollectionUseCase uc_collectiondto.GetCollectionFromCloudUseCase,
	getUserByIsLoggedInUseCase uc_user.GetByIsLoggedInUseCase,
	getUserByEmailUseCase uc_user.GetByEmailUseCase,
) RemoveMemberUseCase {
	logger = logger.Named("RemoveMemberUseCase")
	return &removeMemberUseCase{
		logger:                     logger,
		sharingRepo:                sharingRepo,
		getCollectionUseCase:       getCollectionUseCase,
		getUserByIsLoggedInUseCase: getUserByIsLoggedInUseCase,
		getUserByEmailUseCase:      getUserByEmailUseCase,
	}
}

// Execute removes a member from a collection
func (uc *removeMemberUseCase) Execute(ctx context.Context, input *RemoveMemberInput) (*collectionsharingdto.RemoveMemberResponseDTO, error) {
	// Validate inputs
	if input.CollectionID.IsZero() {
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		return nil, errors.NewAppError("recipient email is required", nil)
	}

	// Get the collection
	coll, err := uc.getCollectionUseCase.Execute(ctx, input.CollectionID)
	if err != nil {
		uc.logger.Error("❌ Failed to get collection", zap.Error(err))
		return nil, errors.NewAppError("failed to get collection", err)
	}
	if coll == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	// Get current user
	currentUser, err := uc.getUserByIsLoggedInUseCase.Execute(ctx)
	if err != nil {
		uc.logger.Error("❌ Failed to get current user", zap.Error(err))
		return nil, errors.NewAppError("failed to get current user", err)
	}
	if currentUser == nil {
		return nil, errors.NewAppError("user not authenticated", nil)
	}

	// Check if current user has permission to remove members
	canRemove := coll.OwnerID == currentUser.ID
	if !canRemove {
		// Check if user is an admin member
		for _, member := range coll.Members {
			if member.RecipientID == currentUser.ID && member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin {
				canRemove = true
				break
			}
		}
	}
	if !canRemove {
		return nil, errors.NewAppError("you don't have permission to remove members from this collection", nil)
	}

	// Get recipient user information
	recipientUser, err := uc.getUserByEmailUseCase.Execute(ctx, input.RecipientEmail)
	if err != nil {
		uc.logger.Error("❌ Failed to get recipient user", zap.String("email", input.RecipientEmail), zap.Error(err))
		return nil, errors.NewAppError("failed to get recipient user", err)
	}
	if recipientUser == nil {
		return nil, errors.NewAppError("recipient user not found", nil)
	}

	// Check if recipient is the owner (owners cannot be removed)
	if recipientUser.ID == coll.OwnerID {
		return nil, errors.NewAppError("cannot remove the collection owner", nil)
	}

	// Check if recipient actually has access
	hasAccess := false
	for _, member := range coll.Members {
		if member.RecipientID == recipientUser.ID {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		return nil, errors.NewAppError("recipient does not have access to this collection", nil)
	}

	// Create remove request
	removeRequest := &collectionsharingdto.RemoveMemberRequestDTO{
		CollectionID:          input.CollectionID,
		RecipientID:           recipientUser.ID,
		RemoveFromDescendants: input.RemoveFromDescendants,
	}

	// Execute remove operation via repository
	response, err := uc.sharingRepo.RemoveMemberInCloud(ctx, removeRequest)
	if err != nil {
		uc.logger.Error("❌ Failed to remove collection member", zap.Error(err))
		return nil, err
	}

	uc.logger.Info("✅ Successfully removed collection member",
		zap.String("collectionID", input.CollectionID.Hex()),
		zap.String("recipientEmail", input.RecipientEmail))

	return response, nil
}
