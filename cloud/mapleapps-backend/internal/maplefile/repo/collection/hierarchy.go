// cloud/mapleapps-backend/internal/maplefile/repo/collection/hierarchy.go
package collection

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

func (impl *collectionRepositoryImpl) MoveCollection(
	ctx context.Context,
	collectionID,
	newParentID gocql.UUID,
	updatedAncestors []gocql.UUID,
	updatedPathSegments []string,
) error {
	// Get the collection
	collection, err := impl.Get(ctx, collectionID)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	if collection == nil {
		return fmt.Errorf("collection not found")
	}

	// Update hierarchy information
	collection.ParentID = newParentID
	collection.AncestorIDs = updatedAncestors
	collection.ModifiedAt = time.Now()
	collection.Version++

	// Single update call handles all the complexity
	return impl.Update(ctx, collection)
}
