// native/desktop/maplefile-cli/internal/usecase/collectionsharingdto/remove.go
package collectionsharingdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	uc_collectiondto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc_user "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/user"
)

// RemoveMemberInput represents input for removing a member from a collection
type RemoveMemberInput struct {
	CollectionID          gocql.UUID `json:"collection_id"`
	RecipientEmail        string     `json:"recipient_email"`
	RemoveFromDescendants bool       `json:"remove_from_descendants"`
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

	// Debug: Let's see exactly what user we got and from where
	uc.logger.Debug("🔍 Current user authentication details",
		zap.String("retrievedUserID", currentUser.ID.Hex()),
		zap.String("retrievedUserEmail", currentUser.Email),
		zap.String("retrievedUserName", currentUser.Name))

	// Debug: Let's also check credentials from config
	// This will help us understand if the issue is in token/credential retrieval
	uc.logger.Debug("🔍 Expected vs Retrieved user mismatch investigation needed")

	// Debug logging for permission check
	uc.logger.Debug("🔍 Permission check details",
		zap.String("currentUserID", currentUser.ID.Hex()),
		zap.String("currentUserEmail", currentUser.Email),
		zap.String("collectionOwnerID", coll.OwnerID.Hex()),
		zap.Int("memberCount", len(coll.Members)))

	// Get recipient user information first to determine what we're trying to do
	recipientUser, err := uc.getUserByEmailUseCase.Execute(ctx, input.RecipientEmail)
	if err != nil {
		uc.logger.Error("❌ Failed to get recipient user", zap.String("email", input.RecipientEmail), zap.Error(err))
		return nil, errors.NewAppError("failed to get recipient user", err)
	}
	if recipientUser == nil {
		return nil, errors.NewAppError("recipient user not found", nil)
	}

	// Check if user is trying to remove themselves (self-removal is always allowed if they're a member)
	isSelfRemoval := recipientUser.ID == currentUser.ID
	uc.logger.Debug("🔍 Self-removal check",
		zap.Bool("isSelfRemoval", isSelfRemoval),
		zap.String("recipientUserID", recipientUser.ID.Hex()),
		zap.String("currentUserID", currentUser.ID.Hex()))

	// Check if current user has permission to remove members
	canRemove := coll.OwnerID == currentUser.ID
	uc.logger.Debug("🔍 Owner check",
		zap.Bool("isOwner", canRemove),
		zap.String("collectionOwnerID", coll.OwnerID.Hex()),
		zap.String("currentUserID", currentUser.ID.Hex()))

	if !canRemove {
		// Check if user is an admin member
		uc.logger.Debug("🔍 Checking admin membership")
		for i, member := range coll.Members {
			uc.logger.Debug("🔍 Checking member",
				zap.Int("memberIndex", i),
				zap.String("memberRecipientID", member.RecipientID.Hex()),
				zap.String("memberEmail", member.RecipientEmail),
				zap.String("memberPermissionLevel", member.PermissionLevel),
				zap.String("expectedAdminLevel", collectionsharingdto.CollectionDTOPermissionAdmin),
				zap.Bool("idMatch", member.RecipientID == currentUser.ID),
				zap.Bool("permissionMatch", member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin))

			if member.RecipientID == currentUser.ID && member.PermissionLevel == collectionsharingdto.CollectionDTOPermissionAdmin {
				uc.logger.Debug("✅ Found admin membership for current user")
				canRemove = true
				break
			}
		}
	}

	// If user can't remove others, check if this is self-removal
	if !canRemove && isSelfRemoval {
		// Allow self-removal if the user is actually a member of the collection
		for _, member := range coll.Members {
			if member.RecipientID == currentUser.ID {
				uc.logger.Debug("✅ Allowing self-removal - user is a member")
				canRemove = true
				break
			}
		}
	}

	if !canRemove {
		if isSelfRemoval {
			uc.logger.Error("🚫 Self-removal denied - user is not a member of this collection",
				zap.String("currentUserID", currentUser.ID.Hex()),
				zap.String("currentUserEmail", currentUser.Email))
			return nil, errors.NewAppError("you are not a member of this collection", nil)
		} else {
			uc.logger.Error("🚫 Permission denied for remove operation",
				zap.String("currentUserID", currentUser.ID.Hex()),
				zap.String("currentUserEmail", currentUser.Email),
				zap.String("collectionOwnerID", coll.OwnerID.Hex()),
				zap.Bool("isOwner", coll.OwnerID == currentUser.ID),
				zap.Int("totalMembers", len(coll.Members)))

			// Log all members for debugging
			for i, member := range coll.Members {
				uc.logger.Debug("🔍 Collection member details",
					zap.Int("index", i),
					zap.String("recipientID", member.RecipientID.Hex()),
					zap.String("recipientEmail", member.RecipientEmail),
					zap.String("permissionLevel", member.PermissionLevel))
			}

			return nil, errors.NewAppError("you don't have permission to remove members from this collection", nil)
		}
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
