// cloud/backend/internal/vault/repo/encryptedfile/mongodb_repo.go
package encryptedfile

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/cloud/backend/config"
	domain "github.com/mapleapps-ca/monorepo/cloud/backend/internal/vault/domain/encryptedfile"
)

// encryptedFileRepository implements the domain.Repository interface
type encryptedFileRepository struct {
	logger     *zap.Logger
	collection *mongo.Collection
	database   *mongo.Database
}

// NewRepository creates a new repository for encrypted files
func NewRepository(
	cfg *config.Configuration,
	logger *zap.Logger,
	dbClient *mongo.Client,
) domain.Repository {
	// Initialize the MongoDB database
	database := dbClient.Database(cfg.DB.VaultName)

	// Initialize the MongoDB collection for file metadata
	collection := database.Collection("encrypted_files")

	// Create indexes for efficient queries
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "file_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexModels)
	if err != nil {
		logger.Error("Failed to create indexes for encrypted files collection", zap.Error(err))
	}

	return &encryptedFileRepository{
		logger:     logger.With(zap.String("component", "encrypted-file-repository")),
		collection: collection,
		database:   database,
	}
}
