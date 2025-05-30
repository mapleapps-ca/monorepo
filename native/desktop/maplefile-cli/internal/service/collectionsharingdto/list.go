// internal/service/collectionsharingdto/list.go
package collectionsharingdto

import (
	"context"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collectiondto"
)

// ListSharedCollectionsOutput represents the output from listing shared collections
type ListSharedCollectionsOutput struct {
	Collections []*collectiondto.CollectionDTO `json:"collections"`
	Count       int                            `json:"count"`
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
