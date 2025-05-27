// cloud/backend/internal/maplefile/repo/collection/get.go
package collection

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Get(ctx context.Context, id primitive.ObjectID) (*dom_collection.Collection, error) {
	filter := bson.M{
		"_id":   id,
		"state": dom_collection.CollectionStateActive, // Only return active collections
	}

	var result dom_collection.Collection
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		impl.Logger.Error("database get by collection id error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

// Add method to get collection regardless of state
func (impl collectionRepositoryImpl) GetWithAnyState(ctx context.Context, id primitive.ObjectID) (*dom_collection.Collection, error) {
	filter := bson.M{"_id": id}

	var result dom_collection.Collection
	err := impl.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		impl.Logger.Error("database get by collection id (any state) error", zap.Any("error", err))
		return nil, err
	}
	return &result, nil
}

func (impl collectionRepositoryImpl) GetAllByUserID(ctx context.Context, ownerID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	// Find active collections owned by this user
	filter := bson.M{
		"owner_id": ownerID,
		"state":    dom_collection.CollectionStateActive, // Only return active collections
	}

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

func (impl collectionRepositoryImpl) GetCollectionsSharedWithUser(ctx context.Context, userID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	// Find active collections where user is in members array as recipient
	filter := bson.M{
		"members.recipient_id": userID,
		"owner_id":             bson.M{"$ne": userID},                // Exclude collections owned by the user
		"state":                dom_collection.CollectionStateActive, // Only return active collections
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

func (impl collectionRepositoryImpl) FindByParent(ctx context.Context, parentID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	filter := bson.M{
		"parent_id": parentID,
		"state":     dom_collection.CollectionStateActive, // Only return active collections
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database find by parent error", zap.Any("error", err))
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

func (impl collectionRepositoryImpl) FindRootCollections(ctx context.Context, ownerID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	// Root collections are those without a parent
	filter := bson.M{
		"owner_id":  ownerID,
		"parent_id": bson.M{"$exists": false},
		"state":     dom_collection.CollectionStateActive, // Only return active collections
	}

	cursor, err := impl.Collection.Find(ctx, filter)
	if err != nil {
		impl.Logger.Error("database find root collections error", zap.Any("error", err))
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

func (impl collectionRepositoryImpl) FindDescendants(ctx context.Context, collectionID primitive.ObjectID) ([]*dom_collection.Collection, error) {
	// Find all collections that have this ID in their ancestors array
	filter := bson.M{"ancestor_ids": collectionID}

	// Sort by path length to get a hierarchical order
	opts := options.Find().SetSort(bson.M{"ancestor_ids": 1})

	cursor, err := impl.Collection.Find(ctx, filter, opts)
	if err != nil {
		impl.Logger.Error("database find descendants error", zap.Any("error", err))
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

func (impl collectionRepositoryImpl) GetFullHierarchy(ctx context.Context, rootID primitive.ObjectID) (*dom_collection.Collection, error) {
	// First get the root collection
	rootCollection, err := impl.Get(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if rootCollection == nil {
		return nil, errors.New("root collection not found")
	}

	// Get all descendants
	descendants, err := impl.FindDescendants(ctx, rootID)
	if err != nil {
		return nil, err
	}

	// If no descendants, return just the root
	if len(descendants) == 0 {
		return rootCollection, nil
	}

	// Build a map of parent ID to children collections
	childrenMap := make(map[primitive.ObjectID][]*dom_collection.Collection)
	for _, desc := range descendants {
		parentID := desc.ParentID
		childrenMap[parentID] = append(childrenMap[parentID], desc)
	}

	// Recursive function to build the hierarchy
	var buildHierarchy func(collection *dom_collection.Collection)
	buildHierarchy = func(collection *dom_collection.Collection) {
		children, exists := childrenMap[collection.ID]
		if exists {
			collection.Children = children
			for _, child := range children {
				buildHierarchy(child)
			}
		}
	}

	// Start building from the root
	buildHierarchy(rootCollection)
	return rootCollection, nil
}
