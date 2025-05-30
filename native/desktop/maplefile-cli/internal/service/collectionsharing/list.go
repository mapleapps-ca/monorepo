// internal/service/collectionsharing/list.go
package collectionsharing

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
)

// CollectionSharingListService represents the output from listing shared collections
type CollectionSharingListService struct {
	Collections []*collectiondto.CollectionDTO `json:"collections"`
	Count       int                            `json:"count"`
}

// ListSharedCollectionsService defines the interface for collection sharing operations
type ListSharedCollectionsService interface {
	Execute(ctx context.Context) (*CollectionSharingListService, error)
}

// collectionSharingListServiceImpl implements the SharingService interface
type collectionSharingListServiceImpl struct {
	logger                       *zap.Logger
	listSharedCollectionsUseCase uc.ListSharedCollectionsUseCase
}

// NewListSharedCollectionsService creates a new collection sharing service
func NewListSharedCollectionsService(
	logger *zap.Logger,
	listSharedCollectionsUseCase uc.ListSharedCollectionsUseCase,
) ListSharedCollectionsService {
	logger = logger.Named("CollectionSharingListService")
	return &collectionSharingListServiceImpl{
		logger:                       logger,
		listSharedCollectionsUseCase: listSharedCollectionsUseCase,
	}
}

// Execute lists all collections shared with the current user
func (s *collectionSharingListServiceImpl) Execute(ctx context.Context) (*CollectionSharingListService, error) {
	// Execute use case
	collections, err := s.listSharedCollectionsUseCase.Execute(ctx)
	if err != nil {
		s.logger.Error("❌ Failed to list shared collections", zap.Error(err))
		return nil, err
	}

	s.logger.Info("✅ Successfully listed shared collections",
		zap.Int("count", len(collections)))

	return &CollectionSharingListService{
		Collections: collections,
		Count:       len(collections),
	}, nil
}
