// cloud/backend/internal/maplefile/repo/collection/get_filtered.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) GetCollectionsWithFilter(ctx context.Context, options dom_collection.CollectionFilterOptions) (*dom_collection.CollectionFilterResult, error) {
	impl.Logger.Debug("Getting collections with filter",
		zap.Any("options", options))

	result := &dom_collection.CollectionFilterResult{
		OwnedCollections:  []*dom_collection.Collection{},
		SharedCollections: []*dom_collection.Collection{},
		TotalCount:        0,
	}

	// Validate filter options
	if !options.IsValid() {
		impl.Logger.Warn("Invalid filter options - no collections will be returned",
			zap.Any("options", options))
		return result, nil
	}

	// Get owned collections if requested
	if options.IncludeOwned {
		ownedCollections, err := impl.getOwnedCollections(ctx, options.UserID)
		if err != nil {
			impl.Logger.Error("Failed to get owned collections",
				zap.Any("error", err),
				zap.Any("user_id", options.UserID))
			return nil, err
		}
		result.OwnedCollections = ownedCollections
		impl.Logger.Debug("Retrieved owned collections",
			zap.Int("count", len(ownedCollections)),
			zap.Any("user_id", options.UserID))
	}

	// Get shared collections if requested
	if options.IncludeShared {
		sharedCollections, err := impl.getSharedCollections(ctx, options.UserID)
		if err != nil {
			impl.Logger.Error("Failed to get shared collections",
				zap.Any("error", err),
				zap.Any("user_id", options.UserID))
			return nil, err
		}
		result.SharedCollections = sharedCollections
		impl.Logger.Debug("Retrieved shared collections",
			zap.Int("count", len(sharedCollections)),
			zap.Any("user_id", options.UserID))
	}

	// Calculate total count
	result.TotalCount = len(result.OwnedCollections) + len(result.SharedCollections)

	impl.Logger.Info("Successfully retrieved filtered collections",
		zap.Int("owned_count", len(result.OwnedCollections)),
		zap.Int("shared_count", len(result.SharedCollections)),
		zap.Int("total_count", result.TotalCount),
		zap.Any("user_id", options.UserID))

	return result, nil
}

// getOwnedCollections retrieves collections owned by the specified user
func (impl collectionRepositoryImpl) getOwnedCollections(ctx context.Context, userID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	filter := bson.M{"owner_id": userID}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get owned collections error", zap.Any("error", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []*dom_collection.Collection
	if err = cursor.All(ctx, &collections); err != nil {
		impl.Logger.Error("database decode owned collections error", zap.Any("error", err))
		return nil, err
	}

	return collections, nil
}

// getSharedCollections retrieves collections shared with the specified user
func (impl collectionRepositoryImpl) getSharedCollections(ctx context.Context, userID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	// Find collections where user is in members array as recipient
	// Exclude collections owned by the user to avoid duplicates
	filter := bson.M{
		"members.recipient_id": userID,
		"owner_id":             bson.M{"$ne": userID},
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get shared collections error", zap.Any("error", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []*dom_collection.Collection
	if err = cursor.All(ctx, &collections); err != nil {
		impl.Logger.Error("database decode shared collections error", zap.Any("error", err))
		return nil, err
	}

	return collections, nil
}
