// cloud/backend/internal/maplefile/repo/file/metadata/impl.go
package metadata

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_file "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/file"
)

type FileMetadataRepository interface {
	Create(file *dom_file.File) error
	CreateMany(files []*dom_file.File) error
	Get(id primitive.ObjectID) (*dom_file.File, error)
	GetByIDs(ids []primitive.ObjectID) ([]*dom_file.File, error)
	GetByEncryptedFileID(encryptedFileID string) (*dom_file.File, error)
	GetByCollection(collectionID primitive.ObjectID) ([]*dom_file.File, error)
	Update(file *dom_file.File) error
	Delete(id primitive.ObjectID) error
	DeleteMany(ids []primitive.ObjectID) error
	CheckIfExistsByID(id primitive.ObjectID) (bool, error)
	CheckIfUserHasAccess(fileID primitive.ObjectID, userID primitive.ObjectID) (bool, error)
}

type fileMetadataRepositoryImpl struct {
	Logger     *zap.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) FileMetadataRepository {
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
			{Key: "_id", Value: 1},
		}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{
			{Key: "encrypted_file_id", Value: 1},
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
