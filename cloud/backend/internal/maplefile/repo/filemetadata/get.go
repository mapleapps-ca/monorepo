// cloud/backend/internal/maplefile/repo/filemetadata/get.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

// Get file by ID
func (impl fileMetadataRepositoryImpl) Get(id primitive.ObjectID) (*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{
		"_id":   id,
		"state": dom_file.FileStateActive, // Only return active files
	}

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

func (impl fileMetadataRepositoryImpl) GetWithAnyState(id primitive.ObjectID) (*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	var result dom_file.File
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		impl.Logger.Error("database get by file id (any state) error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

// GetByIDs gets files by their IDs
func (impl fileMetadataRepositoryImpl) GetByIDs(ids []primitive.ObjectID) ([]*dom_file.File, error) {
	if len(ids) == 0 {
		return []*dom_file.File{}, nil
	}

	ctx := context.Background()
	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get files by IDs error", zap.Any("error", err))
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

// GetByEncryptedFileID gets a file by its client-generated EncryptedFileID
func (impl fileMetadataRepositoryImpl) GetByEncryptedFileID(encryptedFileID string) (*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{"encrypted_file_id": encryptedFileID}

	var result dom_file.File
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		impl.Logger.Error("database get by encrypted_file_id error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

// GetByCollection gets all files in a collection
func (impl fileMetadataRepositoryImpl) GetByCollection(collectionID primitive.ObjectID) ([]*dom_file.File, error) {
	ctx := context.Background()
	filter := bson.M{
		"collection_id": collectionID,
		"state":         bson.M{"$in": []string{dom_file.FileStateActive, dom_file.FileStatePending}}, // Return active and pending files
	}

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
