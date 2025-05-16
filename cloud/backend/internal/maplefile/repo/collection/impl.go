// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/collection/impl.go
package collection

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/collection"
)

type collectionRepositoryImpl struct {
	Logger     *zap.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_collection.CollectionRepository {
	// Initialize collection in the MapleFile database
	cc := client.Database(appCfg.DB.MapleFileName).Collection("collections")

	// Reset indexes for development purposes
	if err := cc.Indexes().DropAll(context.TODO()); err != nil {
		loggerp.Warn("failed deleting all indexes",
			zap.Any("err", err))
		// Continue without returning error
	}

	// Create indexes for efficient queries
	_, err := cc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "id", Value: 1},
		}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{
			{Key: "owner_id", Value: 1},
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "members.recipient_id", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "members.collection_id", Value: 1},
			{Key: "members.recipient_id", Value: 1},
		}},
	})

	if err != nil {
		loggerp.Error("failed creating indexes error", zap.Any("err", err))
		return nil
	}

	return &collectionRepositoryImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: cc,
	}
}
