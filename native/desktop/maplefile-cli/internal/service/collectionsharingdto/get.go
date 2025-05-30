// internal/service/collectionsharingdto/get.go
package collectionsharingdto

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	uc_collectionsharingdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
)

// GetCollectionMembersService defines the interface for collection sharing operations
type GetCollectionMembersService interface {
	Execute(ctx context.Context, collectionID string) ([]*collectiondto.CollectionMembershipDTO, error)
}

// getCollectionMembersService implements the SharingService interface
type getCollectionMembersServiceImpl struct {
	logger                        *zap.Logger
	getCollectionFromCloudUseCase uc_collectionsharingdto.GetCollectionFromCloudUseCase
}

// NewGetCollectionMembersService creates a new collection sharing service
func NewGetCollectionMembersService(
	logger *zap.Logger,
	getCollectionFromCloudUseCase uc_collectionsharingdto.GetCollectionFromCloudUseCase,
) GetCollectionMembersService {
	logger = logger.Named("GetCollectionMembersService")
	return &getCollectionMembersServiceImpl{
		logger:                        logger,
		getCollectionFromCloudUseCase: getCollectionFromCloudUseCase,
	}
}

// Execute retrieves the members of a specific collection
func (s *getCollectionMembersServiceImpl) Execute(ctx context.Context, collectionID string) ([]*collectiondto.CollectionMembershipDTO, error) {
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
