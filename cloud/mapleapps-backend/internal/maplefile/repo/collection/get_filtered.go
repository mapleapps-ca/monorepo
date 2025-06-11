// cloud/mapleapps-backend/internal/maplefile/repo/collection/get_filtered.go
package collection

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
	"go.uber.org/zap"
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
		result.OwnedCollections, err = impl.getOwnedCollectionsOptimized(ctx, options.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get owned collections: %w", err)
		}
	}

	// Get shared collections if requested
	if options.IncludeShared {
		result.SharedCollections, err = impl.getSharedCollectionsOptimized(ctx, options.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get shared collections: %w", err)
		}
	}

	result.TotalCount = len(result.OwnedCollections) + len(result.SharedCollections)

	impl.Logger.Debug("completed filtered collection query",
		zap.String("user_id", options.UserID.String()),
		zap.Bool("include_owned", options.IncludeOwned),
		zap.Bool("include_shared", options.IncludeShared),
		zap.Int("owned_count", len(result.OwnedCollections)),
		zap.Int("shared_count", len(result.SharedCollections)),
		zap.Int("total_count", result.TotalCount))

	return result, nil
}

// OPTIMIZED: Uses the access-type-specific table for maximum efficiency
// This method demonstrates how we can get owned collections without any filtering overhead
func (impl *collectionRepositoryImpl) getOwnedCollectionsOptimized(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	// No memory filtering needed, no wasted I/O, just pure efficiency
	return impl.GetAllByUserID(ctx, userID)
}

// OPTIMIZED: Also uses the access-type-specific table
func (impl *collectionRepositoryImpl) getSharedCollectionsOptimized(ctx context.Context, userID gocql.UUID) ([]*dom_collection.Collection, error) {
	// Direct partition access for member-type collections
	return impl.GetCollectionsSharedWithUser(ctx, userID)
}

// NEW METHOD: Demonstrates the alternative approach when you need both types efficiently
// This method shows when you might want to use the original table instead of making two separate queries
func (impl *collectionRepositoryImpl) GetCollectionsWithFilterSingleQuery(ctx context.Context, options dom_collection.CollectionFilterOptions) (*dom_collection.CollectionFilterResult, error) {
	if !options.IsValid() {
		return nil, fmt.Errorf("invalid filter options: at least one filter must be enabled")
	}

	// Strategy decision: If we need both owned AND shared collections,
	// it might be more efficient to query the original table once and separate them in memory
	// This demonstrates the trade-offs between different approaches

	if options.ShouldIncludeAll() {
		return impl.getAllCollectionsAndSeparate(ctx, options.UserID)
	}

	// If we only need one type, use the optimized single-type methods
	return impl.GetCollectionsWithFilter(ctx, options)
}

// Helper method that demonstrates memory-based separation when it's more efficient
func (impl *collectionRepositoryImpl) getAllCollectionsAndSeparate(ctx context.Context, userID gocql.UUID) (*dom_collection.CollectionFilterResult, error) {
	result := &dom_collection.CollectionFilterResult{
		OwnedCollections:  []*dom_collection.Collection{},
		SharedCollections: []*dom_collection.Collection{},
		TotalCount:        0,
	}

	// Query the original table to get all collections for the user
	allCollections, err := impl.GetAllUserCollections(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all user collections: %w", err)
	}

	// Separate owned from shared collections in memory
	for _, collection := range allCollections {
		if collection.OwnerID == userID {
			result.OwnedCollections = append(result.OwnedCollections, collection)
		} else {
			// If the user is not the owner but has access, they must be a member
			result.SharedCollections = append(result.SharedCollections, collection)
		}
	}

	result.TotalCount = len(result.OwnedCollections) + len(result.SharedCollections)

	impl.Logger.Debug("completed single-query filtered collection retrieval",
		zap.String("user_id", userID.String()),
		zap.Int("total_retrieved", len(allCollections)),
		zap.Int("owned_count", len(result.OwnedCollections)),
		zap.Int("shared_count", len(result.SharedCollections)))

	return result, nil
}

// NEW METHOD: Advanced filtering with pagination support
// This demonstrates how to implement efficient pagination across filtered results
func (impl *collectionRepositoryImpl) GetCollectionsWithFilterPaginated(ctx context.Context, options dom_collection.CollectionFilterOptions, limit int64, cursor *dom_collection.CollectionSyncCursor) (*dom_collection.CollectionFilterResult, error) {
	if !options.IsValid() {
		return nil, fmt.Errorf("invalid filter options: at least one filter must be enabled")
	}

	result := &dom_collection.CollectionFilterResult{
		OwnedCollections:  []*dom_collection.Collection{},
		SharedCollections: []*dom_collection.Collection{},
		TotalCount:        0,
	}

	if options.IncludeOwned {
		ownedCollections, err := impl.getOwnedCollectionsPaginated(ctx, options.UserID, limit, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to get paginated owned collections: %w", err)
		}
		result.OwnedCollections = ownedCollections
	}

	if options.IncludeShared {
		sharedCollections, err := impl.getSharedCollectionsPaginated(ctx, options.UserID, limit, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to get paginated shared collections: %w", err)
		}
		result.SharedCollections = sharedCollections
	}

	result.TotalCount = len(result.OwnedCollections) + len(result.SharedCollections)

	return result, nil
}

// Helper method for paginated owned collections
func (impl *collectionRepositoryImpl) getOwnedCollectionsPaginated(ctx context.Context, userID gocql.UUID, limit int64, cursor *dom_collection.CollectionSyncCursor) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID
	var query string
	var args []any

	// Build paginated query using the access-type-specific table
	if cursor == nil {
		query = `SELECT collection_id FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'owner' AND state = ? LIMIT ?`
		args = []any{userID, dom_collection.CollectionStateActive, limit}
	} else {
		query = `SELECT collection_id FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'owner' AND state = ? AND (modified_at, collection_id) > (?, ?) LIMIT ?`
		args = []any{userID, dom_collection.CollectionStateActive, cursor.LastModified, cursor.LastID, limit}
	}

	iter := impl.Session.Query(query, args...).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get paginated owned collections: %w", err)
	}

	return impl.loadMultipleCollectionsWithMembers(ctx, collectionIDs)
}

// Helper method for paginated shared collections
func (impl *collectionRepositoryImpl) getSharedCollectionsPaginated(ctx context.Context, userID gocql.UUID, limit int64, cursor *dom_collection.CollectionSyncCursor) ([]*dom_collection.Collection, error) {
	var collectionIDs []gocql.UUID
	var query string
	var args []any

	// Build paginated query using the access-type-specific table
	if cursor == nil {
		query = `SELECT collection_id FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'member' AND state = ? LIMIT ?`
		args = []any{userID, dom_collection.CollectionStateActive, limit}
	} else {
		query = `SELECT collection_id FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'member' AND state = ? AND (modified_at, collection_id) > (?, ?) LIMIT ?`
		args = []any{userID, dom_collection.CollectionStateActive, cursor.LastModified, cursor.LastID, limit}
	}

	iter := impl.Session.Query(query, args...).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	for iter.Scan(&collectionID) {
		collectionIDs = append(collectionIDs, collectionID)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get paginated shared collections: %w", err)
	}

	return impl.loadMultipleCollectionsWithMembers(ctx, collectionIDs)
}
