// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/get.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

func (r *localcollectionRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*localcollection.LocalCollection, error) {
	r.logger.Debug("Retrieving collection from local storage", zap.String("collectionID", id.Hex()))

	// Generate key for this collection
	key := r.generateKey(id.Hex())

	// Get from database
	collBytes, err := r.dbClient.Get(key)
	if err != nil {
		r.logger.Error("Failed to retrieve collection from local storage",
			zap.String("key", key),
			zap.Error(err))
		return nil, errors.NewAppError("failed to retrieve collection from local storage", err)
	}

	// Check if collection was found
	if collBytes == nil {
		r.logger.Warn("Collection not found in local storage", zap.String("collectionID", id.Hex()))
		return nil, nil
	}

	// Deserialize the collection
	collection, err := localcollection.NewFromDeserialized(collBytes)
	if err != nil {
		r.logger.Error("Failed to deserialize collection", zap.Error(err))
		return nil, errors.NewAppError("failed to deserialize collection", err)
	}

	r.logger.Debug("Successfully retrieved collection from local storage",
		zap.String("collectionID", id.Hex()),
		zap.String("collectionType", collection.Type))
	return collection, nil
}
