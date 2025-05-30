// internal/service/collectionsharingdto/sharing.go
package collectionsharingdto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectionsharingdto"
	uc_collectionsharingdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
)

// ShareCollectionInput represents input for sharing a collection at the service level
type ShareCollectionInput struct {
	CollectionID         string `json:"collection_id"`
	RecipientEmail       string `json:"recipient_email"`
	PermissionLevel      string `json:"permission_level"`
	ShareWithDescendants bool   `json:"share_with_descendants"`
}

// ShareCollectionOutput represents the output from sharing a collection
type ShareCollectionOutput struct {
	Success            bool   `json:"success"`
	Message            string `json:"message"`
	MembershipsCreated int    `json:"memberships_created"`
}

// RemoveMemberInput represents input for removing a member at the service level
type RemoveMemberInput struct {
	CollectionID          string `json:"collection_id"`
	RecipientEmail        string `json:"recipient_email"`
	RemoveFromDescendants bool   `json:"remove_from_descendants"`
}

// RemoveMemberOutput represents the output from removing a member
type RemoveMemberOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ListSharedCollectionsOutput represents the output from listing shared collections
type ListSharedCollectionsOutput struct {
	Collections []*collectiondto.CollectionDTO `json:"collections"`
	Count       int                            `json:"count"`
}

// SharingService defines the interface for collection sharing operations
type SharingService interface {
	ShareCollection(ctx context.Context, input *ShareCollectionInput, userPassword string) (*ShareCollectionOutput, error)
	RemoveMember(ctx context.Context, input *RemoveMemberInput) (*RemoveMemberOutput, error)
	ListSharedCollections(ctx context.Context) (*ListSharedCollectionsOutput, error)
	GetCollectionMembers(ctx context.Context, collectionID string) ([]*collectiondto.CollectionMembershipDTO, error)
}

// sharingService implements the SharingService interface
type sharingService struct {
	logger                        *zap.Logger
	shareCollectionUseCase        uc.ShareCollectionUseCase
	removeMemberUseCase           uc.RemoveMemberUseCase
	listSharedCollectionsUseCase  uc.ListSharedCollectionsUseCase
	getCollectionFromCloudUseCase uc_collectionsharingdto.GetCollectionFromCloudUseCase
}

// NewSharingService creates a new collection sharing service
func NewSharingService(
	logger *zap.Logger,
	shareCollectionUseCase uc.ShareCollectionUseCase,
	removeMemberUseCase uc.RemoveMemberUseCase,
	listSharedCollectionsUseCase uc.ListSharedCollectionsUseCase,
	getCollectionFromCloudUseCase uc_collectionsharingdto.GetCollectionFromCloudUseCase,
) SharingService {
	logger = logger.Named("CollectionSharingService")
	return &sharingService{
		logger:                        logger,
		shareCollectionUseCase:        shareCollectionUseCase,
		removeMemberUseCase:           removeMemberUseCase,
		listSharedCollectionsUseCase:  listSharedCollectionsUseCase,
		getCollectionFromCloudUseCase: getCollectionFromCloudUseCase,
	}
}

// ShareCollection shares a collection with another user
func (s *sharingService) ShareCollection(ctx context.Context, input *ShareCollectionInput, userPassword string) (*ShareCollectionOutput, error) {
	// Validate inputs
	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		s.logger.Error("❌ Recipient email is required")
		return nil, errors.NewAppError("recipient email is required", nil)
	}
	if input.PermissionLevel == "" {
		s.logger.Error("❌ Permission level is required")
		return nil, errors.NewAppError("permission level is required", nil)
	}
	if userPassword == "" {
		s.logger.Error("❌ User password is required for E2EE operations")
		return nil, errors.NewAppError("user password is required for E2EE operations", nil)
	}

	// Convert string ID to ObjectID
	collectionObjectID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		s.logger.Error("❌ Invalid collection ID format", zap.String("id", input.CollectionID), zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Validate permission level
	if err := collectionsharingdto.ValidatePermissionLevel(input.PermissionLevel); err != nil {
		s.logger.Error("❌ Invalid permission level", zap.String("level", input.PermissionLevel), zap.Error(err))
		return nil, errors.NewAppError("invalid permission level", err)
	}

	// Create use case input
	useCaseInput := &uc.ShareCollectionInputDTO{
		CollectionID:         collectionObjectID,
		RecipientEmail:       input.RecipientEmail,
		PermissionLevel:      input.PermissionLevel,
		ShareWithDescendants: input.ShareWithDescendants,
	}

	// Execute use case
	response, err := s.shareCollectionUseCase.Execute(ctx, useCaseInput, userPassword)
	if err != nil {
		s.logger.Error("❌ Failed to share collection",
			zap.String("collectionID", input.CollectionID),
			zap.String("recipientEmail", input.RecipientEmail),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully shared collection",
		zap.String("collectionID", input.CollectionID),
		zap.String("recipientEmail", input.RecipientEmail),
		zap.String("permissionLevel", input.PermissionLevel))

	return &ShareCollectionOutput{
		Success:            response.Success,
		Message:            response.Message,
		MembershipsCreated: response.MembershipsCreated,
	}, nil
}

// RemoveMember removes a user's access to a collection
func (s *sharingService) RemoveMember(ctx context.Context, input *RemoveMemberInput) (*RemoveMemberOutput, error) {
	// Validate inputs
	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		s.logger.Error("❌ Recipient email is required")
		return nil, errors.NewAppError("recipient email is required", nil)
	}

	// Convert string ID to ObjectID
	collectionObjectID, err := primitive.ObjectIDFromHex(input.CollectionID)
	if err != nil {
		s.logger.Error("❌ Invalid collection ID format", zap.String("id", input.CollectionID), zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Create use case input
	useCaseInput := &uc.RemoveMemberInput{
		CollectionID:          collectionObjectID,
		RecipientEmail:        input.RecipientEmail,
		RemoveFromDescendants: input.RemoveFromDescendants,
	}

	// Execute use case
	response, err := s.removeMemberUseCase.Execute(ctx, useCaseInput)
	if err != nil {
		s.logger.Error("❌ Failed to remove collection member",
			zap.String("collectionID", input.CollectionID),
			zap.String("recipientEmail", input.RecipientEmail),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully removed collection member",
		zap.String("collectionID", input.CollectionID),
		zap.String("recipientEmail", input.RecipientEmail))

	return &RemoveMemberOutput{
		Success: response.Success,
		Message: response.Message,
	}, nil
}

// ListSharedCollections lists all collections shared with the current user
func (s *sharingService) ListSharedCollections(ctx context.Context) (*ListSharedCollectionsOutput, error) {
	// Execute use case
	collections, err := s.listSharedCollectionsUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to list shared collections", zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully listed shared collections",
		zap.Int("count", len(collections)))

	return &ListSharedCollectionsOutput{
		Collections: collections,
		Count:       len(collections),
	}, nil
}

// GetCollectionMembers retrieves the members of a specific collection
func (s *sharingService) GetCollectionMembers(ctx context.Context, collectionID string) ([]*collectiondto.CollectionMembershipDTO, error) {
	// Validate input
	if collectionID == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Convert string ID to ObjectID
	collectionObjectID, err := primitive.ObjectIDFromHex(collectionID)
	if err != nil {
		s.logger.Error("❌ Invalid collection ID format", zap.String("id", collectionID), zap.Error(err))
		return nil, errors.NewAppError("invalid collection ID format", err)
	}

	// Get collection with members
	coll, err := s.getCollectionFromCloudUseCase.Execute(ctx, collectionObjectID)
	if err != nil {
		s.logger.Error("❌ Failed to get collection", zap.String("collectionID", collectionID), zap.Error(err))
		return nil, err
	}
	if coll == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	s.logger.Info("✅ Successfully retrieved collection members",
		zap.String("collectionID", collectionID),
		zap.Int("memberCount", len(coll.Members)))

	return coll.Members, nil
}
