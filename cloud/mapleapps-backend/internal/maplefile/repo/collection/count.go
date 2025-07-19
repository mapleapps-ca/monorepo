// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/collection/count.go
package collection

import (
	"context"
	"fmt"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

// CountOwnedCollections counts all collections (folders + albums) owned by the user
func (impl *collectionRepositoryImpl) CountOwnedCollections(ctx context.Context, userID gocql.UUID) (int, error) {
	return impl.countCollectionsByUserAndType(ctx, userID, "owner", "")
}

// CountSharedCollections counts all collections (folders + albums) shared with the user
func (impl *collectionRepositoryImpl) CountSharedCollections(ctx context.Context, userID gocql.UUID) (int, error) {
	return impl.countCollectionsByUserAndType(ctx, userID, "member", "")
}

// CountOwnedFolders counts only folders owned by the user
func (impl *collectionRepositoryImpl) CountOwnedFolders(ctx context.Context, userID gocql.UUID) (int, error) {
	return impl.countCollectionsByUserAndType(ctx, userID, "owner", dom_collection.CollectionTypeFolder)
}

// CountSharedFolders counts only folders shared with the user
func (impl *collectionRepositoryImpl) CountSharedFolders(ctx context.Context, userID gocql.UUID) (int, error) {
	return impl.countCollectionsByUserAndType(ctx, userID, "member", dom_collection.CollectionTypeFolder)
}

// countCollectionsByUserAndType is a helper method that efficiently counts collections
// filterType: empty string for all types, or specific type like "folder"
func (impl *collectionRepositoryImpl) countCollectionsByUserAndType(ctx context.Context, userID gocql.UUID, accessType, filterType string) (int, error) {
	// Use the access-type-specific table for efficient querying
	query := `SELECT collection_id FROM maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ? AND access_type = ?`

	iter := impl.Session.Query(query, userID, accessType).WithContext(ctx).Iter()

	count := 0
	var collectionID gocql.UUID

	// Iterate through results and count based on criteria
	for iter.Scan(&collectionID) {
		// Get the collection to check state and type
		collection, err := impl.getBaseCollection(ctx, collectionID)
		if err != nil {
			impl.Logger.Warn("failed to get collection for counting",
				zap.String("collection_id", collectionID.String()),
				zap.Error(err))
			continue
		}

		if collection == nil {
			continue
		}

		// Only count active collections
		if collection.State != dom_collection.CollectionStateActive {
			continue
		}

		// Filter by type if specified
		if filterType != "" && collection.CollectionType != filterType {
			continue
		}

		count++
	}

	if err := iter.Close(); err != nil {
		impl.Logger.Error("failed to count collections",
			zap.String("user_id", userID.String()),
			zap.String("access_type", accessType),
			zap.String("filter_type", filterType),
			zap.Error(err))
		return 0, fmt.Errorf("failed to count collections: %w", err)
	}

	impl.Logger.Debug("counted collections successfully",
		zap.String("user_id", userID.String()),
		zap.String("access_type", accessType),
		zap.String("filter_type", filterType),
		zap.Int("count", count))

	return count, nil
}

// Alternative optimized implementation for when you need both owned and shared counts
// This reduces database round trips by querying once and separating in memory
func (impl *collectionRepositoryImpl) countCollectionsSummary(ctx context.Context, userID gocql.UUID, filterType string) (ownedCount, sharedCount int, err error) {
	// Query all collections for the user using the general table
	query := `SELECT collection_id, access_type FROM maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
		WHERE user_id = ?`

	iter := impl.Session.Query(query, userID).WithContext(ctx).Iter()

	var collectionID gocql.UUID
	var accessType string

	for iter.Scan(&collectionID, &accessType) {
		// Get the collection to check state and type
		collection, getErr := impl.getBaseCollection(ctx, collectionID)
		if getErr != nil {
			impl.Logger.Warn("failed to get collection for counting summary",
				zap.String("collection_id", collectionID.String()),
				zap.Error(getErr))
			continue
		}

		if collection == nil {
			continue
		}

		// Only count active collections
		if collection.State != dom_collection.CollectionStateActive {
			continue
		}

		// Filter by type if specified
		if filterType != "" && collection.CollectionType != filterType {
			continue
		}

		// Count based on access type
		switch accessType {
		case "owner":
			ownedCount++
		case "member":
			sharedCount++
		}
	}

	if err = iter.Close(); err != nil {
		impl.Logger.Error("failed to count collections summary",
			zap.String("user_id", userID.String()),
			zap.String("filter_type", filterType),
			zap.Error(err))
		return 0, 0, fmt.Errorf("failed to count collections summary: %w", err)
	}

	impl.Logger.Debug("counted collections summary successfully",
		zap.String("user_id", userID.String()),
		zap.String("filter_type", filterType),
		zap.Int("owned_count", ownedCount),
		zap.Int("shared_count", sharedCount))

	return ownedCount, sharedCount, nil
}
