// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/delete.go
package collection

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
)

func (r *localcollectionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	r.logger.Debug("Deleting collection from local storage", zap.String("collectionID", id.Hex()))

	// Generate key for this collection
	key := r.generateKey(id.Hex())

	// Delete from database
	if err := r.dbClient.Delete(key); err != nil {
		r.logger.Error("Failed to delete collection from local storage",
			zap.String("key", key),
			zap.Error(err))
		return errors.NewAppError("failed to delete collection from local storage", err)
	}

	r.logger.Info("Collection deleted successfully from local storage",
		zap.String("collectionID", id.Hex()))
	return nil
}
