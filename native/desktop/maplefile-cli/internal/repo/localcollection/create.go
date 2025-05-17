// monorepo/native/desktop/maplefile-cli/internal/repo/localcollection/create.go
package collection

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"

	"github.com/mapleapps-ca/monorepo/native/desktop/maplefile-cli/internal/domain/localcollection"
)

func (r *localcollectionRepository) Create(ctx context.Context, collection *localcollection.LocalCollection) error {
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
