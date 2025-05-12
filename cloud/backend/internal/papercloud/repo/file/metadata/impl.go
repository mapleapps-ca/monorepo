// cloud/backend/internal/papercloud/repo/file/metadata/impl.go
package metadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/papercloud/domain/file"
)

type FileMetadataRepository interface {
	Create(file *dom_file.File) error
	Get(id string) (*dom_file.File, error)
	GetByFileID(fileID string) (*dom_file.File, error)
	GetByCollection(collectionID string) ([]*dom_file.File, error)
	Update(file *dom_file.File) error
	Delete(id string) error
	CheckIfExistsByID(id string) (bool, error)
	CheckIfUserHasAccess(fileID string, userID string) (bool, error)
}

type fileMetadataRepositoryImpl struct {
	Logger     *zap.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) FileMetadataRepository {
	// Initialize collection in the PaperCloud database
	fc := client.Database(appCfg.DB.PaperCloudName).Collection("files")

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
			{Key: "id", Value: 1},
		}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{
			{Key: "file_id", Value: 1},
		}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{
			{Key: "collection_id", Value: 1},
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "owner_id", Value: 1},
			{Key: "created_at", Value: -1},
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
