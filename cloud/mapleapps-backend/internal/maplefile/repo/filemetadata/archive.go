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

func (impl fileMetadataRepositoryImpl) Archive(id primitive.ObjectID) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	// Soft delete: Update state to deleted instead of removing document
	update := bson.M{
		"$set": bson.M{
			"state":       dom_file.FileStateArchived,
			"modified_at": time.Now(),
		},
	}

	result, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database failed archive error",
			zap.Any("error", err),
			zap.Any("id", id))
		return err
	}

	if result.ModifiedCount == 0 {
		impl.Logger.Warn("no file was archived, may not exist",
			zap.Any("id", id))
		return errors.New("file not found")
	}

	return nil
}
