// internal/service/collectionsharingdto/get.go
package collectionsharingdto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

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
