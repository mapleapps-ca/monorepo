// cloud/mapleapps-backend/internal/maplefile/repo/collection/restore.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) Restore(ctx context.Context, id gocql.UUID) error {
	collection, err := impl.GetWithAnyState(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collection for restore: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Validate state transition
	if err := dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateActive); err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	// Update collection state
	collection.State = dom_collection.CollectionStateActive
	collection.ModifiedAt = time.Now()
	collection.Version++
	collection.TombstoneVersion = 0
	collection.TombstoneExpiry = time.Time{}

	return impl.Update(ctx, collection)
}
