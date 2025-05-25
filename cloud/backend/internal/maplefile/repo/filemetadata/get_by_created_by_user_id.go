// cloud/backend/internal/maplefile/repo/filemetadata/get_by_created_by_user_id.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

// GetByCreatedByUserID gets all files created by a specific user
func (impl fileMetadataRepositoryImpl) GetByCreatedByUserID(createdByUserID primitive.ObjectID) ([]*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{"created_by_user_id": createdByUserID}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get files by created_by_user_id error", zap.Any("error", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var files []*dom_file.File
	if err = cursor.All(ctx, &files); err != nil {
		impl.Logger.Error("database decode files error", zap.Any("error", err))
		return nil, err
	}

	return files, nil
}
