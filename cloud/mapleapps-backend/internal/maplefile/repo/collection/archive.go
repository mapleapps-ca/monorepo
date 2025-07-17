// monorepo/cloud/mapleapps-backend/internal/maplefile/repo/collection/archive.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
	dom_collection "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/internal/maplefile/domain/collection"
)

func (impl *collectionRepositoryImpl) Archive(ctx context.Context, id gocql.UUID) error {
	collection, err := impl.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get collection for archive: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Validate state transition
	if err := dom_collection.IsValidStateTransition(collection.State, dom_collection.CollectionStateArchived); err != nil {
		return fmt.Errorf("invalid state transition: %w", err)
	}

	// Update collection state
	collection.State = dom_collection.CollectionStateArchived
	collection.ModifiedAt = time.Now()
	collection.Version++

	return impl.Update(ctx, collection)
}
