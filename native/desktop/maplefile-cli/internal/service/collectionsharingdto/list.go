// internal/service/collectionsharingdto/list.go
package collectionsharingdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
	uc "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/usecase/collectionsharingdto"
)

// ListSharedCollectionsOutput represents the output from listing shared collections
type ListSharedCollectionsOutput struct {
	Collections []*collectiondto.CollectionDTO `json:"collections"`
	Count       int                            `json:"count"`
}

// ListSharedCollectionsService defines the interface for collection sharing operations
type ListSharedCollectionsService interface {
	Execute(ctx context.Context) (*ListSharedCollectionsOutput, error)
}

// listSharedCollectionsServiceImpl implements the SharingService interface
type listSharedCollectionsServiceImpl struct {
	logger                       *zap.Logger
	listSharedCollectionsUseCase uc.ListSharedCollectionsUseCase
}

// NewListSharedCollectionsService creates a new collection sharing service
func NewListSharedCollectionsService(
	logger *zap.Logger,
	listSharedCollectionsUseCase uc.ListSharedCollectionsUseCase,
) ListSharedCollectionsService {
	logger = logger.Named("ListSharedCollectionsService")
	return &listSharedCollectionsServiceImpl{
		logger:                       logger,
		listSharedCollectionsUseCase: listSharedCollectionsUseCase,
	}
}

// Execute lists all collections shared with the current user
func (s *listSharedCollectionsServiceImpl) Execute(ctx context.Context) (*ListSharedCollectionsOutput, error) {
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
