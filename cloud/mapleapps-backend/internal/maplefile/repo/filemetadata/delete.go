// cloud/backend/internal/maplefile/repo/filemetadata/delete.go
package filemetadata

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/file"
)

func (impl fileMetadataRepositoryImpl) SoftDelete(id primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	// Soft delete: Update state to deleted instead of removing document.
	// Zero out sensitive fields but keep key identifiers and audit trails.
	// Fields retained for auditing: id, collection_id, owner_id, created_at, created_by_user_id, modified_at, modified_by_user_id, state, version, tombstone_version, tombstone_expiry
	update := bson.M{
		"$set": bson.M{
			"state":       dom_file.FileStateDeleted,
			"modified_at": time.Now(),
			// Zeroing out sensitive/non-auditing fields:
			"encrypted_metadata":                "",
			"encrypted_file_key":                "", // Assuming zero value is empty string/byte slice
			"encryption_version":                "",
			"encrypted_hash":                    "",
			"encrypted_file_object_key":         "",
			"encrypted_file_size_in_bytes":      int64(0),
			"encrypted_thumbnail_object_key":    "",
			"encrypted_thumbnail_size_in_bytes": int64(0),
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
		// Check if the document existed but wasn't modified (e.g., already deleted)
		// A common approach is to check if MatchedCount is 0. If MatchedCount is 0,
		// the document wasn't found. If MatchedCount is 1 and ModifiedCount is 0,
		// it implies the document state was already the target state or another field prevented modification.
		// For a simple soft delete to a fixed state, ModifiedCount==0 usually means it didn't exist or was already deleted.
		// We'll keep the original logic for now, assuming ModifiedCount == 0 is sufficient for "not found/already deleted".
		impl.Logger.Warn("no file was soft deleted, may not exist or already deleted",
			zap.Any("id", id))
		return errors.New("file not found or already deleted")
	}

	return nil
}

func (impl fileMetadataRepositoryImpl) SoftDeleteMany(ids []primitive.ObjectID) error {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	filter := bson.M{"_id": bson.M{"$in": ids}}

	// Soft delete: Update state to deleted for multiple files.
	// Zero out sensitive fields but keep key identifiers and audit trails.
	// Fields retained for auditing: id, collection_id, owner_id, created_at, created_by_user_id, modified_at, modified_by_user_id, state, version, tombstone_version, tombstone_expiry
	update := bson.M{
		"$set": bson.M{
			"state":       dom_file.FileStateDeleted,
			"modified_at": time.Now(),
			// Zeroing out sensitive/non-auditing fields:
			"encrypted_metadata":                "",
			"encrypted_file_key":                "", // Assuming zero value is empty string/byte slice
			"encryption_version":                "",
			"encrypted_hash":                    "",
			"encrypted_file_object_key":         "",
			"encrypted_file_size_in_bytes":      int64(0),
			"encrypted_thumbnail_object_key":    "",
			"encrypted_thumbnail_size_in_bytes": int64(0),
		},
	}

	result, err := impl.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database failed batch soft deletion error",
			zap.Any("error", err),
			zap.Any("ids", ids))
		return err
	}

	impl.Logger.Debug("Soft deleted files metadata",
		zap.Int64("modifiedCount", result.ModifiedCount), // Log modified count
		zap.Int64("matchedCount", result.MatchedCount),   // Also log matched count for clarity
		zap.Int("requestedCount", len(ids)))

	// Unlike Delete (single), we don't return an error if ModifiedCount is less than requestedCount.
	// This is common for batch operations; some IDs might not exist or were already deleted.
	// The log provides information on how many were actually modified.

	return nil
}

// Add hard delete method for permanent removal
func (impl fileMetadataRepositoryImpl) HardDelete(id primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	_, err := impl.Collection.DeleteOne(ctx, filter)
	if err != nil {
		impl.Logger.Error("database failed hard deletion error",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	return nil
}

func (impl fileMetadataRepositoryImpl) HardDeleteMany(ids []primitive.ObjectID) error {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	filter := bson.M{"_id": bson.M{"$in": ids}}

	result, err := impl.Collection.DeleteMany(ctx, filter)
	if err != nil {
		impl.Logger.Error("database failed batch deletion error",
			zap.Any("error", err),
			zap.Any("ids", ids))
		return err
	}

	impl.Logger.Debug("Deleted files metadata",
		zap.Int64("deletedCount", result.DeletedCount),
		zap.Int("requestedCount", len(ids)))

	return nil
}
