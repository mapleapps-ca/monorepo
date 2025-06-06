// cloud/backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Archive(ctx context.Context, id gocql.UUID) error {
	filter := bson.M{"_id": id}

	// Soft delete: Update state to deleted instead of removing document
	update := bson.M{
		"$set": bson.M{
			"state":       dom_collection.CollectionStateArchived,
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
		impl.Logger.Warn("no collection was archived, may not exist",
			zap.Any("id", id))
		return errors.New("collection not found")
	}

	// Also soft delete all children collections
	childrenFilter := bson.M{"ancestor_ids": id}
	childrenUpdate := bson.M{
		"$set": bson.M{
			"state":       dom_collection.CollectionStateArchived,
			"modified_at": time.Now(),
		},
	}

	_, err = impl.Collection.UpdateMany(ctx, childrenFilter, childrenUpdate)
	if err != nil {
		impl.Logger.Error("failed to archive child collections",
			zap.Any("error", err),
			zap.Any("parent_id", id))
		return err
	}

	return nil
}
