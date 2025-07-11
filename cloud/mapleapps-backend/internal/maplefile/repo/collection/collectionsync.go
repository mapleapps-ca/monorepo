// cloud/mapleapps-backend/internal/maplefile/repo/collection/collectionsync.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

// GetAllByUserIDAndAnyType uses the general table when you need all collections regardless of access type
// This method demonstrates querying without access_type filtering to avoid ALLOW FILTERING
func (impl *collectionRepositoryImpl) GetAllByUserIDAndAnyType(ctx context.Context, userID gocql.UUID, cursor *dom_collection.CollectionSyncCursor, limit int64) (*dom_collection.CollectionSyncResponse, error) {
	var query string
	var args []any

	// Key Insight: We can query all collections for a user efficiently because user_id is the partition key
	// We select access_type in the result set so we can filter or categorize after retrieval
	if cursor == nil {
		query = `SELECT collection_id, modified_at, access_type FROM
			maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? LIMIT ?`
		args = []any{userID, limit}
	} else {
		query = `SELECT collection_id, modified_at, access_type FROM
			maplefile_collections_by_user_id_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND (modified_at, collection_id) > (?, ?) LIMIT ?`
		args = []any{userID, cursor.LastModified, cursor.LastID, limit}
	}

	iter := impl.Session.Query(query, args...).WithContext(ctx).Iter()

	var syncItems []dom_collection.CollectionSyncItem
	var lastModified time.Time
	var lastID gocql.UUID

	// Critical Fix: We must scan all three selected columns
	var collectionID gocql.UUID
	var modifiedAt time.Time
	var accessType string

	for iter.Scan(&collectionID, &modifiedAt, &accessType) {
		// Get minimal sync data for this collection
		syncItem, err := impl.getCollectionSyncItem(ctx, collectionID)
		if err != nil {
			impl.Logger.Warn("failed to get sync item for collection",
				zap.String("collection_id", collectionID.String()),
				zap.String("access_type", accessType),
				zap.Error(err))
			continue
		}

		if syncItem != nil {
			syncItems = append(syncItems, *syncItem)
			lastModified = modifiedAt
			lastID = collectionID
		}
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get collection sync data: %w", err)
	}

	// Prepare response
	response := &dom_collection.CollectionSyncResponse{
		Collections: syncItems,
		HasMore:     len(syncItems) == int(limit),
	}

	// Set next cursor if there are more results
	if response.HasMore {
		response.NextCursor = &dom_collection.CollectionSyncCursor{
			LastModified: lastModified,
			LastID:       lastID,
		}
	}

	return response, nil
}

// GetCollectionSyncData uses the access-type-specific table for optimal performance
// This method demonstrates the power of compound partition keys in Cassandra
func (impl *collectionRepositoryImpl) GetCollectionSyncData(ctx context.Context, userID gocql.UUID, cursor *dom_collection.CollectionSyncCursor, limit int64) (*dom_collection.CollectionSyncResponse, error) {
	var query string
	var args []any

	// Key Insight: With the compound partition key (user_id, access_type), this query is lightning fast
	// Cassandra can directly access the specific partition without any filtering or scanning
	if cursor == nil {
		query = `SELECT collection_id, modified_at FROM
			maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'owner' LIMIT ?`
		args = []any{userID, limit}
	} else {
		query = `SELECT collection_id, modified_at FROM
			maplefile_collections_by_user_id_and_access_type_with_desc_modified_at_and_asc_collection_id
			WHERE user_id = ? AND access_type = 'owner' AND (modified_at, collection_id) > (?, ?) LIMIT ?`
		args = []any{userID, cursor.LastModified, cursor.LastID, limit}
	}

	iter := impl.Session.Query(query, args...).WithContext(ctx).Iter()

	var syncItems []dom_collection.CollectionSyncItem
	var lastModified time.Time
	var lastID gocql.UUID

	var collectionID gocql.UUID
	var modifiedAt time.Time

	for iter.Scan(&collectionID, &modifiedAt) {
		// Get minimal sync data for this collection
		syncItem, err := impl.getCollectionSyncItem(ctx, collectionID)
		if err != nil {
			impl.Logger.Warn("failed to get sync item for collection",
				zap.String("collection_id", collectionID.String()),
				zap.Error(err))
			continue
		}

		if syncItem != nil {
			syncItems = append(syncItems, *syncItem)
			lastModified = modifiedAt
			lastID = collectionID
		}
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to get collection sync data: %w", err)
	}

	// Prepare response
	response := &dom_collection.CollectionSyncResponse{
		Collections: syncItems,
		HasMore:     len(syncItems) == int(limit),
	}

	// Set next cursor if there are more results
	if response.HasMore {
		response.NextCursor = &dom_collection.CollectionSyncCursor{
			LastModified: lastModified,
			LastID:       lastID,
		}
	}

	return response, nil
}

// Helper method to get minimal sync data for a collection
func (impl *collectionRepositoryImpl) getCollectionSyncItem(ctx context.Context, collectionID gocql.UUID) (*dom_collection.CollectionSyncItem, error) {
	var (
		id                          gocql.UUID
		version, tombstoneVersion   uint64
		modifiedAt, tombstoneExpiry time.Time
		state                       string
		parentID                    gocql.UUID
	)

	query := `SELECT id, version, modified_at, state, parent_id, tombstone_version, tombstone_expiry
		FROM maplefile_collections_by_id WHERE id = ?`

	err := impl.Session.Query(query, collectionID).WithContext(ctx).Scan(
		&id, &version, &modifiedAt, &state, &parentID, &tombstoneVersion, &tombstoneExpiry)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collection sync item: %w", err)
	}

	syncItem := &dom_collection.CollectionSyncItem{
		ID:               id,
		Version:          version,
		ModifiedAt:       modifiedAt,
		State:            state,
		TombstoneVersion: tombstoneVersion,
		TombstoneExpiry:  tombstoneExpiry,
	}

	// Only include ParentID if it's valid
	if impl.isValidUUID(parentID) {
		syncItem.ParentID = &parentID
	}

	return syncItem, nil
}
