// cloud/backend/internal/maplefile/repo/filemetadata/get_sync_data.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	dom_sync "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

func (impl fileMetadataRepositoryImpl) GetSyncData(ctx context.Context, userID primitive.ObjectID, cursor *dom_sync.FileSyncCursor, limit int64) (*dom_sync.FileSyncResponse, error) {
	impl.Logger.Debug("Getting file sync data",
		zap.Any("user_id", userID),
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

	// First, get all collection IDs that the user has access to
	// This is a bit complex because we need to join with collections to check access
	collectionFilter := bson.M{
		"$or": []bson.M{
			{"owner_id": userID},             // Collections owned by user
			{"members.recipient_id": userID}, // Collections shared with user
		},
	}

	// Get accessible collection IDs
	collectionCursor, err := impl.Collection.Find(ctx, collectionFilter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		impl.Logger.Error("Failed to get accessible collections for file sync", zap.Error(err))
		return nil, err
	}
	defer collectionCursor.Close(ctx)

	var accessibleCollectionIDs []primitive.ObjectID
	for collectionCursor.Next(ctx) {
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := collectionCursor.Decode(&result); err != nil {
			impl.Logger.Error("Failed to decode collection ID", zap.Error(err))
			continue
		}
		accessibleCollectionIDs = append(accessibleCollectionIDs, result.ID)
	}

	if err := collectionCursor.Err(); err != nil {
		impl.Logger.Error("Cursor error during collection access check", zap.Error(err))
		return nil, err
	}

	// If user has no accessible collections, return empty result
	if len(accessibleCollectionIDs) == 0 {
		impl.Logger.Info("User has no accessible collections for file sync", zap.Any("user_id", userID))
		return &dom_sync.FileSyncResponse{
			Files:      []dom_sync.FileSyncItem{},
			NextCursor: nil,
			HasMore:    false,
		}, nil
	}

	// Build the base filter for files in accessible collections
	baseFilter := bson.M{
		"collection_id": bson.M{"$in": accessibleCollectionIDs},
	}

	// Add cursor-based pagination if cursor is provided
	if cursor != nil && !cursor.LastModified.IsZero() {
		// Find documents modified after the cursor, or same modification time but with ID > cursor ID
		paginationFilter := bson.M{
			"$or": []bson.M{
				{"modified_at": bson.M{"$gt": cursor.LastModified}},
				{
					"$and": []bson.M{
						{"modified_at": cursor.LastModified},
						{"_id": bson.M{"$gt": cursor.LastID}},
					},
				},
			},
		}

		baseFilter = bson.M{
			"$and": []bson.M{
				{"collection_id": bson.M{"$in": accessibleCollectionIDs}},
				paginationFilter,
			},
		}
	}

	// Set up options for pagination and sorting
	findOptions := options.Find().
		SetSort(bson.M{"modified_at": 1, "_id": 1}). // Sort by modified_at ASC, then _id ASC
		SetLimit(limit + 1)                          // Request one extra to check if there are more results

	// Project only the fields we need for sync
	findOptions.SetProjection(bson.M{
		"_id":               1,
		"collection_id":     1,
		"version":           1,
		"modified_at":       1,
		"state":             1,
		"tombstone_version": 1,
		"tombstone_expiry":  1,
	})

	impl.Logger.Debug("Executing file sync query",
		zap.Any("filter", baseFilter),
		zap.Any("options", findOptions),
		zap.Int("accessible_collections", len(accessibleCollectionIDs)))

	// Execute the query
	cursor_result, err := impl.Collection.Find(ctx, baseFilter, findOptions)
	if err != nil {
		impl.Logger.Error("Failed to execute file sync query", zap.Error(err))
		return nil, err
	}
	defer cursor_result.Close(ctx)

	// Decode results
	var syncItems []dom_sync.FileSyncItem
	for cursor_result.Next(ctx) {
		var item dom_sync.FileSyncItem
		if err := cursor_result.Decode(&item); err != nil {
			impl.Logger.Error("Failed to decode file sync item", zap.Error(err))
			continue
		}
		syncItems = append(syncItems, item)
	}

	if err := cursor_result.Err(); err != nil {
		impl.Logger.Error("Cursor error during file sync", zap.Error(err))
		return nil, err
	}

	// Check if there are more results and prepare response
	hasMore := false
	var nextCursor *dom_sync.FileSyncCursor

	if int64(len(syncItems)) > limit {
		hasMore = true
		// Remove the extra item we fetched
		syncItems = syncItems[:limit]
	}

	// Set next cursor if there are more results
	if hasMore && len(syncItems) > 0 {
		lastItem := syncItems[len(syncItems)-1]
		nextCursor = &dom_sync.FileSyncCursor{
			LastModified: lastItem.ModifiedAt,
			LastID:       lastItem.ID,
		}
	}

	response := &dom_sync.FileSyncResponse{
		Files:      syncItems,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}

	impl.Logger.Info("Successfully retrieved file sync data",
		zap.Int("items_count", len(syncItems)),
		zap.Bool("has_more", hasMore),
		zap.Any("user_id", userID))

	return response, nil
}
