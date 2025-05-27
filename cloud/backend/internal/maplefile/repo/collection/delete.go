// cloud/backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}

	// Soft delete: Update state to deleted instead of removing document
	update := bson.M{
		"$set": bson.M{
			"state":       dom_collection.CollectionStateDeleted,
			"modified_at": time.Now(),
		},
	}

	result, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database failed soft deletion error",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	if result.ModifiedCount == 0 {
		impl.Logger.Warn("no collection was soft deleted, may not exist",
			zap.Any("id", id))
		return errors.New("collection not found")
	}

	// Also soft delete all children collections
	childrenFilter := bson.M{"ancestor_ids": id}
	childrenUpdate := bson.M{
		"$set": bson.M{
			"state":       dom_collection.CollectionStateDeleted,
			"modified_at": time.Now(),
		},
	}

	_, err = impl.Collection.UpdateMany(ctx, childrenFilter, childrenUpdate)
	if err != nil {
		impl.Logger.Error("failed to soft delete child collections",
			zap.Any("error", err),
			zap.Any("parent_id", id))
		return err
	}

	return nil
}

func (impl collectionRepositoryImpl) HardDelete(ctx context.Context, id primitive.ObjectID) error {
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
		impl.Logger.Warn("hard deleting collection with children",
			zap.Any("id", id),
			zap.Int64("child_count", childCount))

		// Hard delete all children (recursive delete)
		_, err = impl.Collection.DeleteMany(ctx, bson.M{"ancestor_ids": id})
		if err != nil {
			impl.Logger.Error("failed to hard delete child collections",
				zap.Any("error", err),
				zap.Any("parent_id", id))
			return err
		}
	}

	// Now hard delete the collection itself
	_, err = impl.Collection.DeleteOne(ctx, filter)
	if err != nil {
		impl.Logger.Error("database failed hard deletion error",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	return nil
}
