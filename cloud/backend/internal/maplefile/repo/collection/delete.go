// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/collection/delete.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func (impl collectionRepositoryImpl) Delete(id string) error {
	ctx := context.Background()
	filter := bson.M{"id": id}

	_, err := impl.Collection.DeleteOne(ctx, filter)
	if err != nil {
		impl.Logger.Error("database failed deletion error",
			zap.Any("error", err),
			zap.String("id", id))
		return err
	}

	return nil
}
