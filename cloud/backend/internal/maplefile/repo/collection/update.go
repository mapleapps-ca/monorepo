// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/collection/update.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Update(collection *dom_collection.Collection) error {
	ctx := context.Background()
	filter := bson.M{"id": collection.ID}

	// Update the UpdatedAt timestamp
	collection.UpdatedAt = time.Now()

	update := bson.M{
		"$set": collection,
	}

	_, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database update collection error",
			zap.Any("error", err),
			zap.String("id", collection.ID))
		return err
	}

	return nil
}
