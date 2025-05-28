// cloud/backend/internal/maplefile/repo/collection/impl.go
package collection

import (
	"context"
	"log"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

type collectionRepositoryImpl struct {
	Logger     *zap.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_collection.CollectionRepository {
	loggerp = loggerp.Named("CollectionRepository")

	// Initialize collection in the MapleFile database
	cc := client.Database(appCfg.DB.MapleFileName).Collection("collections")

	// Reset indexes for development purposes
	// Use context.Background() instead of context.TODO()
	if err := cc.Indexes().DropAll(context.Background()); err != nil {
		loggerp.Warn("failed deleting all indexes",
			zap.Any("err", err))
		// Continue without returning error
	}

	// Create indexes for efficient queries
	// Use context.Background() instead of context.TODO()
	_, err := cc.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "owner_id", Value: 1},
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "parent_id", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "ancestor_ids", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "members.recipient_id", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "members.collection_id", Value: 1},
			{Key: "members.recipient_id", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "members.is_inherited", Value: 1},
		}},
		// Indexes for sync operations
		{Keys: bson.D{
			{Key: "owner_id", Value: 1},
			{Key: "modified_at", Value: 1},
			{Key: "_id", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "members.recipient_id", Value: 1},
			{Key: "modified_at", Value: 1},
			{Key: "_id", Value: 1},
		}},
	})

	if err != nil {
		log.Fatalf("failed creating indexes error: %v\n", err)
	}

	return &collectionRepositoryImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: cc,
	}
}
