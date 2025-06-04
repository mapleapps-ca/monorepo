// github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/repo/user/impl.go
package user

import (
	"context"

	"go.uber.org/zap"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	dom_user "github.com/mapleapps-ca/monorepo/cloud/backend/internal/maplefile/domain/user"
)

type userStorerImpl struct {
	Logger     *zap.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewRepository(appCfg *config.Configuration, loggerp *zap.Logger, client *mongo.Client) dom_user.Repository {
	loggerp = loggerp.Named("UserRepository")

	// ctx := context.Background()
	uc := client.Database(appCfg.DB.MapleFileName).Collection("users")

	// Create the repository instance first
	repo := &userStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}

	// For debugging purposes only or if you are going to recreate new indexes.
	if err := uc.Indexes().DropAll(context.TODO()); err != nil {
		loggerp.Warn("failed deleting all indexes",
			zap.Any("err", err))
		// Do not crash app, just continue.
	}

	// Note:
	// * 1 for ascending
	// * -1 for descending
	// * "text" for text indexes

	// The following few lines of code will create the index for our app for this
	// collection.
	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{
			{Key: "created_at", Value: -1},
		}},
		{
			Keys:    bson.D{{Key: "email", Value: -1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "created_at", Value: -1},
		}},
		{Keys: bson.D{
			{Key: "name", Value: "text"},
			{Key: "lexical_name", Value: "text"},
			{Key: "email", Value: "text"},
			{Key: "phone", Value: "text"},
			{Key: "country", Value: "text"},
			{Key: "region", Value: "text"},
			{Key: "city", Value: "text"},
			{Key: "postal_code", Value: "text"},
			{Key: "address_line1", Value: "text"},
		}},
	})

	if err != nil {
		// Instead of returning nil, log the error and continue
		// The repository can still function without indexes, just with reduced performance
		loggerp.Error("failed creating indexes, continuing without optimal indexes", zap.Any("err", err))
		loggerp.Warn("repository will function but may have performance implications")
	} else {
		loggerp.Info("successfully created all database indexes")
	}

	// Always return the valid repository instance
	return repo
}

// ListAll retrieves all users from the database
func (impl userStorerImpl) ListAll(ctx context.Context) ([]*dom_user.User, error) {
	impl.Logger.Debug("listing all users")

	cursor, err := impl.Collection.Find(ctx, bson.M{})
	if err != nil {
		impl.Logger.Error("failed to query users", zap.Any("error", err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*dom_user.User
	if err = cursor.All(ctx, &users); err != nil {
		impl.Logger.Error("failed to decode users", zap.Any("error", err))
		return nil, err
	}

	impl.Logger.Debug("successfully retrieved all users", zap.Any("count", len(users)))
	return users, nil
}
