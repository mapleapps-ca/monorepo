// cloud/backend/internal/maplefile/repo/collection/sync.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	dom_sync "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/sync"
)

func (impl collectionRepositoryImpl) GetCollectionSyncData(ctx context.Context, userID primitive.ObjectID, cursor *dom_sync.SyncCursor, limit int64) (*dom_sync.CollectionSyncResponse, error) {
	impl.Logger.Debug("Getting collection sync data",
		zap.Any("user_id", userID),
		zap.Any("cursor", cursor),
		zap.Int64("limit", limit))

	// Build the base filter for collections the user has access to
	// User has access if they are either owner or member
	baseFilter := bson.M{
		"$or": []bson.M{
			{"owner_id": userID},             // Collections owned by user
			{"members.recipient_id": userID}, // Collections shared with user
		},
	}

	// Add cursor-based pagination if cursor is provided
	if cursor != nil && !cursor.LastModified.IsZero() {
		// Find documents modified after the cursor, or same modification time but with ID > cursor ID
		baseFilter["$or"] = []bson.M{
			{"modified_at": bson.M{"$gt": cursor.LastModified}},
			{
				"$and": []bson.M{
					{"modified_at": cursor.LastModified},
					{"_id": bson.M{"$gt": cursor.LastID}},
				},
			},
		}

		// Ensure we still respect access control by ANDing the access filter
		accessFilter := bson.M{
			"$or": []bson.M{
				{"owner_id": userID},
				{"members.recipient_id": userID},
			},
		}

		baseFilter = bson.M{
			"$and": []bson.M{
				accessFilter,
				baseFilter,
			},
		}
	}

	// Set up options for pagination and sorting
	findOptions := options.Find().
		SetSort(bson.D{bson.E{Key: "modified_at", Value: 1}, bson.E{Key: "_id", Value: 1}}). // Sort by modified_at ASC, then _id ASC
		SetLimit(limit + 1)                                                                  // Request one extra to check if there are more results

	// Project only the fields we need for sync
	findOptions.SetProjection(bson.M{
		"_id":         1,
		"version":     1,
		"modified_at": 1,
		"state":       1,
		"parent_id":   1,
	})

	impl.Logger.Debug("Executing collection sync query",
		zap.Any("filter", baseFilter),
		zap.Any("options", findOptions))

	// Execute the query
	cursor_result, err := impl.Collection.Find(ctx, baseFilter, findOptions)
	if err != nil {
		impl.Logger.Error("Failed to execute collection sync query", zap.Error(err))
		return nil, err
	}
	defer cursor_result.Close(ctx)

	// Decode results
	var syncItems []dom_sync.CollectionSyncItem
	for cursor_result.Next(ctx) {
		var item dom_sync.CollectionSyncItem
		if err := cursor_result.Decode(&item); err != nil {
			impl.Logger.Error("Failed to decode collection sync item", zap.Error(err))
			continue
		}
		syncItems = append(syncItems, item)
	}

	if err := cursor_result.Err(); err != nil {
		impl.Logger.Error("Cursor error during collection sync", zap.Error(err))
		return nil, err
	}

	// Check if there are more results and prepare response
	hasMore := false
	var nextCursor *dom_sync.SyncCursor

	if int64(len(syncItems)) > limit {
		hasMore = true
		// Remove the extra item we fetched
		syncItems = syncItems[:limit]
	}

	// Set next cursor if there are more results
	if hasMore && len(syncItems) > 0 {
		lastItem := syncItems[len(syncItems)-1]
		nextCursor = &dom_sync.SyncCursor{
			LastModified: lastItem.ModifiedAt,
			LastID:       lastItem.ID,
		}
	}

	response := &dom_sync.CollectionSyncResponse{
		Collections: syncItems,
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}

	impl.Logger.Info("Successfully retrieved collection sync data",
		zap.Int("items_count", len(syncItems)),
		zap.Bool("has_more", hasMore),
		zap.Any("user_id", userID))

	return response, nil
}
