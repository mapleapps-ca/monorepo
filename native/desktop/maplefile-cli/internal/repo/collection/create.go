// monorepo/native/desktop/maplefile-cli/internal/repo/collection/create.go
package collection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/collection"
)

func (r *collectionRepository) Create(ctx context.Context, collection *collection.Collection) error {
	r.logger.Debug("Creating new local collection",
		zap.String("type", collection.Type),
		zap.Any("parentID", collection.ParentID))

	// Ensure collection has an ID
	if collection.ID.IsZero() {
		collection.ID = primitive.NewObjectID()
	}

	// Set timestamps
	now := time.Now()
	if collection.CreatedAt.IsZero() {
		collection.CreatedAt = now
	}
	collection.ModifiedAt = now

	// Set as local only initially
	collection.IsModifiedLocally = true

	// Save to local storage
	return r.Save(ctx, collection)
}
