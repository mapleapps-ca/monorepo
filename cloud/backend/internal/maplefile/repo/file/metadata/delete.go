// cloud/backend/internal/maplefile/repo/file/metadata/delete.go
package metadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Delete a file's metadata
func (impl fileMetadataRepositoryImpl) Delete(id primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	_, err := impl.Collection.DeleteOne(ctx, filter)
	if err != nil {
		impl.Logger.Error("database failed deletion error",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	return nil
}

// DeleteMany deletes multiple files' metadata
func (impl fileMetadataRepositoryImpl) DeleteMany(ids []primitive.ObjectID) error {
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
