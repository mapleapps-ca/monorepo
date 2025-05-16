// cloud/backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (impl collectionRepositoryImpl) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}

	// First check if this collection has children
	childrenFilter := bson.M{"parent_id": id}
	childCount, err := impl.Collection.CountDocuments(ctx, childrenFilter)
	if err != nil {
		impl.Logger.Error("failed to check for children collections",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	// If collection has children, we need to handle them
	if childCount > 0 {
		impl.Logger.Warn("deleting collection with children",
			zap.Any("id", id),
			zap.Int64("child_count", childCount))

		// Option 1: Delete all children (recursive delete)
		_, err = impl.Collection.DeleteMany(ctx, bson.M{"ancestor_ids": id})
		if err != nil {
			impl.Logger.Error("failed to delete child collections",
				zap.Any("error", err),
				zap.Any("parent_id", id))
			return err
		}
	}

	// Now delete the collection itself
	_, err = impl.Collection.DeleteOne(ctx, filter)
	if err != nil {
		impl.Logger.Error("database failed deletion error",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	return nil
}
