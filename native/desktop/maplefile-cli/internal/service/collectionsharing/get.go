// internal/service/collectionsharing/get.go
package collectionsharing

import (
	"context"

	"go.uber.org/zap"

	"github.com/gocql/gocql"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	uc_collectionsharingdto "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectiondto"
)

// CollectionSharingGetMembersService defines the interface for getting members via collection sharing operations
type CollectionSharingGetMembersService interface {
	Execute(ctx context.Context, collectionID gocql.UUID) ([]*collectiondto.CollectionMembershipDTO, error)
}

// collectionSharingGetMembersServiceImpl implements the CollectionSharingGetMembersService interface
type collectionSharingGetMembersServiceImpl struct {
	logger                        *zap.Logger
	getCollectionFromCloudUseCase uc_collectionsharingdto.GetCollectionFromCloudUseCase
}

// NewGetCollectionMembersService creates a new get members via collection sharing service
func NewGetCollectionMembersService(
	logger *zap.Logger,
	getCollectionFromCloudUseCase uc_collectionsharingdto.GetCollectionFromCloudUseCase,
) CollectionSharingGetMembersService {
	logger = logger.Named("CollectionSharingGetMembersService")
	return &collectionSharingGetMembersServiceImpl{
		logger:                        logger,
		getCollectionFromCloudUseCase: getCollectionFromCloudUseCase,
	}
}

// Execute retrieves the members of a specific collection
func (s *collectionSharingGetMembersServiceImpl) Execute(ctx context.Context, collectionID gocql.UUID) ([]*collectiondto.CollectionMembershipDTO, error) {
	// Validate input
	if collectionID.String() == "" {
		s.logger.Error("❌ Collection ID is required")
		return nil, errors.NewAppError("collection ID is required", nil)
	}

	// Get collection with members
	coll, err := s.getCollectionFromCloudUseCase.Execute(ctx, collectionID)
	if err != nil {
		s.logger.Error("❌ Failed to get collection",
			zap.String("collectionID", collectionID.String()),
			zap.Error(err))
		return nil, err
	}
	if coll == nil {
		return nil, errors.NewAppError("collection not found", nil)
	}

	s.logger.Info("✅ Successfully retrieved collection members",
		zap.String("collectionID", collectionID.String()),
		zap.Int("memberCount", len(coll.Members)))

	return coll.Members, nil
}
