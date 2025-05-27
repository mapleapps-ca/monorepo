// cloud/backend/internal/maplefile/repo/filemetadata/delete.go
package filemetadata

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

func (impl fileMetadataRepositoryImpl) Delete(id primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	// Soft delete: Update state to deleted instead of removing document
	update := bson.M{
		"$set": bson.M{
			"state":       dom_file.FileStateDeleted,
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
		impl.Logger.Warn("no file was soft deleted, may not exist",
			zap.Any("id", id))
		return errors.New("file not found")
	}

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

func (impl fileMetadataRepositoryImpl) DeleteMany(ids []primitive.ObjectID) error {
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	filter := bson.M{"_id": bson.M{"$in": ids}}

	// Soft delete: Update state to deleted for multiple files
	update := bson.M{
		"$set": bson.M{
			"state":       dom_file.FileStateDeleted,
			"modified_at": time.Now(),
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
		zap.Int64("deletedCount", result.ModifiedCount),
		zap.Int("requestedCount", len(ids)))

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
