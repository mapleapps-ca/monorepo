// internal/service/collectionsharing/remove.go
package collectionsharing

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
)

// RemoveMemberInput represents input for removing a member at the service level
type RemoveMemberInput struct {
	CollectionID          gocql.UUID `json:"collection_id"`
	RecipientEmail        string     `json:"recipient_email"`
	RemoveFromDescendants bool       `json:"remove_from_descendants"`
}

// RemoveMemberOutput represents the output from removing a member
type RemoveMemberOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// CollectionSharingRemoveMembersService defines the interface for collection sharing operations
type CollectionSharingRemoveMembersService interface {
	Execute(ctx context.Context, input *RemoveMemberInput) (*RemoveMemberOutput, error)
}

type collectionSharingRemoveMembersServiceImpl struct {
	logger              *zap.Logger
	removeMemberUseCase uc.RemoveMemberUseCase
}

// NewRemoveMemberCollectionSharingService creates a new collection sharing service
func NewRemoveMemberCollectionSharingService(
	logger *zap.Logger,
	removeMemberUseCase uc.RemoveMemberUseCase,
) CollectionSharingRemoveMembersService {
	logger = logger.Named("CollectionSharingRemoveMembersService")
	return &collectionSharingRemoveMembersServiceImpl{
		logger:              logger,
		removeMemberUseCase: removeMemberUseCase,
	}
}

func (s *collectionSharingRemoveMembersServiceImpl) Execute(ctx context.Context, input *RemoveMemberInput) (*RemoveMemberOutput, error) {
	// Validate inputs
	if input == nil {
		s.logger.Error("❌ Input is required")
		return nil, errors.NewAppError("input is required", nil)
	}
	if input.CollectionID.String() == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}
	if input.RecipientEmail == "" {
		s.logger.Error("❌ Recipient email is required")
		return nil, errors.NewAppError("recipient email is required", nil)
	}

	// Convert string ID to ObjectID
	collectionObjectID := input.CollectionID

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
			zap.String("collectionID", input.CollectionID.String()),
			zap.String("recipientEmail", input.RecipientEmail),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully removed collection member",
		zap.String("collectionID", input.CollectionID.String()),
		zap.String("recipientEmail", input.RecipientEmail))

	return &RemoveMemberOutput{
		Success: response.Success,
		Message: response.Message,
	}, nil
}
