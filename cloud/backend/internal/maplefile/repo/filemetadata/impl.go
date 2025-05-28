// cloud/backend/internal/maplefile/repo/filemetadata/impl.go
package filemetadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

type fileMetadataRepositoryImpl struct {
	Logger     *zap.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_file.FileMetadataRepository {
	loggerp = loggerp.Named("FileMetadataRepository")

	// Initialize collection in the MapleFile database
	fc := client.Database(appCfg.DB.MapleFileName).Collection("files")

	// Reset indexes for development purposes
	if err := fc.Indexes().DropAll(context.TODO()); err != nil {
		loggerp.Warn("failed deleting all indexes",
			zap.Any("err", err))
	}

	// Create indexes for efficient queries
	_, err := fc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "collection_id", Value: 1},
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "owner_id", Value: 1},
			{Key: "created_at", Value: -1},
		}},
		// Indexes for sync operations
		{Keys: bson.D{
			{Key: "owner_id", Value: 1},
			{Key: "modified_at", Value: 1},
			{Key: "_id", Value: 1},
		}},
		{Keys: bson.D{
			{Key: "collection_id", Value: 1},
			{Key: "modified_at", Value: 1},
			{Key: "_id", Value: 1},
		}},
	})

	if err != nil {
		loggerp.Error("failed creating indexes error", zap.Any("err", err))
		return nil
	}

	return &fileMetadataRepositoryImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: fc,
	}
}
