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

	// Soft delete: Update state to deleted and nullify most fields, but keep key auditing and hierarchical fields.
	// We use $set for the required audit fields and state, and $unset for others.
	// We keep fields like _id, owner_id, parent_id, ancestor_ids, children, created_at, created_by_user_id,
	// modified_at, modified_by_user_id, state for auditing and historical tracking purposes,
	// including maintaining the hierarchy structure for potential recovery or analysis.
	// Zeroing/nullifying other fields minimizes data retention for deleted items.
	update := bson.M{
		"$set": bson.M{
			"state":       dom_collection.CollectionStateDeleted,
			"modified_at": time.Now(),
			// modified_by_user_id should ideally be set here if available in context for full audit.
		},
		"$unset": bson.M{
			// Unset/remove fields that are not needed after soft deletion,
			// based on the Collection model and the list of fields to keep.
			"encrypted_name":           "",
			"collection_type":          "",
			"encrypted_collection_key": "",
			"members":                  "",
			"version":                  "",
			// Do NOT unset: _id, owner_id, parent_id, ancestor_ids, children,
			// created_at, created_by_user_id, modified_at, modified_by_user_id, state.
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
		// If modifiedCount is 0, it means the document wasn't found or the update didn't change anything
		// (e.g., it was already deleted).
		// We assume this means the document was not found for deletion or was already in the deleted state.
		impl.Logger.Warn("no collection was soft deleted, may not exist or already deleted",
			zap.Any("id", id))
		return errors.New("collection not found or already deleted")
	}

	// Also soft delete all children collections recursively using ancestor_ids.
	// This update applies the same soft-delete logic as the parent.
	childrenFilter := bson.M{"ancestor_ids": id}
	childrenUpdate := bson.M{
		"$set": bson.M{
			"state":       dom_collection.CollectionStateDeleted,
			"modified_at": time.Now(),
			// modified_by_user_id should ideally be set here if available.
		},
		"$unset": bson.M{
			// Unset/remove fields for children as well, keeping audit and hierarchy fields.
			"encrypted_name":           "",
			"collection_type":          "",
			"encrypted_collection_key": "",
			"members":                  "",
			"version":                  "",
			// Do NOT unset: _id, owner_id, parent_id, ancestor_ids, children,
			// created_at, created_by_user_id, modified_at, modified_by_user_id, state.
		},
	}

	// We don't check ModifiedCount for children update as it's valid to delete a collection with no children.
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
