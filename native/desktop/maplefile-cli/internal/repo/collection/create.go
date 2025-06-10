// monorepo/native/desktop/maplefile-cli/internal/repo/collection/create.go
package collection

import (
	"context"
	"time"

	"github.com/gocql/gocql"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

func (r *collectionRepository) Create(ctx context.Context, collection *collection.Collection) error {
	// Ensure collection has an ID
	if collection.ID.String() == "" {
		collection.ID = gocql.TimeUUID()
	}

	// Set timestamps
	now := time.Now()
	if collection.CreatedAt.String() == "" {
		collection.CreatedAt = now
	}
	collection.ModifiedAt = now

	// Save to local storage
	return r.Save(ctx, collection)
}
