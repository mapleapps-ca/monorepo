// cloud/backend/internal/maplefile/repo/collection/update.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Update(ctx context.Context, collection *dom_collection.Collection) error {
	filter := bson.M{"_id": collection.ID}

	// Update the ModifiedAt timestamp
	collection.ModifiedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"encrypted_name":           collection.EncryptedName,
			"collection_type":          collection.CollectionType,
			"modified_at":              collection.ModifiedAt,
			"encrypted_collection_key": collection.EncryptedCollectionKey,
			"encrypted_path_segments":  collection.EncryptedPathSegments,
		},
	}

	_, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database update collection error",
			zap.Any("error", err),
			zap.Any("id", collection.ID))
		return err
	}

	return nil
}
