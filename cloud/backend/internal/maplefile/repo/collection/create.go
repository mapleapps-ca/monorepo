// cloud/backend/internal/maplefile/repo/collection/create.go
package collection

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

func (impl collectionRepositoryImpl) Create(ctx context.Context, collection *dom_collection.Collection) error {
	// Validate owner ID
	if collection.OwnerID.IsZero() {
		impl.Logger.Error("owner ID is required but not provided")
		return errors.New("owner ID is required")
	}

	// Generate ObjectID if not provided
	if collection.ID.IsZero() {
		collection.ID = primitive.NewObjectID()
	}

	// Set creation time if not set
	if collection.CreatedAt.IsZero() {
		collection.CreatedAt = time.Now()
	}

	// Set update time to match creation time
	collection.ModifiedAt = collection.CreatedAt

	// Initialize empty members array if not set
	if collection.Members == nil {
		collection.Members = []dom_collection.CollectionMembership{}
	}

	// Initialize empty children array if not set
	if collection.Children == nil {
		collection.Children = []*dom_collection.Collection{}
	}

	// If this is a child collection (has parent), set up hierarchical data
	if !collection.ParentID.IsZero() {
		// Find the parent to get its ancestors
		parentFilter := bson.M{"_id": collection.ParentID}
		var parent dom_collection.Collection
		err := impl.Collection.FindOne(ctx, parentFilter).Decode(&parent)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return errors.New("parent collection not found")
			}
			impl.Logger.Error("failed to retrieve parent collection", zap.Any("error", err))
			return err
		}

		// Build ancestors array (parent + parent's ancestors)
		collection.AncestorIDs = append([]primitive.ObjectID{parent.ID}, parent.AncestorIDs...)
	}

	// Insert collection document
	_, err := impl.Collection.InsertOne(ctx, collection)
	if err != nil {
		impl.Logger.Error("database failed create collection error",
			zap.Any("error", err),
			zap.Any("id", collection.ID))
		return err
	}

	return nil
}
