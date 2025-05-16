// cloud/backend/internal/maplefile/repo/file/metadata/update.go
package metadata

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

// Update a file's metadata
func (impl fileMetadataRepositoryImpl) Update(file *dom_file.File) error {
	ctx := context.Background()
	filter := bson.M{"_id": file.ID}

	// Update modification time
	file.ModifiedAt = time.Now()

	update := bson.M{
		"$set": file,
	}

	_, err := impl.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		impl.Logger.Error("database update file error",
			zap.Any("error", err),
			zap.Any("id", file.ID))
		return err
	}

	return nil
}
