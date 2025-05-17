// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/list.go
package collection

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

func (r *localcollectionRepository) List(ctx context.Context, filter localcollection.LocalCollectionFilter) ([]*localcollection.LocalCollection, error) {
	r.logger.Debug("Listing collections from local storage",
		zap.Any("filter", filter))

	collections := make([]*localcollection.LocalCollection, 0)

	// Iterate through all collections in the database
	err := r.dbClient.Iterate(func(key, value []byte) error {
		// Check if the key has our collection prefix
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, collectionKeyPrefix) {
			// Not a collection key, skip
			return nil
		}

		// Deserialize the collection
		collection, err := localcollection.NewFromDeserialized(value)
		if err != nil {
			r.logger.Error("Failed to deserialize collection while listing",
				zap.String("key", keyStr),
				zap.Error(err))
			return nil // Continue iteration despite error
		}

		// Apply filters

		// Filter by parentID if specified
		if filter.ParentID != nil {
			// For root collections (parentID is nil), we check for zero ObjectID
			if filter.ParentID.IsZero() {
				if !collection.ParentID.IsZero() {
					return nil // Skip, not a root collection
				}
			} else if collection.ParentID != *filter.ParentID {
				return nil // Skip, parent doesn't match
			}
		}

		// Filter by type if specified
		if filter.Type != "" && collection.Type != filter.Type {
			return nil // Skip, type doesn't match
		}

		// Filter by sync status if specified
		if filter.SyncStatus != nil {
			var matches bool

			switch *filter.SyncStatus {
			case localcollection.SyncStatusLocalOnly:
				// Consider it local-only if it's modified locally and has never been synced
				matches = collection.IsModifiedLocally && collection.LastSyncedAt.IsZero()
			case localcollection.SyncStatusModifiedLocally:
				// Modified locally but has been synced before
				matches = collection.IsModifiedLocally && !collection.LastSyncedAt.IsZero()
			case localcollection.SyncStatusSynced:
				// Not modified locally and has been synced
				matches = !collection.IsModifiedLocally && !collection.LastSyncedAt.IsZero()
			}

			if !matches {
				return nil // Skip, sync status doesn't match
			}
		}

		// Add to results
		collections = append(collections, collection)
		return nil
	})

	if err != nil {
		r.logger.Error("Error iterating through collections in local storage", zap.Error(err))
		return nil, errors.NewAppError("failed to list collections from local storage", err)
	}

	r.logger.Info("Successfully listed collections from local storage",
		zap.Int("count", len(collections)))
	return collections, nil
}
