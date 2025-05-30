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
