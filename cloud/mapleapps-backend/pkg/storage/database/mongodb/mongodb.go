package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.uber.org/zap"

	c "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config"
)

func NewProvider(appCfg *c.Configuration, logger *zap.Logger) *mongo.Client {
	logger = logger.Named("MongoDB Provider")
	logger.Debug("storage initializing...")

	// DEVELOPERS NOTE:
	// If you uncommented the ABOVE code then comment out the BOTTOM code.
	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	client, err := mongo.Connect(options.Client().ApplyURI(appCfg.DB.URI))
	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

	if err != nil {
		log.Fatalf("backend.pkg.storage.database.mongodb.NewProvider: Error: %v\n", err)
	}

	// The MongoDB client provides a Ping() method to tell you if a MongoDB database has been found and connected.
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	logger.Debug("storage initialized successfully")
	return client
}
