// github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/repo/collection/get.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/collection"
)

func (impl collectionRepositoryImpl) Get(id string) (*dom_collection.Collection, error) {
	ctx := context.Background()
	filter := bson.M{"id": id}

	var result dom_collection.Collection
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No matching document found
			return nil, nil
		}
		impl.Logger.Error("database get by collection id error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl collectionRepositoryImpl) GetAllByUserID(userID string) ([]*dom_collection.Collection, error) {
	ctx := context.Background()
	// Find collections owned by this user
	filter := bson.M{"owner_id": userID}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get collections by user id error", zap.Any("error", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []*dom_collection.Collection
	if err = cursor.All(ctx, &collections); err != nil {
		impl.Logger.Error("database decode collections error", zap.Any("error", err))
		return nil, err
	}

	return collections, nil
}

func (impl collectionRepositoryImpl) GetCollectionsSharedWithUser(userID string) ([]*dom_collection.Collection, error) {
	ctx := context.Background()
	// Find collections where user is in members array as recipient
	filter := bson.M{
		"members.recipient_id": userID,
		// Exclude collections owned by the user to avoid duplicates
		"owner_id": bson.M{"$ne": userID},
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database get shared collections error", zap.Any("error", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var collections []*dom_collection.Collection
	if err = cursor.All(ctx, &collections); err != nil {
		impl.Logger.Error("database decode shared collections error", zap.Any("error", err))
		return nil, err
	}

	return collections, nil
}
