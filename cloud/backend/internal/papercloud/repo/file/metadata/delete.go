// cloud/backend/internal/papercloud/repo/file/metadata/delete.go
package metadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Delete a file's metadata
func (impl fileMetadataRepositoryImpl) Delete(id string) error {
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
