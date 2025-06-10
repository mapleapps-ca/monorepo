// cloud/mapleapps-backend/internal/maplefile/repo/collection/get_filtered.go
package collection

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) GetCollectionsWithFilter(ctx context.Context, options dom_collection.CollectionFilterOptions) (*dom_collection.CollectionFilterResult, error) {
	if !options.IsValid() {
		return nil, fmt.Errorf("invalid filter options: at least one filter must be enabled")
	}

	result := &dom_collection.CollectionFilterResult{
		OwnedCollections:  []*dom_collection.Collection{},
		SharedCollections: []*dom_collection.Collection{},
		TotalCount:        0,
	}

	var err error

	// Get owned collections if requested
	if options.IncludeOwned {
		result.OwnedCollections, err = impl.getOwnedCollections(ctx, options.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get owned collections: %w", err)
		}
	}

	// Get shared collections if requested
	if options.IncludeShared {
		result.SharedCollections, err = impl.getSharedCollections(ctx, options.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get shared collections: %w", err)
		}
	}

	result.TotalCount = len(result.OwnedCollections) + len(result.SharedCollections)

	return result, nil
}

func (impl *collectionRepositoryImpl) getOwnedCollections(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	return impl.GetAllByUserID(ctx, userID)
}

func (impl *collectionRepositoryImpl) getSharedCollections(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	return impl.GetCollectionsSharedWithUser(ctx, userID)
}
