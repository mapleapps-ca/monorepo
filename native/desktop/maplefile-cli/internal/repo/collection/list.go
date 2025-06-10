// monorepo/native/desktop/maplefile-cli/internal/repo/collection/list.go
package collection

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/common/errors"
	dom_collection "github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

func (r *collectionRepository) List(ctx context.Context, filter dom_collection.CollectionFilter) ([]*dom_collection.Collection, error) {
	collections := make([]*dom_collection.Collection, 0)

	// Iterate through all collections in the database
	err := r.dbClient.Iterate(func(key, value []byte) error {
		// Check if the key has our collection prefix
		keyStr := string(key)
		if !strings.HasPrefix(keyStr, collectionKeyPrefix) {
			// Not a collection key, skip
			return nil
		}

		// Deserialize the collection
		collection, err := dom_collection.NewFromDeserialized(value)
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
			if filter.ParentID.String() == "" {
				if !(collection.ParentID.String() == "") {
					return nil // Skip, not a root collection
				}
			} else if collection.ParentID != *filter.ParentID {
				return nil // Skip, parent doesn't match
			}
		}

		// Filter by collection type if specified
		if filter.CollectionType != "" && collection.CollectionType != filter.CollectionType {
			return nil // Skip, type doesn't match
		}

		// Filter by state if specified
		if filter.State != nil {
			if collection.State != *filter.State {
				return nil // Skip, state doesn't match
			}
		} else if !filter.IncludeDeleted {
			// Default behavior: exclude deleted collections unless explicitly requested
			if collection.State == dom_collection.CollectionStateDeleted {
				return nil // Skip deleted collections
			}
		}

		// Validate collection state (defensive programming)
		if collection.State == "" {
			// Set default state for collections that don't have state set
			collection.State = dom_collection.GetDefaultState()
			r.logger.Warn("Collection found without state, setting default",
				zap.String("collectionID", collection.ID.String()),
				zap.String("defaultState", collection.State))

			// Update the collection in storage with the default state
			if err := r.Save(ctx, collection); err != nil {
				r.logger.Error("Failed to update collection with default state",
					zap.String("collectionID", collection.ID.String()),
					zap.Error(err))
			}
		} else {
			// Validate existing state
			if err := dom_collection.ValidateState(collection.State); err != nil {
				r.logger.Error("Collection has invalid state, skipping",
					zap.String("collectionID", collection.ID.String()),
					zap.String("invalidState", collection.State),
					zap.Error(err))
				return nil // Skip collections with invalid state
			}
		}

		// Filter by sync status if specified
		if filter.SyncStatus != nil {
			var matches bool

			//TODO: IMPL.
			// switch *filter.SyncStatus {
			// case dom_collection.SyncStatusLocalOnly:
			// 	// Consider it local-only if it's modified locally and has never been synced
			// 	matches = collection.IsModifiedLocally && collection.LastSyncedAt.String() == ""
			// case dom_collection.SyncStatusModifiedLocally:
			// 	// Modified locally but has been synced before
			// 	matches = collection.IsModifiedLocally && !collection.LastSyncedAt.String() == ""
			// case dom_collection.SyncStatusSynced:
			// 	// Not modified locally and has been synced
			// 	matches = !collection.IsModifiedLocally && !collection.LastSyncedAt.String() == ""
			// }

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

	r.logger.Debug("Listed collections with filters",
		zap.Int("count", len(collections)),
		zap.Any("filter", filter))

	return collections, nil
}
