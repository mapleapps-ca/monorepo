// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/save.go
package collection

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

func (r *localcollectionRepository) Save(ctx context.Context, collection *localcollection.LocalCollection) error {
	r.logger.Debug("Saving collection to local storage",
		zap.String("collectionID", collection.ID.Hex()),
		zap.String("collectionType", collection.Type),
		zap.Bool("isModifiedLocally", collection.IsModifiedLocally))

	// Update modified timestamp
	collection.ModifiedAt = time.Now()

	// Serialize the collection
	collBytes, err := collection.Serialize()
	if err != nil {
		r.logger.Error("Failed to serialize collection", zap.Error(err))
		return errors.NewAppError("failed to serialize collection", err)
	}

	// Generate key for this collection using the ID
	key := r.generateKey(collection.ID.Hex())

	// Save to database
	if err := r.dbClient.Set(key, collBytes); err != nil {
		r.logger.Error("Failed to save collection to local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to save collection to local storage", err)
	}

	r.logger.Info("Collection saved successfully to local storage",
		zap.String("collectionID", collection.ID.Hex()),
		zap.Bool("isModifiedLocally", collection.IsModifiedLocally))
	return nil
}
