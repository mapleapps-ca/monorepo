// cloud/backend/internal/papercloud/repo/file/metadata/get.go
package metadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
)

// Get file by ID
func (impl fileMetadataRepositoryImpl) Get(id string) (*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{"id": id}

	var result dom_file.File
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		impl.Logger.Error("database get by file id error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

// GetByFileID gets a file by its client-generated FileID
func (impl fileMetadataRepositoryImpl) GetByFileID(fileID string) (*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{"file_id": fileID}

	var result dom_file.File
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		impl.Logger.Error("database get by file_id error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

// GetByCollection gets all files in a collection
func (impl fileMetadataRepositoryImpl) GetByCollection(collectionID string) ([]*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{"collection_id": collectionID}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get files by collection id error", zap.Any("error", err))
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
